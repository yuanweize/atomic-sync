package engine

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yuanweize/atomic-sync/internal/model"
	"github.com/yuanweize/atomic-sync/internal/store"
)

var (
	ErrJobActive = errors.New("job is already running")
	ErrJobPaused = errors.New("job is paused")
)

type Event struct {
	Type     string          `json:"type"`
	Run      model.Run       `json:"run,omitempty"`
	Analysis *model.Analysis `json:"analysis,omitempty"`
}

type executor func(context.Context, ...string) ([]byte, error)

type Runner struct {
	store       *store.Store
	rclone      string
	rcloneArgs  []string
	sem         chan struct{}
	analysisSem chan struct{}
	execute     executor
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.Mutex
	active      map[string]bool
	subscribers map[chan Event]struct{}
}

type listed struct {
	Path    string     `json:"Path"`
	ModTime rcloneTime `json:"ModTime"`
	IsDir   bool       `json:"IsDir"`
	Size    int64      `json:"Size"`
}

// rcloneTime accepts the empty string emitted by `rclone lsjson --no-modtime`.
// Discovery still treats an empty value as unknown and fails closed whenever a
// stable-window decision depends on it.
type rcloneTime struct {
	time.Time
}

func (value *rcloneTime) UnmarshalJSON(data []byte) error {
	if string(data) == `""` || string(data) == "null" {
		value.Time = time.Time{}
		return nil
	}
	return value.Time.UnmarshalJSON(data)
}

type inventoryUnit struct {
	files       map[string]int64
	directories map[string]struct{}
}

type inventory struct {
	units map[string]*inventoryUnit
	kinds map[string]bool // true for directories, false for files
}

type unitExecution struct {
	job             model.Job
	unit            string
	destination     model.Destination
	destinationName string
	run             model.Run
	source          string
	stage           string
	final           string
}

func New(s *store.Store, rclone string, concurrency int) *Runner {
	return NewWithLimits(s, rclone, concurrency, 2, 2, 2)
}

func NewWithLimits(s *store.Store, rclone string, concurrency, transfers, checkers, tpsLimit int) *Runner {
	if concurrency < 1 {
		concurrency = 2
	}
	transfers = min(max(1, transfers), 64)
	checkers = min(max(1, checkers), 64)
	tpsLimit = min(max(1, tpsLimit), 64)
	ctx, cancel := context.WithCancel(context.Background())
	r := &Runner{
		store: s, rclone: rclone,
		rcloneArgs: []string{
			"--transfers", strconv.Itoa(transfers), "--checkers", strconv.Itoa(checkers),
			"--tpslimit", strconv.Itoa(tpsLimit), "--tpslimit-burst", "1",
		},
		sem: make(chan struct{}, concurrency), analysisSem: make(chan struct{}, 1),
		ctx: ctx, cancel: cancel, active: map[string]bool{}, subscribers: map[chan Event]struct{}{},
	}
	r.execute = r.execRclone
	return r
}

func (r *Runner) Subscribe() (chan Event, func()) {
	ch := make(chan Event, 16)
	r.mu.Lock()
	r.subscribers[ch] = struct{}{}
	r.mu.Unlock()
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.subscribers, ch)
			close(ch)
			r.mu.Unlock()
		})
	}
}

func (r *Runner) emit(e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for ch := range r.subscribers {
		select {
		case ch <- e:
		default:
		}
	}
}

func (r *Runner) IsActive(jobID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active[jobID]
}

func (r *Runner) ActiveCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.active)
}

// Start schedules a job against the Runner lifecycle. Shutdown cancels and
// waits for every job started through this method.
func (r *Runner) Start(j model.Job) error {
	if err := j.Validate(); err != nil {
		return err
	}
	if j.Paused {
		return ErrJobPaused
	}
	select {
	case <-r.ctx.Done():
		return context.Canceled
	default:
	}
	if err := r.begin(j.ID); err != nil {
		return err
	}
	invalidateCtx, invalidateCancel := context.WithTimeout(r.ctx, 5*time.Second)
	if err := r.invalidateAnalysis(invalidateCtx, j); err != nil {
		invalidateCancel()
		r.end(j.ID)
		return err
	}
	invalidateCancel()
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.end(j.ID)
		if err := r.run(r.ctx, j); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("job run failed", "job", j.ID, "error", err)
		}
	}()
	return nil
}

