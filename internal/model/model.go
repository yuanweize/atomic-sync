package model

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	ConflictFail           = "fail"
	ConflictMergeImmutable = "merge-immutable"

	ArchiveArchived      = "archived"
	ArchiveReadyToVerify = "ready-to-verify"
	ArchivePartial       = "partial"
	ArchivePending       = "pending"
	ArchiveConflict      = "conflict"
	ArchiveEmpty         = "empty"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type Destination struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Weight int    `json:"weight"`
}

type Job struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Source         string        `json:"source"`
	Destinations   []Destination `json:"destinations"`
	Mode           string        `json:"mode"`
	Grouping       string        `json:"grouping"`
	Depth          int           `json:"depth"`
	Schedule       string        `json:"schedule"`
	SettleSeconds  int           `json:"settleSeconds"`
	Concurrency    int           `json:"concurrency"`
	Verify         string        `json:"verify"`
	ConflictPolicy string        `json:"conflictPolicy"`
	DeleteSource   bool          `json:"deleteSource"`
	DryRun         bool          `json:"dryRun"`
	Paused         bool          `json:"paused"`
	Include        []string      `json:"include,omitempty"`
	Exclude        []string      `json:"exclude,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type Run struct {
	ID          string     `json:"id"`
	JobID       string     `json:"jobId"`
	Unit        string     `json:"unit"`
	Destination string     `json:"destination"`
	State       string     `json:"state"`
	Message     string     `json:"message,omitempty"`
	StartedAt   time.Time  `json:"startedAt"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
}

// Analysis is a metadata-first comparison of the physical source and
// destination branches. Matching size is useful for planning, but it is not
// proof that two files have identical contents.
type Analysis struct {
	JobID      string         `json:"jobId"`
	State      string         `json:"state"`
	Message    string         `json:"message,omitempty"`
	Summary    map[string]int `json:"summary"`
	Units      []UnitAnalysis `json:"units,omitempty"`
	StartedAt  time.Time      `json:"startedAt"`
	FinishedAt *time.Time     `json:"finishedAt,omitempty"`
}

type UnitAnalysis struct {
	Unit                       string   `json:"unit"`
	Destination                string   `json:"destination"`
	Status                     string   `json:"status"`
	Coverage                   int      `json:"coverage"`
	SourcePresent              bool     `json:"sourcePresent"`
	DestinationPresent         bool     `json:"destinationPresent"`
	SourceFiles                int      `json:"sourceFiles"`
	DestinationFiles           int      `json:"destinationFiles"`
	DestinationOnlyFiles       int      `json:"destinationOnlyFiles"`
	MatchingFiles              int      `json:"matchingFiles"`
	MissingFiles               int      `json:"missingFiles"`
	ConflictingFiles           int      `json:"conflictingFiles"`
	SourceBytes                int64    `json:"sourceBytes"`
	DestinationBytes           int64    `json:"destinationBytes"`
	DestinationOnlyBytes       int64    `json:"destinationOnlyBytes"`
	MatchingBytes              int64    `json:"matchingBytes"`
	UnexpectedDestinationFiles int      `json:"unexpectedDestinationFiles"`
	UnexpectedDestinationBytes int64    `json:"unexpectedDestinationBytes"`
	UnexpectedDestinations     []string `json:"unexpectedDestinations,omitempty"`
	MissingSamples             []string `json:"missingSamples,omitempty"`
	DestinationOnlySamples     []string `json:"destinationOnlySamples,omitempty"`
	ConflictSamples            []string `json:"conflictSamples,omitempty"`
}

var transitions = map[string]map[string]bool{
	"discovered": {"staging": true, "completed": true, "failed": true},
	"staging":    {"verifying": true, "failed": true},
	"verifying":  {"publishing": true, "failed": true},
	"publishing": {"completed": true, "failed": true},
	"failed":     {"staging": true},
}

func CanTransition(from, to string) bool { return transitions[from][to] }

// Normalize applies non-destructive defaults. DryRun is intentionally handled
// by the API because a bool alone cannot distinguish omitted from false.
func (j *Job) Normalize() {
	j.Name = strings.TrimSpace(j.Name)
	j.Source = strings.TrimSpace(j.Source)
	if j.Mode == "" {
		j.Mode = "copy"
	}
	if j.Grouping == "" {
		j.Grouping = "folder"
	}
	if j.Concurrency == 0 {
		j.Concurrency = 2
	}
	if j.Verify == "" {
		j.Verify = "checksum"
	}
	if j.ConflictPolicy == "" {
		j.ConflictPolicy = ConflictFail
	}
	for i := range j.Destinations {
		j.Destinations[i].Name = strings.TrimSpace(j.Destinations[i].Name)
		j.Destinations[i].Path = strings.TrimSpace(j.Destinations[i].Path)
	}
}

// SamePlacement reports whether two jobs resolve every unit to the same
// physical source, grouping boundary, and weighted destinations.
func (j Job) SamePlacement(other Job) bool {
	if j.Source != other.Source || j.Grouping != other.Grouping || j.Depth != other.Depth || len(j.Destinations) != len(other.Destinations) {
		return false
	}
	for index := range j.Destinations {
		if j.Destinations[index] != other.Destinations[index] {
			return false
		}
	}
	return true
}

// PlacementOverlaps reports whether two jobs can address any equal or nested
// physical path. Rejecting these configurations prevents independent workers
// from analyzing, publishing, or deleting inside one another's units.
func (j Job) PlacementOverlaps(other Job) bool {
	left := make([]string, 1, len(j.Destinations)+1)
	left[0] = j.Source
	for _, destination := range j.Destinations {
		left = append(left, destination.Path)
	}
	right := make([]string, 1, len(other.Destinations)+1)
	right[0] = other.Source
	for _, destination := range other.Destinations {
		right = append(right, destination.Path)
	}
	for _, a := range left {
		for _, b := range right {
			if overlappingEndpoint(a, b) {
				return true
			}
		}
	}
	return false
}

