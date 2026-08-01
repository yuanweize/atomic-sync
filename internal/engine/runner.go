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
	"os"
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

// Keep ignoring staging data left by pre-v0.2 runs; new transfers never create it.
const destinationStagingNamespace = model.LegacyStagingNamespace

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

type unitFingerprintEntry struct {
	isDir   bool
	size    int64
	modTime time.Time
}

type unitFingerprint map[string]unitFingerprintEntry

type unitPlan struct {
	path        string
	fingerprint unitFingerprint
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
	fingerprint     unitFingerprint
	destination     model.Destination
	destinationName string
	run             model.Run
	source          string
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
	if tpsLimit < 0 {
		tpsLimit = 2
	}
	ctx, cancel := context.WithCancel(context.Background())
	rcloneArgs := []string{"--transfers", strconv.Itoa(transfers), "--checkers", strconv.Itoa(checkers)}
	if tpsLimit > 0 {
		tpsLimit = min(tpsLimit, 64)
		rcloneArgs = append(rcloneArgs, "--tpslimit", strconv.Itoa(tpsLimit), "--tpslimit-burst", strconv.Itoa(tpsLimit))
	}
	r := &Runner{
		store: s, rclone: rclone,
		rcloneArgs: rcloneArgs,
		sem:        make(chan struct{}, concurrency), analysisSem: make(chan struct{}, 1),
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

func (r *Runner) runUnits(ctx context.Context, job model.Job, units []unitPlan) error {
	unitPaths := make([]string, len(units))
	for index := range units {
		unitPaths[index] = units[index].path
	}
	if err := validateIndependentUnits(unitPaths); err != nil {
		return err
	}
	workers := min(max(1, job.Concurrency), len(units))
	unitCh := make(chan unitPlan)
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

func (r *Runner) unitWorker(ctx context.Context, job model.Job, units <-chan unitPlan, workers *sync.WaitGroup, errMu *sync.Mutex, errs *[]error) {
	defer workers.Done()
	for unit := range units {
		err := r.runUnit(ctx, job, unit)
		if err == nil {
			continue
		}
		slog.Error("unit failed", "job", job.Name, "unit", unit.path, "error", err)
		errMu.Lock()
		*errs = append(*errs, fmt.Errorf("%s: %w", unit.path, err))
		errMu.Unlock()
	}
}

func sendUnits(ctx context.Context, target chan<- unitPlan, units []unitPlan) {
	for _, unit := range units {
		select {
		case target <- unit:
		case <-ctx.Done():
			return
		}
	}
}

func (r *Runner) discover(ctx context.Context, j model.Job) ([]unitPlan, error) {
	out, err := r.command(ctx, "lsjson", j.Source, "--recursive")
	if err != nil {
		return nil, err
	}
	var entries []listed
	if err = json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("decode rclone listing: %w", err)
	}
	latest := map[string]time.Time{}
	seen := make(map[string]struct{}, len(entries))
	for index := range entries {
		entry := &entries[index]
		entry.Path = strings.ReplaceAll(entry.Path, `\`, "/")
		if !safeRelative(entry.Path) {
			return nil, fmt.Errorf("listing returned unsafe path %q", entry.Path)
		}
		if _, exists := seen[entry.Path]; exists {
			return nil, fmt.Errorf("listing returned ambiguous duplicate path %q", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		if entry.IsDir {
			continue
		}
		if entry.Size < 0 {
			return nil, fmt.Errorf("listing returned negative size for file %q", entry.Path)
		}
		unit := UnitFor(entry.Path, j.Grouping, j.Depth)
		if unit == "" {
			return nil, fmt.Errorf("file %q is not inside a valid %s directory unit", entry.Path, j.Grouping)
		}
		if entry.ModTime.IsZero() {
			if j.SettleSeconds > 0 {
				return nil, fmt.Errorf("file %q has no modification time; stable-window eligibility cannot be proven", entry.Path)
			}
			if _, exists := latest[unit]; !exists {
				latest[unit] = time.Time{}
			}
			continue
		}
		if entry.ModTime.After(latest[unit]) {
			latest[unit] = entry.ModTime.Time
		}
	}
	cutoff := time.Now().Add(-time.Duration(j.SettleSeconds) * time.Second)
	unitPaths := make([]string, 0, len(latest))
	for unit, modified := range latest {
		if j.SettleSeconds <= 0 || modified.Before(cutoff) {
			unitPaths = append(unitPaths, unit)
		}
	}
	sort.Strings(unitPaths)
	if err = validateIndependentUnits(unitPaths); err != nil {
		return nil, err
	}
	plans := make([]unitPlan, len(unitPaths))
	byPath := make(map[string]*unitPlan, len(unitPaths))
	for index, unit := range unitPaths {
		plans[index] = unitPlan{path: unit, fingerprint: unitFingerprint{}}
		byPath[unit] = &plans[index]
	}
	for _, entry := range entries {
		unit := UnitFor(entry.Path, j.Grouping, j.Depth)
		if entry.IsDir {
			unit = analysisUnit(entry, j)
		}
		plan := byPath[unit]
		if plan == nil || entry.Path == unit {
			continue
		}
		relative := strings.TrimPrefix(entry.Path, unit+"/")
		if !safeRelative(relative) {
			return nil, fmt.Errorf("listing returned unsafe unit-relative path %q", relative)
		}
		plan.fingerprint[relative] = unitFingerprintEntry{
			isDir: entry.IsDir, size: entry.Size, modTime: entry.ModTime.Time,
		}
	}
	return plans, nil
}

func validateIndependentUnits(units []string) error {
	seen := make(map[string]struct{}, len(units))
	for _, unit := range units {
		if !safeRelative(unit) {
			return fmt.Errorf("unsafe unit path %q", unit)
		}
		if internalControlPath(unit) {
			return fmt.Errorf("unit path %q uses a reserved Atomic Sync namespace", unit)
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
		args = append(args, "--exclude", "/"+destinationStagingNamespace+"/**")
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
	if internalControlPath(entry.Path) {
		if excludeInternal {
			return nil
		}
		return fmt.Errorf("source listing uses reserved Atomic Sync path %q", entry.Path)
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

func internalControlPath(relative string) bool {
	return relative == destinationStagingNamespace || strings.HasPrefix(relative, destinationStagingNamespace+"/")
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

func (r *Runner) runUnit(ctx context.Context, j model.Job, unit unitPlan) error {
	execution, err := r.prepareUnit(ctx, j, unit.path)
	if err != nil {
		return err
	}
	execution.fingerprint = unit.fingerprint
	fail := func(cause error) error { return r.failUnit(execution, cause) }
	if err = r.transferUnit(ctx, execution); err != nil {
		return fail(err)
	}
	if j.DryRun {
		return r.complete(execution.run.ID, execution.run, "rclone "+j.Mode+" dry run completed; no source or destination media objects changed")
	}
	if j.Mode == model.ModeMove {
		return r.complete(execution.run.ID, execution.run, "rclone move completed; source files removed by rclone after successful transfer")
	}
	return r.complete(execution.run.ID, execution.run, "rclone copy completed and verified; source preserved")
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
		final:  join(selected.Path, unit),
	}, nil
}

func (r *Runner) failUnit(execution *unitExecution, cause error) error {
	transitionCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.store.Transition(transitionCtx, execution.run.ID, "failed", cause.Error()); err != nil {
		cause = errors.Join(cause, fmt.Errorf("persist failed run state: %w", err))
		slog.Error("failed run state was not persisted", "run", execution.run.ID, "error", err)
	}
	execution.run.State = "failed"
	execution.run.Message = cause.Error()
	r.emit(Event{Type: "run", Run: execution.run})
	return cause
}

func (r *Runner) transferUnit(ctx context.Context, execution *unitExecution) error {
	if err := r.store.Transition(ctx, execution.run.ID, "transferring", ""); err != nil {
		return err
	}
	if execution.job.ConflictPolicy == model.ConflictFail {
		exists, err := r.pathExists(ctx, execution.final)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("destination already exists: %s", execution.final)
		}
	}
	// Keep source revalidation as the last preflight. A destination can still
	// race after its existence check, but --immutable fails closed in that case.
	// Revalidating here minimizes the unprotected interval for an actively
	// written source before rclone starts reading it.
	if err := r.ensureUnitStable(ctx, execution); err != nil {
		return err
	}
	manifest, err := createTransferManifest(execution.fingerprint)
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.Remove(manifest); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			slog.Warn("failed to remove transfer manifest", "path", manifest, "error", removeErr)
		}
	}()
	args := []string{
		execution.job.Mode, execution.source, execution.final,
		"--immutable", "--files-from-raw", manifest,
	}
	if execution.job.SettleSeconds > 0 {
		args = append(args, "--min-age", strconv.Itoa(execution.job.SettleSeconds)+"s")
	}
	if execution.job.Verify == "checksum" {
		args = append(args, "--checksum")
	} else {
		args = append(args, "--size-only")
	}
	if execution.job.Mode == model.ModeMove {
		// Never let rclone's equality fallback turn an existing destination
		// object into authorization to remove the source. In particular,
		// --checksum can fall back to size when the backends share no hash.
		// Native --ignore-existing keeps every overlap at the source while
		// still moving missing objects at full rclone throughput.
		args = append(args, "--ignore-existing", "--delete-empty-src-dirs")
	} else {
		args = append(args, "--create-empty-src-dirs")
	}
	if execution.job.DryRun {
		args = append(args, "--dry-run")
	}
	if _, err := r.command(ctx, args...); err != nil {
		if execution.job.DryRun {
			return fmt.Errorf("rclone %s dry run failed; no source or destination media objects changed: %w", execution.job.Mode, err)
		}
		if execution.job.Mode == model.ModeMove {
			return fmt.Errorf("rclone move failed; source and destination may each contain part of the unit: %w", err)
		}
		return fmt.Errorf("rclone copy failed; source preserved: %w", err)
	}
	if !execution.job.DryRun {
		if err := r.verifyTransferDestination(ctx, execution); err != nil {
			if execution.job.Mode == model.ModeMove {
				return fmt.Errorf("rclone move finished but the destination does not contain the complete discovered unit; treat the unit as partial: %w", err)
			}
			return fmt.Errorf("rclone copy finished but the destination does not contain the complete discovered unit; source preserved: %w", err)
		}
	}
	if execution.job.Mode == model.ModeMove && !execution.job.DryRun {
		hasFiles, err := r.pathHasFiles(ctx, execution.source)
		if err != nil {
			return fmt.Errorf("rclone move transferred data but the remaining source could not be verified; treat the unit as partial: %w", err)
		}
		if hasFiles {
			return errors.New("rclone move preserved source files whose destination paths already existed; the unit is partial and requires review")
		}
	}
	return nil
}

func (r *Runner) ensureUnitStable(ctx context.Context, execution *unitExecution) error {
	if execution.fingerprint == nil {
		return errors.New("source unit has no discovery fingerprint")
	}
	current, err := r.scanUnitFingerprint(ctx, execution.source)
	if err != nil {
		return err
	}
	cutoff := time.Now().Add(-time.Duration(execution.job.SettleSeconds) * time.Second)
	for relative, entry := range current {
		if entry.isDir || execution.job.SettleSeconds <= 0 {
			continue
		}
		if entry.modTime.IsZero() {
			return fmt.Errorf("file %q has no modification time; stable-window eligibility cannot be proven", relative)
		}
		if !entry.modTime.Before(cutoff) {
			return fmt.Errorf("source unit changed after discovery: file %q has not satisfied the stable window", relative)
		}
	}
	return compareUnitFingerprints(execution.fingerprint, current)
}

func (r *Runner) scanUnitFingerprint(ctx context.Context, source string) (unitFingerprint, error) {
	return r.scanFingerprint(ctx, source, "source revalidation")
}

func (r *Runner) scanFingerprint(ctx context.Context, target, operation string) (unitFingerprint, error) {
	out, err := r.command(ctx, "lsjson", target, "--recursive")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	var entries []listed
	if err = json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("decode %s: %w", operation, err)
	}
	fingerprint := make(unitFingerprint, len(entries))
	for _, entry := range entries {
		entry.Path = strings.ReplaceAll(entry.Path, `\`, "/")
		if !safeRelative(entry.Path) {
			return nil, fmt.Errorf("%s returned unsafe path %q", operation, entry.Path)
		}
		if entry.Size < 0 && !entry.IsDir {
			return nil, fmt.Errorf("%s returned negative size for file %q", operation, entry.Path)
		}
		if _, exists := fingerprint[entry.Path]; exists {
			return nil, fmt.Errorf("%s returned ambiguous duplicate path %q", operation, entry.Path)
		}
		fingerprint[entry.Path] = unitFingerprintEntry{
			isDir: entry.IsDir, size: entry.Size, modTime: entry.ModTime.Time,
		}
	}
	return fingerprint, nil
}

func (r *Runner) verifyTransferDestination(ctx context.Context, execution *unitExecution) error {
	current, err := r.scanFingerprint(ctx, execution.final, "verify transfer destination")
	if err != nil {
		return err
	}
	for _, relative := range sortedFingerprintPaths(execution.fingerprint) {
		expected := execution.fingerprint[relative]
		if expected.isDir {
			continue
		}
		actual, exists := current[relative]
		if !exists {
			return fmt.Errorf("destination is missing discovered file %q", relative)
		}
		if actual.isDir {
			return fmt.Errorf("destination path %q is a directory, expected a file", relative)
		}
		if actual.size != expected.size {
			return fmt.Errorf("destination file %q has size %d, expected %d", relative, actual.size, expected.size)
		}
	}
	if execution.job.ConflictPolicy == model.ConflictFail {
		for _, relative := range sortedFingerprintPaths(current) {
			if _, expected := execution.fingerprint[relative]; !expected {
				return fmt.Errorf("destination contains unexpected path %q after fail-closed transfer", relative)
			}
		}
	}
	return nil
}

func createTransferManifest(fingerprint unitFingerprint) (string, error) {
	file, err := os.CreateTemp("", "atomic-sync-files-*.txt")
	if err != nil {
		return "", fmt.Errorf("create transfer manifest: %w", err)
	}
	name := file.Name()
	remove := func() { _ = os.Remove(name) }
	files := 0
	for _, relative := range sortedFingerprintPaths(fingerprint) {
		if fingerprint[relative].isDir {
			continue
		}
		if _, err = file.WriteString(relative + "\n"); err != nil {
			_ = file.Close()
			remove()
			return "", fmt.Errorf("write transfer manifest: %w", err)
		}
		files++
	}
	if err = file.Close(); err != nil {
		remove()
		return "", fmt.Errorf("close transfer manifest: %w", err)
	}
	if files == 0 {
		remove()
		return "", errors.New("source unit fingerprint contains no files")
	}
	return name, nil
}

func compareUnitFingerprints(expected, current unitFingerprint) error {
	currentPaths := sortedFingerprintPaths(current)
	for _, relative := range currentPaths {
		if _, exists := expected[relative]; !exists {
			return fmt.Errorf("source unit changed after discovery: added path %q", relative)
		}
	}
	expectedPaths := sortedFingerprintPaths(expected)
	for _, relative := range expectedPaths {
		currentEntry, exists := current[relative]
		if !exists {
			return fmt.Errorf("source unit changed after discovery: deleted path %q", relative)
		}
		expectedEntry := expected[relative]
		if currentEntry.isDir != expectedEntry.isDir {
			return fmt.Errorf("source unit changed after discovery: type changed for path %q", relative)
		}
		if currentEntry.size != expectedEntry.size {
			return fmt.Errorf("source unit changed after discovery: size changed for path %q", relative)
		}
		if !currentEntry.modTime.Equal(expectedEntry.modTime) {
			return fmt.Errorf("source unit changed after discovery: modification time changed for path %q", relative)
		}
	}
	return nil
}

func sortedFingerprintPaths(fingerprint unitFingerprint) []string {
	paths := make([]string, 0, len(fingerprint))
	for relative := range fingerprint {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	return paths
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

func (r *Runner) pathHasFiles(ctx context.Context, target string) (bool, error) {
	out, err := r.command(ctx, "lsf", target, "--recursive", "--files-only")
	if err == nil {
		return len(bytes.TrimSpace(out)) > 0, nil
	}
	if isNotFound(out, err) {
		return false, nil
	}
	return false, fmt.Errorf("inspect remaining source: %w", err)
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

func (r *Runner) complete(id string, run model.Run, message string) error {
	transitionCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.store.Transition(transitionCtx, id, "completed", message); err != nil {
		return fmt.Errorf("rclone operation completed but durable completion state could not be persisted; outcome requires reconciliation: %w", err)
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