// StartAnalysis compares the physical source and destination branches by
// relative path and size. It never writes to either branch.
func (r *Runner) StartAnalysis(j model.Job) error {
	if err := j.Validate(); err != nil {
		return err
	}
	select {
	case <-r.ctx.Done():
		return context.Canceled
	default:
	}
	if err := r.begin(j.ID); err != nil {
		return err
	}
	analysis := model.Analysis{
		JobID: j.ID, State: "running", Summary: newAnalysisSummary(), StartedAt: time.Now().UTC(),
	}
	if err := r.persistAnalysis(analysis); err != nil {
		r.end(j.ID)
		return err
	}
	r.wg.Add(1)
	go r.processAnalysis(j, analysis)
	return nil
}

func (r *Runner) processAnalysis(job model.Job, analysis model.Analysis) {
	defer r.wg.Done()
	defer r.end(job.ID)
	units, err := r.serializedAnalysis(job)
	finishAnalysis(&analysis, units, err)
	if saveErr := r.persistAnalysis(analysis); saveErr != nil {
		slog.Error("save archive analysis", "job", job.ID, "error", saveErr)
	}
	eventAnalysis := analysis
	eventAnalysis.Units = nil
	r.emit(Event{Type: "analysis", Analysis: &eventAnalysis})
}

func (r *Runner) serializedAnalysis(job model.Job) ([]model.UnitAnalysis, error) {
	select {
	case r.analysisSem <- struct{}{}:
		defer func() { <-r.analysisSem }()
		return r.analyze(r.ctx, job)
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	}
}

func finishAnalysis(analysis *model.Analysis, units []model.UnitAnalysis, err error) {
	finished := time.Now().UTC()
	analysis.FinishedAt = &finished
	if err != nil {
		analysis.State = "failed"
		analysis.Message = err.Error()
		return
	}
	analysis.State = "completed"
	analysis.Units = units
	for _, unit := range units {
		analysis.Summary[unit.Status]++
	}
}

func (r *Runner) persistAnalysis(analysis model.Analysis) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.store.SaveAnalysis(ctx, analysis)
}

// Run executes synchronously and is primarily useful to callers that own the
// context, including tests and future schedulers.
func (r *Runner) Run(ctx context.Context, j model.Job) error {
	if err := j.Validate(); err != nil {
		return err
	}
	if j.Paused {
		return ErrJobPaused
	}
	if err := r.begin(j.ID); err != nil {
		return err
	}
	defer r.end(j.ID)
	if err := r.invalidateAnalysis(ctx, j); err != nil {
		return err
	}
	return r.run(ctx, j)
}

func (r *Runner) invalidateAnalysis(ctx context.Context, job model.Job) error {
	if job.DryRun {
		return nil
	}
	return r.store.DeleteAnalysis(ctx, job.ID)
}

func (r *Runner) begin(jobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active[jobID] {
		return ErrJobActive
	}
	r.active[jobID] = true
	return nil
}

func (r *Runner) end(jobID string) {
	r.mu.Lock()
	delete(r.active, jobID)
	r.mu.Unlock()
}

func (r *Runner) Shutdown(ctx context.Context) error {
	r.cancel()
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) run(ctx context.Context, j model.Job) error {
	units, err := r.discover(ctx, j)
	if err != nil {
		return err
	}
	if len(units) == 0 {
		return nil
	}
	return r.runUnits(ctx, j, units)
}

func (r *Runner) runUnits(ctx context.Context, job model.Job, units []string) error {
	if err := validateIndependentUnits(units); err != nil {
		return err
	}
	workers := min(max(1, job.Concurrency), len(units))
	unitCh := make(chan string)
	var workersWG sync.WaitGroup
	var errMu sync.Mutex
	var errs []error
	workersWG.Add(workers)
	for range workers {
		go r.unitWorker(ctx, job, unitCh, &workersWG, &errMu, &errs)
	}
	sendUnits(ctx, unitCh, units)
	close(unitCh)
	workersWG.Wait()
	if ctx.Err() != nil {
		errs = append(errs, ctx.Err())
	}
	return errors.Join(errs...)
}

func (r *Runner) unitWorker(ctx context.Context, job model.Job, units <-chan string, workers *sync.WaitGroup, errMu *sync.Mutex, errs *[]error) {
	defer workers.Done()
	for unit := range units {
		err := r.runUnit(ctx, job, unit)
		if err == nil {
			continue
		}
		slog.Error("unit failed", "job", job.Name, "unit", unit, "error", err)
		errMu.Lock()
		*errs = append(*errs, fmt.Errorf("%s: %w", unit, err))
		errMu.Unlock()
	}
}

func sendUnits(ctx context.Context, target chan<- string, units []string) {
	for _, unit := range units {
		select {
		case target <- unit:
		case <-ctx.Done():
			return
		}
	}
}

