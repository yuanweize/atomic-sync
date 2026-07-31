package engine

import (
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
	Path    string    `json:"Path"`
	ModTime time.Time `json:"ModTime"`
	IsDir   bool      `json:"IsDir"`
	Size    int64     `json:"Size"`
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
	if concurrency < 1 {
		concurrency = 2
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &Runner{
		store: s, rclone: rclone, sem: make(chan struct{}, concurrency), analysisSem: make(chan struct{}, 1),
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
		select {
		case r.sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		err := r.runUnit(ctx, job, unit)
		<-r.sem
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
		if unit == "" || file.ModTime.IsZero() {
			continue
		}
		if file.ModTime.After(latest[unit]) {
			latest[unit] = file.ModTime
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

func (r *Runner) analyze(ctx context.Context, job model.Job) ([]model.UnitAnalysis, error) {
	source, err := r.inventory(ctx, job.Source, job)
	if err != nil {
		return nil, fmt.Errorf("scan source inventory: %w", err)
	}
	destinations := make(map[string]inventory, len(job.Destinations))
	for _, destination := range job.Destinations {
		items, inventoryErr := r.inventory(ctx, destination.Path, job)
		if inventoryErr != nil {
			return nil, fmt.Errorf("scan destination %s: %w", destination.Name, inventoryErr)
		}
		destinations[destination.Name] = items
	}
	analyses := make([]model.UnitAnalysis, 0, len(source.units))
	for unitName, sourceUnit := range source.units {
		destination, destinationErr := r.analysisDestination(ctx, job, unitName)
		if destinationErr != nil {
			return nil, destinationErr
		}
		destinationInventory := destinations[destination.Name]
		destinationUnit := destinationInventory.units[unitName]
		analysis := compareInventory(unitName, destination.Name, sourceUnit, destinationUnit, source.kinds, destinationInventory.kinds)
		analyses = append(analyses, analysis)
	}
	for _, destination := range job.Destinations {
		destinationInventory := destinations[destination.Name]
		for unitName, destinationUnit := range destinationInventory.units {
			if _, exists := source.units[unitName]; exists {
				continue
			}
			analyses = append(analyses, compareInventory(unitName, destination.Name, nil, destinationUnit, source.kinds, destinationInventory.kinds))
		}
	}
	sort.Slice(analyses, func(i, k int) bool {
		if analyses[i].Unit == analyses[k].Unit {
			return analyses[i].Destination < analyses[k].Destination
		}
		return analyses[i].Unit < analyses[k].Unit
	})
	return analyses, nil
}

func (r *Runner) inventory(ctx context.Context, root string, job model.Job) (inventory, error) {
	out, err := r.command(ctx, "lsjson", root, "--recursive", "--no-modtime", "--no-mimetype")
	if err != nil {
		return inventory{}, err
	}
	var entries []listed
	if err = json.Unmarshal(out, &entries); err != nil {
		return inventory{}, fmt.Errorf("decode inventory listing: %w", err)
	}
	result := inventory{units: map[string]*inventoryUnit{}, kinds: map[string]bool{}}
	for _, entry := range entries {
		if err = addInventoryEntry(&result, entry, job); err != nil {
			return inventory{}, err
		}
	}
	return result, nil
}

func addInventoryEntry(result *inventory, entry listed, job model.Job) error {
	entry.Path = strings.ReplaceAll(entry.Path, `\\`, "/")
	if !safeRelative(entry.Path) {
		return fmt.Errorf("listing returned unsafe path %q", entry.Path)
	}
	if existing, exists := result.kinds[entry.Path]; exists && existing != entry.IsDir {
		return fmt.Errorf("listing returned file/directory collision at %q", entry.Path)
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
		return UnitFor(entry.Path, job.Grouping, job.Depth)
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
		if len(parts) >= job.Depth {
			return strings.Join(parts[:job.Depth], "/")
		}
	}
	return ""
}

func (r *Runner) analysisDestination(ctx context.Context, job model.Job, unit string) (model.Destination, error) {
	name, err := r.store.Assignment(ctx, job.ID, unit)
	if errors.Is(err, sql.ErrNoRows) {
		weights := make([]int, len(job.Destinations))
		for index := range job.Destinations {
			weights[index] = job.Destinations[index].Weight
		}
		return job.Destinations[PickDestination(unit, weights)], nil
	}
	if err != nil {
		return model.Destination{}, err
	}
	for _, destination := range job.Destinations {
		if destination.Name == name {
			return destination, nil
		}
	}
	return model.Destination{}, fmt.Errorf("assigned destination %q no longer exists", name)
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
	if source == nil || target == nil {
		compareRootType(&analysis, unitName, sourceKinds, targetKinds)
	}
	if analysis.SourceFiles > 0 {
		analysis.Coverage = analysis.MatchingFiles * 100 / analysis.SourceFiles
	}
	analysis.Status = classifyInventory(analysis)
	if analysis.Status == model.ArchiveArchived {
		analysis.Coverage = 100
	}
	return analysis
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
	for relative, sourceSize := range source.files {
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
	for relative := range source.directories {
		if _, targetIsFile := target.files[relative]; targetIsFile {
			analysis.ConflictingFiles++
			analysis.ConflictSamples = appendSample(analysis.ConflictSamples, analysisPath(unitName, relative))
		}
	}
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
	if j.Mode == "move" || j.DeleteSource {
		if _, err = r.command(ctx, "purge", execution.source); err != nil {
			return fail(fmt.Errorf("published but source cleanup failed: %w", err))
		}
	}
	return r.complete(ctx, execution.run.ID, execution.run, "verified and published")
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
	copyArgs := appendFilters([]string{"copy", execution.source, execution.stage, "--create-empty-src-dirs"}, execution.job)
	if _, err := r.command(ctx, copyArgs...); err != nil {
		return err
	}
	if err := r.store.Transition(ctx, execution.run.ID, "verifying", ""); err != nil {
		return err
	}
	if err := r.check(ctx, execution.source, execution.stage, execution.job, true); err != nil {
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
	if err := r.check(ctx, execution.source, execution.final, execution.job, true); err != nil {
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
	if err := r.check(ctx, execution.stage, execution.final, execution.job, false); err != nil {
		return fmt.Errorf("immutable merge verification failed: %w", err)
	}
	if _, err := r.command(ctx, "purge", execution.stage); err != nil {
		return fmt.Errorf("published but staging cleanup failed: %w", err)
	}
	return nil
}

func (r *Runner) check(ctx context.Context, source, destination string, j model.Job, filtered bool) error {
	args := []string{"check", source, destination, "--one-way"}
	if j.Verify == "size" {
		args = append(args, "--size-only")
	}
	if filtered {
		args = appendFilters(args, j)
	}
	_, err := r.command(ctx, args...)
	return err
}

func appendFilters(args []string, j model.Job) []string {
	for _, value := range j.Include {
		args = append(args, "--include", value)
	}
	for _, value := range j.Exclude {
		args = append(args, "--exclude", value)
	}
	return args
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
	return r.execute(ctx, args...)
}

func (r *Runner) execRclone(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.rclone, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(out))
		if len(message) > 16*1024 {
			message = message[len(message)-16*1024:]
		}
		return out, fmt.Errorf("rclone %s: %w: %s", args[0], err, message)
	}
	return out, nil
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