func (j Job) Validate() error {
	if j.ID != "" && !identifierPattern.MatchString(j.ID) {
		return invalid("id must contain only letters, numbers, dot, underscore, or dash")
	}
	if j.Name == "" || len(j.Name) > 128 {
		return invalid("name is required and must be at most 128 characters")
	}
	if !validSourceEndpoint(j.Source) {
		return invalid("source must be an absolute path below /sources")
	}
	if err := j.validateExecutionOptions(); err != nil {
		return err
	}
	if err := j.validateDestinations(); err != nil {
		return err
	}
	return j.validateFilters()
}

func (j Job) validateExecutionOptions() error {
	if len(j.Destinations) == 0 || len(j.Destinations) > 16 {
		return invalid("between 1 and 16 destinations are required")
	}
	if j.Mode != "copy" {
		return invalid("mode must be copy; automatic source deletion is not supported")
	}
	if j.Grouping != "folder" && j.Grouping != "show" && j.Grouping != "season" && j.Grouping != "depth" {
		return invalid("grouping must be folder, show, season, or depth")
	}
	if j.Grouping == "depth" && j.Depth < 1 {
		return invalid("depth grouping requires a positive depth")
	}
	if j.Depth > 32 {
		return invalid("depth must not exceed 32")
	}
	if j.SettleSeconds < 0 || j.SettleSeconds > int((10*365*24*time.Hour)/time.Second) {
		return invalid("settleSeconds is outside the supported range")
	}
	if j.Concurrency < 1 || j.Concurrency > 8 {
		return invalid("concurrency must be between 1 and 8")
	}
	if j.Verify != "checksum" && j.Verify != "size" {
		return invalid("verify must be checksum or size")
	}
	if j.ConflictPolicy != ConflictFail && j.ConflictPolicy != ConflictMergeImmutable {
		return invalid("conflictPolicy must be fail or merge-immutable")
	}
	if j.DeleteSource {
		return invalid("deleteSource is not supported; source data is always preserved")
	}
	return nil
}

func (j Job) validateDestinations() error {
	names := make(map[string]struct{}, len(j.Destinations))
	paths := make([]string, 0, len(j.Destinations))
	for _, d := range j.Destinations {
		if !identifierPattern.MatchString(d.Name) {
			return invalid("destination names must be safe identifiers")
		}
		if _, exists := names[d.Name]; exists {
			return invalid("destination names must be unique")
		}
		names[d.Name] = struct{}{}
		if !validDestinationEndpoint(d.Path) {
			return invalid("destination paths must be rclone paths or absolute paths below /destinations")
		}
		if overlappingEndpoint(j.Source, d.Path) {
			return invalid("source and destination must not overlap")
		}
		for _, existing := range paths {
			if overlappingEndpoint(existing, d.Path) {
				return invalid("destination paths must not overlap")
			}
		}
		if d.Weight < 1 || d.Weight > 1000 {
			return invalid("destination weight must be between 1 and 1000")
		}
		paths = append(paths, d.Path)
	}
	return nil
}

func (j Job) validateFilters() error {
	if len(j.Include) > 0 || len(j.Exclude) > 0 {
		return invalid("include and exclude filters are not supported; each directory unit is copied in full")
	}
	return nil
}

func validSourceEndpoint(value string) bool {
	return validLocalEndpointBelow(value, "/sources")
}

func validDestinationEndpoint(value string) bool {
	return validRemoteEndpoint(value) || validLocalEndpointBelow(value, "/destinations")
}

func validRemoteEndpoint(value string) bool {
	if value == "" || len(value) > 1024 || strings.HasPrefix(value, "-") || strings.ContainsAny(value, "\x00\r\n") {
		return false
	}
	if i := strings.Index(value, ":"); i > 0 && !strings.HasPrefix(value, "/") {
		remotePath := strings.Trim(value[i+1:], "/")
		return identifierPattern.MatchString(value[:i]) && remotePath != "" && remotePath != "." && !strings.HasPrefix(path.Clean(remotePath), "../")
	}
	return false
}

func validLocalEndpointBelow(value, root string) bool {
	if value == "" || len(value) > 1024 || strings.HasPrefix(value, "-") || strings.ContainsAny(value, "\\\x00\r\n") {
		return false
	}
	clean := path.Clean(value)
	return strings.HasPrefix(clean, root+"/")
}

func overlappingEndpoint(a, b string) bool {
	kindAndPath := func(value string) (string, string) {
		value = strings.ReplaceAll(value, `\\`, "/")
		if index := strings.Index(value, ":"); index > 0 && !strings.HasPrefix(value, "/") {
			return value[:index], "/" + strings.Trim(path.Clean(value[index+1:]), "/")
		}
		return "local", strings.TrimSuffix(path.Clean(value), "/")
	}
	ak, ap := kindAndPath(a)
	bk, bp := kindAndPath(b)
	if ak != bk {
		return false
	}
	return ap == bp || strings.HasPrefix(ap, bp+"/") || strings.HasPrefix(bp, ap+"/")
}

type validationError string

func (e validationError) Error() string { return string(e) }

func (e validationError) Is(target error) bool { return target == ErrInvalidJob }

func invalid(reason string) error {
	return validationError(fmt.Sprintf("invalid job configuration: %s", reason))
}

const ErrInvalidJob validationError = "invalid job configuration"