func (r *Runner) discover(ctx context.Context, j model.Job) ([]string, error) {
	out, err := r.command(ctx, "lsjson", j.Source, "--recursive", "--files-only")
	if err != nil {
		return nil, err
	}
	var files []listed
	if err = json.Unmarshal(out, &files); err != nil {
		return nil, fmt.Errorf("decode rclone listing: %w", err)
	}
	latest := map[string]time.Time{}
	for _, file := range files {
		unit := UnitFor(file.Path, j.Grouping, j.Depth)
		if unit == "" {
			return nil, fmt.Errorf("file %q is not inside a valid %s directory unit", file.Path, j.Grouping)
		}
		if file.ModTime.IsZero() {
			if j.SettleSeconds > 0 {
				return nil, fmt.Errorf("file %q has no modification time; stable-window eligibility cannot be proven", file.Path)
			}
			if _, exists := latest[unit]; !exists {
				latest[unit] = time.Time{}
			}
			continue
		}
		if file.ModTime.After(latest[unit]) {
			latest[unit] = file.ModTime.Time
		}
	}
	cutoff := time.Now().Add(-time.Duration(j.SettleSeconds) * time.Second)
	units := make([]string, 0, len(latest))
	for unit, modified := range latest {
		if j.SettleSeconds <= 0 || modified.Before(cutoff) {
			units = append(units, unit)
		}
	}
	sort.Strings(units)
	return units, nil
}

func validateIndependentUnits(units []string) error {
	seen := make(map[string]struct{}, len(units))
	for _, unit := range units {
		if !safeRelative(unit) {
			return fmt.Errorf("unsafe unit path %q", unit)
		}
		if internalDestinationPath(unit) {
			return fmt.Errorf("unit path %q uses the reserved .atomic-sync-staging namespace", unit)
		}
		seen[unit] = struct{}{}
	}
	for _, unit := range units {
		for parent := path.Dir(unit); parent != "."; parent = path.Dir(parent) {
			if _, exists := seen[parent]; exists {
				return fmt.Errorf("overlapping migration units %q and %q", parent, unit)
			}
		}
	}
	return nil
}

func (r *Runner) analyze(ctx context.Context, job model.Job) ([]model.UnitAnalysis, error) {
	if len(job.Destinations) == 0 {
		return nil, errors.New("archive analysis requires at least one destination")
	}
	source, err := r.inventory(ctx, job.Source, job)
	if err != nil {
		return nil, fmt.Errorf("scan source inventory: %w", err)
	}
	destinations := make(map[string]inventory, len(job.Destinations))
	for _, destination := range job.Destinations {
		items, inventoryErr := r.destinationInventory(ctx, destination.Path, job)
		if inventoryErr != nil {
			return nil, fmt.Errorf("scan destination %s: %w", destination.Name, inventoryErr)
		}
		destinations[destination.Name] = items
	}
	unitNames := logicalInventoryUnits(source, destinations)
	analyses := make([]model.UnitAnalysis, 0, len(unitNames))
	for _, unitName := range unitNames {
		sourceUnit := source.units[unitName]
		destination, destinationErr := r.analysisDestination(ctx, job, unitName, sourceUnit, destinations)
		if destinationErr != nil {
			return nil, destinationErr
		}
		destinationInventory := destinations[destination.Name]
		destinationUnit := destinationInventory.units[unitName]
		analysis := compareInventory(unitName, destination.Name, sourceUnit, destinationUnit, source.kinds, destinationInventory.kinds)
		compareUnexpectedDestinations(&analysis, unitName, destination.Name, job.Destinations, destinations)
		finalizeInventoryAnalysis(&analysis)
		analyses = append(analyses, analysis)
	}
	return analyses, nil
}

func logicalInventoryUnits(source inventory, destinations map[string]inventory) []string {
	names := make(map[string]struct{}, len(source.units))
	for unitName := range source.units {
		names[unitName] = struct{}{}
	}
	for _, destination := range destinations {
		for unitName := range destination.units {
			names[unitName] = struct{}{}
		}
	}
	units := make([]string, 0, len(names))
	for unitName := range names {
		units = append(units, unitName)
	}
	sort.Strings(units)
	result := make([]string, 0, len(units))
	for _, unitName := range units {
		isEmpty := inventoryUnitIsEmpty(source.units[unitName])
		for _, destination := range destinations {
			isEmpty = isEmpty && inventoryUnitIsEmpty(destination.units[unitName])
		}
		descendant := sort.SearchStrings(units, unitName+"/")
		hasDescendant := descendant < len(units) && strings.HasPrefix(units[descendant], unitName+"/")
		if isEmpty && hasDescendant {
			continue
		}
		result = append(result, unitName)
	}
	return result
}

func inventoryUnitIsEmpty(unit *inventoryUnit) bool {
	return unit == nil || len(unit.files) == 0
}

func (r *Runner) inventory(ctx context.Context, root string, job model.Job) (inventory, error) {
	return r.scanInventory(ctx, root, job, false)
}

func (r *Runner) destinationInventory(ctx context.Context, root string, job model.Job) (inventory, error) {
	return r.scanInventory(ctx, root, job, true)
}

func (r *Runner) scanInventory(ctx context.Context, root string, job model.Job, excludeInternal bool) (inventory, error) {
	args := []string{"lsjson", root, "--recursive", "--no-modtime", "--no-mimetype"}
	if excludeInternal {
		args = append(args, "--exclude", "/.atomic-sync-staging/**")
	}
	out, err := r.command(ctx, args...)
	if err != nil {
		return inventory{}, err
	}
	var entries []listed
	if err = json.Unmarshal(out, &entries); err != nil {
		return inventory{}, fmt.Errorf("decode inventory listing: %w", err)
	}
	result := inventory{units: map[string]*inventoryUnit{}, kinds: map[string]bool{}}
	for _, entry := range entries {
		if err = addInventoryEntry(&result, entry, job, excludeInternal); err != nil {
			return inventory{}, err
		}
	}
	addUnmappedEmptyUnits(&result, job)
	return result, nil
}

func addInventoryEntry(result *inventory, entry listed, job model.Job, excludeInternal bool) error {
	entry.Path = strings.ReplaceAll(entry.Path, `\\`, "/")
	if !safeRelative(entry.Path) {
		return fmt.Errorf("listing returned unsafe path %q", entry.Path)
	}
	if internalDestinationPath(entry.Path) {
		if excludeInternal {
			return nil
		}
		return fmt.Errorf("source listing uses reserved .atomic-sync-staging path %q", entry.Path)
	}
	if existing, exists := result.kinds[entry.Path]; exists {
		if existing != entry.IsDir {
			return fmt.Errorf("listing returned file/directory collision at %q", entry.Path)
		}
		return fmt.Errorf("listing returned ambiguous duplicate path %q", entry.Path)
	}
	result.kinds[entry.Path] = entry.IsDir
	unitName := analysisUnit(entry, job)
	if unitName == "" {
		if entry.IsDir {
			return nil
		}
		return fmt.Errorf("listing returned unmappable path %q", entry.Path)
	}
	unit := result.units[unitName]
	if unit == nil {
		unit = &inventoryUnit{files: map[string]int64{}, directories: map[string]struct{}{}}
		result.units[unitName] = unit
	}
	relative := strings.TrimPrefix(strings.TrimPrefix(entry.Path, unitName), "/")
	if relative != "" && !safeRelative(relative) {
		return fmt.Errorf("listing returned unsafe relative path %q", relative)
	}
	return addInventoryUnitEntry(unit, relative, entry)
}

func internalDestinationPath(relative string) bool {
	return relative == ".atomic-sync-staging" || strings.HasPrefix(relative, ".atomic-sync-staging/")
}

func addUnmappedEmptyUnits(result *inventory, job model.Job) {
	candidates := make([]string, 0)
	for relative, isDirectory := range result.kinds {
		if !isDirectory || analysisUnit(listed{Path: relative, IsDir: true}, job) != "" {
			continue
		}
		if hasUnitAtOrBelow(result.units, relative) {
			continue
		}
		candidates = append(candidates, relative)
	}
	sort.Slice(candidates, func(i, k int) bool {
		return strings.Count(candidates[i], "/") > strings.Count(candidates[k], "/")
	})
	for _, relative := range candidates {
		if hasUnitAtOrBelow(result.units, relative) {
			continue
		}
		result.units[relative] = &inventoryUnit{
			files: map[string]int64{}, directories: map[string]struct{}{"": {}},
		}
	}
}

func hasUnitAtOrBelow(units map[string]*inventoryUnit, relative string) bool {
	if relative == "" {
		return false
	}
	for unitName := range units {
		if unitName == relative || strings.HasPrefix(unitName, relative+"/") {
			return true
		}
	}
	return false
}

func addInventoryUnitEntry(unit *inventoryUnit, relative string, entry listed) error {
	if entry.IsDir {
		if _, exists := unit.files[relative]; exists {
			return fmt.Errorf("listing returned file/directory collision at %q", entry.Path)
		}
		unit.directories[relative] = struct{}{}
		return nil
	}
	if entry.Size < 0 {
		return fmt.Errorf("listing returned negative size for file %q", entry.Path)
	}
	if _, exists := unit.directories[relative]; exists {
		return fmt.Errorf("listing returned file/directory collision at %q", entry.Path)
	}
	unit.files[relative] = entry.Size
	return nil
}

func analysisUnit(entry listed, job model.Job) string {
	if !entry.IsDir {
		return analysisFileUnit(entry.Path, job)
	}
	clean := path.Clean(strings.ReplaceAll(entry.Path, `\\`, "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return ""
	}
	parts := strings.Split(clean, "/")
	switch job.Grouping {
	case "folder", "show":
		return parts[0]
	case "season":
		if len(parts) >= 2 {
			return strings.Join(parts[:2], "/")
		}
	case "depth":
		if job.Depth > 0 && len(parts) >= job.Depth {
			return strings.Join(parts[:job.Depth], "/")
		}
	}
	return ""
}

func analysisFileUnit(relative string, job model.Job) string {
	clean := path.Clean(strings.ReplaceAll(relative, `\\`, "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return ""
	}
	parts := strings.Split(clean, "/")
	n := 1
	switch job.Grouping {
	case "season":
		if len(parts) >= 3 {
			n = 2
		}
	case "depth":
		n = min(max(1, job.Depth), len(parts))
	case "folder", "show":
	default:
		return ""
	}
	return strings.Join(parts[:n], "/")
}

func (r *Runner) analysisDestination(ctx context.Context, job model.Job, unit string, source *inventoryUnit, destinations map[string]inventory) (model.Destination, error) {
	name, err := r.store.Assignment(ctx, job.ID, unit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.Destination{}, err
	}
	if err == nil {
		for _, destination := range job.Destinations {
			if destination.Name == name {
				return destination, nil
			}
		}
		return model.Destination{}, fmt.Errorf("assigned destination %q no longer exists", name)
	}
	if source == nil || len(source.files) == 0 {
		for _, destination := range job.Destinations {
			if unitInventory := destinations[destination.Name].units[unit]; unitInventory != nil && len(unitInventory.files) > 0 {
				return destination, nil
			}
		}
		for _, destination := range job.Destinations {
			if destinations[destination.Name].units[unit] != nil {
				return destination, nil
			}
		}
	}
	weights := make([]int, len(job.Destinations))
	for index := range job.Destinations {
		weights[index] = job.Destinations[index].Weight
	}
	return job.Destinations[PickDestination(unit, weights)], nil
}

func compareInventory(unitName, destination string, source, target *inventoryUnit, sourceKinds, targetKinds map[string]bool) model.UnitAnalysis {
	analysis := model.UnitAnalysis{Unit: unitName, Destination: destination}
	populateInventoryTotals(&analysis, source, target)
	if _, exists := sourceKinds[unitName]; exists {
		analysis.SourcePresent = true
	}
	if _, exists := targetKinds[unitName]; exists {
		analysis.DestinationPresent = true
	}
	compareSourceFiles(&analysis, unitName, source, target)
	compareDestinationOnlyFiles(&analysis, unitName, source, target)
	if source == nil || target == nil {
		compareRootType(&analysis, unitName, sourceKinds, targetKinds)
	}
	if analysis.SourceFiles > 0 {
		analysis.Coverage = analysis.MatchingFiles * 100 / analysis.SourceFiles
	}
	return analysis
}

func finalizeInventoryAnalysis(analysis *model.UnitAnalysis) {
	analysis.Status = classifyInventory(*analysis)
	if analysis.Status == model.ArchiveArchived {
		analysis.Coverage = 100
	}
}

func compareRootType(analysis *model.UnitAnalysis, unitName string, sourceKinds, targetKinds map[string]bool) {
	sourceIsDirectory, sourceExists := sourceKinds[unitName]
	targetIsDirectory, targetExists := targetKinds[unitName]
	if sourceExists && targetExists && sourceIsDirectory != targetIsDirectory {
		analysis.ConflictingFiles++
		analysis.ConflictSamples = appendSample(analysis.ConflictSamples, unitName)
	}
}

func populateInventoryTotals(analysis *model.UnitAnalysis, source, target *inventoryUnit) {
	if source != nil {
		analysis.SourcePresent = true
		analysis.SourceFiles = len(source.files)
		for _, size := range source.files {
			analysis.SourceBytes += size
		}
	}
	if target == nil {
		return
	}
	analysis.DestinationPresent = true
	analysis.DestinationFiles = len(target.files)
	for _, size := range target.files {
		analysis.DestinationBytes += size
	}
}

func compareSourceFiles(analysis *model.UnitAnalysis, unitName string, source, target *inventoryUnit) {
	if source == nil || target == nil {
		return
	}
	for _, relative := range sortedFilePaths(source.files) {
		sourceSize := source.files[relative]
		targetSize, exists := target.files[relative]
		_, targetIsDirectory := target.directories[relative]
		switch {
		case targetIsDirectory:
			analysis.ConflictingFiles++
			analysis.ConflictSamples = appendSample(analysis.ConflictSamples, analysisPath(unitName, relative))
		case !exists:
			analysis.MissingFiles++
			analysis.MissingSamples = appendSample(analysis.MissingSamples, analysisPath(unitName, relative))
		case sourceSize != targetSize:
			analysis.ConflictingFiles++
			analysis.ConflictSamples = appendSample(analysis.ConflictSamples, analysisPath(unitName, relative))
		default:
			analysis.MatchingFiles++
			analysis.MatchingBytes += sourceSize
		}
	}
	for _, relative := range sortedDirectoryPaths(source.directories) {
		if _, targetIsFile := target.files[relative]; targetIsFile {
			analysis.ConflictingFiles++
			analysis.ConflictSamples = appendSample(analysis.ConflictSamples, analysisPath(unitName, relative))
		}
	}
}

func compareDestinationOnlyFiles(analysis *model.UnitAnalysis, unitName string, source, target *inventoryUnit) {
	if target == nil {
		return
	}
	for _, relative := range sortedFilePaths(target.files) {
		targetSize := target.files[relative]
		if source != nil {
			if _, exists := source.files[relative]; exists {
				continue
			}
			if _, sourceIsDirectory := source.directories[relative]; sourceIsDirectory {
				continue
			}
		}
		analysis.DestinationOnlyFiles++
		analysis.DestinationOnlyBytes += targetSize
		analysis.DestinationOnlySamples = appendSample(analysis.DestinationOnlySamples, analysisPath(unitName, relative))
	}
}

func compareUnexpectedDestinations(analysis *model.UnitAnalysis, unitName, selected string, configured []model.Destination, inventories map[string]inventory) {
	for _, destination := range configured {
		if destination.Name == selected {
			continue
		}
		unit := inventories[destination.Name].units[unitName]
		if unit == nil {
			continue
		}
		analysis.UnexpectedDestinations = append(analysis.UnexpectedDestinations, destination.Name)
		for _, relative := range sortedFilePaths(unit.files) {
			size := unit.files[relative]
			analysis.UnexpectedDestinationFiles++
			analysis.UnexpectedDestinationBytes += size
			analysis.ConflictingFiles++
			value := destination.Name + ":" + analysisPath(unitName, relative)
			analysis.ConflictSamples = appendSample(analysis.ConflictSamples, value)
		}
	}
}

func sortedFilePaths(files map[string]int64) []string {
	paths := make([]string, 0, len(files))
	for relative := range files {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	return paths
}

func sortedDirectoryPaths(directories map[string]struct{}) []string {
	paths := make([]string, 0, len(directories))
	for relative := range directories {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	return paths
}

func classifyInventory(analysis model.UnitAnalysis) string {
	switch {
	case analysis.ConflictingFiles > 0:
		return model.ArchiveConflict
	case analysis.SourceFiles == 0 && analysis.DestinationFiles == 0:
		return model.ArchiveEmpty
	case analysis.SourceFiles == 0:
		return model.ArchiveArchived
	case analysis.DestinationFiles == 0:
		return model.ArchivePending
	case analysis.MissingFiles > 0:
		return model.ArchivePartial
	default:
		return model.ArchiveReadyToVerify
	}
}

func analysisPath(unitName, relative string) string {
	if relative == "" {
		return unitName
	}
	return path.Join(unitName, relative)
}

func appendSample(samples []string, value string) []string {
	if len(samples) < 12 {
		return append(samples, value)
	}
	return samples
}

func newAnalysisSummary() map[string]int {
	return map[string]int{
		model.ArchiveArchived: 0, model.ArchiveReadyToVerify: 0, model.ArchivePartial: 0,
		model.ArchivePending: 0, model.ArchiveConflict: 0, model.ArchiveEmpty: 0,
	}
}

func (r *Runner) runUnit(ctx context.Context, j model.Job, unit string) error {
	execution, err := r.prepareUnit(ctx, j, unit)
	if err != nil {
		return err
	}
	fail := func(cause error) error { return r.failUnit(execution, cause) }
	if j.DryRun {
		return r.complete(ctx, execution.run.ID, execution.run, "dry run: discovered and planned; no changes made")
	}
	if err = r.stageUnit(ctx, execution); err != nil {
		return fail(err)
	}
	if err = r.publishUnit(ctx, execution); err != nil {
		return fail(err)
	}
	return r.complete(ctx, execution.run.ID, execution.run, "verified and published; source preserved")
}

func (r *Runner) prepareUnit(ctx context.Context, job model.Job, unit string) (*unitExecution, error) {
	if !safeRelative(unit) {
		return nil, fmt.Errorf("unsafe unit path %q", unit)
	}
	destinationName, err := r.store.Assignment(ctx, job.ID, unit)
	if errors.Is(err, sql.ErrNoRows) {
		weights := make([]int, len(job.Destinations))
		for index := range job.Destinations {
			weights[index] = job.Destinations[index].Weight
		}
		picked := job.Destinations[PickDestination(unit, weights)].Name
		destinationName, err = r.store.Assign(ctx, job.ID, unit, picked)
	}
	if err != nil {
		return nil, err
	}
	var selected model.Destination
	for _, candidate := range job.Destinations {
		if candidate.Name == destinationName {
			selected = candidate
			break
		}
	}
	if selected.Path == "" {
		return nil, fmt.Errorf("assigned destination %q no longer exists", destinationName)
	}
	id := newRunID()
	run := model.Run{ID: id, JobID: job.ID, Unit: unit, Destination: destinationName, State: "discovered", StartedAt: time.Now().UTC()}
	if err = r.store.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	r.emit(Event{Type: "run", Run: run})
	return &unitExecution{
		job: job, unit: unit, destination: selected, destinationName: destinationName, run: run,
		source: join(job.Source, unit),
		stage:  join(selected.Path, path.Join(".atomic-sync-staging", job.ID, id, unit)),
		final:  join(selected.Path, unit),
	}, nil
}

func (r *Runner) failUnit(execution *unitExecution, cause error) error {
	transitionCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.store.Transition(transitionCtx, execution.run.ID, "failed", cause.Error())
	execution.run.State = "failed"
	execution.run.Message = cause.Error()
	r.emit(Event{Type: "run", Run: execution.run})
	return cause
}

func (r *Runner) stageUnit(ctx context.Context, execution *unitExecution) error {
	if err := r.store.Transition(ctx, execution.run.ID, "staging", ""); err != nil {
		return err
	}
	if _, err := r.command(ctx, "copy", execution.source, execution.stage, "--create-empty-src-dirs"); err != nil {
		return err
	}
	if err := r.store.Transition(ctx, execution.run.ID, "verifying", ""); err != nil {
		return err
	}
	if err := r.check(ctx, execution.source, execution.stage, execution.job, false); err != nil {
		return err
	}
	return r.store.Transition(ctx, execution.run.ID, "publishing", "")
}

func (r *Runner) publishUnit(ctx context.Context, execution *unitExecution) error {
	switch execution.job.ConflictPolicy {
	case model.ConflictFail:
		if err := r.publishNewUnit(ctx, execution); err != nil {
			return err
		}
	case model.ConflictMergeImmutable:
		if err := r.publishImmutableMerge(ctx, execution); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported conflict policy %q", execution.job.ConflictPolicy)
	}
	oneWay := execution.job.ConflictPolicy == model.ConflictMergeImmutable
	if err := r.check(ctx, execution.source, execution.final, execution.job, oneWay); err != nil {
		return fmt.Errorf("final verification failed; source preserved: %w", err)
	}
	return nil
}

func (r *Runner) publishNewUnit(ctx context.Context, execution *unitExecution) error {
	exists, err := r.pathExists(ctx, execution.final)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("destination already exists: %s", execution.final)
	}
	if _, err = r.command(ctx, "moveto", execution.stage, execution.final, "--immutable"); err != nil {
		return fmt.Errorf("publish failed (staging preserved): %w", err)
	}
	return nil
}

func (r *Runner) publishImmutableMerge(ctx context.Context, execution *unitExecution) error {
	if _, err := r.command(ctx, "copy", execution.stage, execution.final, "--immutable", "--create-empty-src-dirs"); err != nil {
		return fmt.Errorf("immutable merge failed (source and staging preserved): %w", err)
	}
	if err := r.check(ctx, execution.stage, execution.final, execution.job, true); err != nil {
		return fmt.Errorf("immutable merge verification failed: %w", err)
	}
	return nil
}

func (r *Runner) check(ctx context.Context, source, destination string, j model.Job, oneWay bool) error {
	// Successful rclone checks emit routine NOTICE summaries on stderr. Quiet
	// suppresses those summaries while preserving non-zero exits and ERROR
	// diagnostics for mismatches and I/O failures.
	args := []string{"check", source, destination, "--quiet"}
	if oneWay {
		args = append(args, "--one-way")
	}
	if j.Verify == "size" {
		args = append(args, "--size-only")
	} else {
		args = append(args, "--download")
	}
	_, err := r.command(ctx, args...)
	return err
}

func (r *Runner) pathExists(ctx context.Context, target string) (bool, error) {
	out, err := r.command(ctx, "lsf", target, "--max-depth", "1")
	if err == nil {
		return true, nil
	}
	if isNotFound(out, err) {
		return false, nil
	}
	return false, fmt.Errorf("inspect destination: %w", err)
}

func isNotFound(output []byte, err error) bool {
	if err == nil {
		return false
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode() == 3
	}
	message := strings.ToLower(string(output) + " " + err.Error())
	for _, marker := range []string{"directory not found", "doesn't exist", "does not exist"} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}

func (r *Runner) complete(ctx context.Context, id string, run model.Run, message string) error {
	if err := r.store.Transition(ctx, id, "completed", message); err != nil {
		return err
	}
	run.State = "completed"
	run.Message = message
	r.emit(Event{Type: "run", Run: run})
	return nil
}

func (r *Runner) command(ctx context.Context, args ...string) ([]byte, error) {
	select {
	case r.sem <- struct{}{}:
		defer func() { <-r.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	limited := make([]string, 0, len(args)+len(r.rcloneArgs))
	limited = append(limited, args...)
	limited = append(limited, r.rcloneArgs...)
	return r.execute(ctx, limited...)
}

func (r *Runner) execRclone(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.rclone, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	rawDiagnostic := stderr.String()
	diagnostic := boundedDiagnostic(stderr.Bytes())
	if err != nil {
		combined := append(append([]byte(nil), stdout.Bytes()...), stderr.Bytes()...)
		if diagnostic == "" {
			diagnostic = "no diagnostic output"
		}
		return combined, fmt.Errorf("rclone %s: %w: %s", args[0], err, diagnostic)
	}
	if diagnostic != "" {
		if len(args) > 0 && args[0] == "lsjson" {
			if onlySharedDriveClientIDNotices(rawDiagnostic) {
				slog.Warn("rclone inventory uses the retiring shared Google Drive client_id; configure a dedicated OAuth client")
				return stdout.Bytes(), nil
			}
			return stdout.Bytes(), fmt.Errorf("rclone lsjson emitted diagnostics; inventory is not trusted: %s", diagnostic)
		}
		slog.Warn("rclone command emitted diagnostics", "command", args[0], "diagnostic", diagnostic)
	}
	return stdout.Bytes(), nil
}

const sharedDriveClientIDNotice = `This remote uses rclone's shared Google Drive client_id, which is being retired and will stop working during 2026. Create your own client_id to avoid interruption: https://rclone.org/drive/#making-your-own-client-id`
const maxTrustedRcloneNoticeBytes = 64 * 1024

func onlySharedDriveClientIDNotices(diagnostic string) bool {
	const noticeMarker = " NOTICE: "
	const logTimeLayout = "2006/01/02 15:04:05"
	if diagnostic == "" || len(diagnostic) > maxTrustedRcloneNoticeBytes {
		return false
	}
	lines := strings.Split(diagnostic, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return false
	}
	for _, line := range lines {
		if line == "" || strings.TrimSpace(line) != line {
			return false
		}
		marker := strings.Index(line, noticeMarker)
		if marker != len(logTimeLayout) || strings.Count(line, noticeMarker) != 1 {
			return false
		}
		if _, err := time.Parse(logTimeLayout, line[:marker]); err != nil {
			return false
		}
		body := line[marker+len(noticeMarker):]
		suffix := ": " + sharedDriveClientIDNotice
		if !strings.HasSuffix(body, suffix) {
			return false
		}
		remote := strings.TrimSuffix(body, suffix)
		if !safeRcloneRemoteLogName(remote) {
			return false
		}
	}
	return true
}

func safeRcloneRemoteLogName(remote string) bool {
	if remote == "" || len(remote) > 64 {
		return false
	}
	for _, char := range remote {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func boundedDiagnostic(value []byte) string {
	message := strings.TrimSpace(string(value))
	if len(message) > 16*1024 {
		message = message[len(message)-16*1024:]
	}
	return message
}

func newRunID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err == nil {
		return "run_" + hex.EncodeToString(value)
	}
	return fmt.Sprintf("run_%d", time.Now().UnixNano())
}

func safeRelative(value string) bool {
	if value == "" || strings.ContainsAny(value, "\x00\r\n") || strings.HasPrefix(value, "/") {
		return false
	}
	clean := path.Clean(strings.ReplaceAll(value, `\\`, "/"))
	return clean == value && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}

func join(root, child string) string {
	if strings.HasSuffix(root, ":") {
		return root + strings.TrimPrefix(child, "/")
	}
	if index := strings.Index(root, ":"); index > 0 {
		return root[:index+1] + path.Join(root[index+1:], child)
	}
	return path.Join(root, child)
}
