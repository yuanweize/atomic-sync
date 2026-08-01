package model

import (
	"errors"
	"strings"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	for _, transition := range [][2]string{
		{"discovered", "transferring"}, {"transferring", "completed"},
	} {
		if !CanTransition(transition[0], transition[1]) {
			t.Fatalf("expected %s -> %s to be valid", transition[0], transition[1])
		}
	}
	if CanTransition("transferring", "publishing") || CanTransition("discovered", "staging") {
		t.Fatal("direct transfers must not enter legacy staging or publishing states")
	}
}

func TestNormalizeUsesSafeDefaults(t *testing.T) {
	job := Job{Name: " archive ", Source: " /source/media ", Destinations: []Destination{{Name: " gd ", Path: " GD:data/media ", Weight: 1}}}
	job.Normalize()
	if job.Name != "archive" || job.Source != "/source/media" || job.Destinations[0].Name != "gd" {
		t.Fatalf("values were not trimmed: %#v", job)
	}
	if job.Mode != "copy" || job.Grouping != "folder" || job.Verify != "checksum" || job.ConflictPolicy != ConflictFail || job.Concurrency != 2 {
		t.Fatalf("unsafe or missing defaults: %#v", job)
	}
}

func TestJobValidation(t *testing.T) {
	valid := func() Job {
		return Job{
			ID: "job_safe-1", Name: "Archive movies", Source: "/sources/storagebox/movies",
			Destinations: []Destination{{Name: "gd-primary", Path: "GD:data/media/movies", Weight: 1}},
			Mode:         "copy", Grouping: "folder", Concurrency: 2, Verify: "checksum", ConflictPolicy: ConflictFail,
		}
	}
	if err := valid().Validate(); err != nil {
		t.Fatalf("valid job rejected: %v", err)
	}
	move := valid()
	move.Mode = ModeMove
	move.DeleteSource = true
	move.Verify = "size"
	if err := move.Validate(); err != nil {
		t.Fatalf("valid fail-closed size-verified move job rejected: %v", err)
	}
	move.ConflictPolicy = ConflictMergeImmutable
	move.Verify = "checksum"
	if err := move.Validate(); err != nil {
		t.Fatalf("valid immutable checksum move job rejected: %v", err)
	}
	filteredCopy := valid()
	filteredCopy.Include = []string{"*.mkv"}
	if err := filteredCopy.Validate(); err == nil || !strings.Contains(err.Error(), "copied in full") {
		t.Fatalf("filtered copy job was accepted: %v", err)
	} else if !errors.Is(err, ErrInvalidJob) {
		t.Fatalf("validation error does not match ErrInvalidJob: %v", err)
	}
	tests := map[string]func(*Job){
		"unsafe id":    func(job *Job) { job.ID = `x"><script>` },
		"root source":  func(job *Job) { job.Source = "/" },
		"config exfil": func(job *Job) { job.Source = "/config" },
		"legacy staging source": func(job *Job) {
			job.Source = "/sources/storagebox/.atomic-sync-staging/recovery"
		},
		"backslash escape":    func(job *Job) { job.Source = `/sources\\media` },
		"remote source":       func(job *Job) { job.Source = "GD:data/media" },
		"remote root":         func(job *Job) { job.Source = "GD:" },
		"overlap local":       func(job *Job) { job.Destinations[0].Path = "/sources/storagebox/movies/archive" },
		"overlap remote":      func(job *Job) { job.Source, job.Destinations[0].Path = "GD:data", "GD:data/media" },
		"zero weight":         func(job *Job) { job.Destinations[0].Weight = 0 },
		"unsupported sync":    func(job *Job) { job.Mode = "sync" },
		"high concurrency":    func(job *Job) { job.Concurrency = 9 },
		"bad conflict":        func(job *Job) { job.ConflictPolicy = "overwrite" },
		"bad filter":          func(job *Job) { job.Include = []string{"*.mkv\n--delete"} },
		"move without delete": func(job *Job) { job.Mode = ModeMove },
		"copy with delete":    func(job *Job) { job.DeleteSource = true },
		"local dest escape":   func(job *Job) { job.Destinations[0].Path = "/data/archive" },
		"legacy staging destination": func(job *Job) {
			job.Destinations[0].Path = "GD:data/.atomic-sync-staging/recovery"
		},
		"duplicate dest": func(job *Job) {
			job.Destinations = append(job.Destinations, Destination{Name: "gd-primary", Path: "GD2:data/media", Weight: 1})
		},
		"overlapping dest": func(job *Job) {
			job.Destinations = append(job.Destinations, Destination{Name: "gd-secondary", Path: "GD:data/media/movies/archive", Weight: 1})
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			job := valid()
			mutate(&job)
			if err := job.Validate(); err == nil || !strings.Contains(err.Error(), "invalid job configuration") {
				t.Fatalf("invalid job accepted: %#v, error=%v", job, err)
			}
		})
	}
}

func TestPlacementOverlapAcrossJobs(t *testing.T) {
	movies := Job{
		Source:       "/sources/storagebox/movies",
		Destinations: []Destination{{Name: "gd", Path: "GD:data/media/movies", Weight: 1}},
	}
	tv := Job{
		Source:       "/sources/storagebox/tvseries",
		Destinations: []Destination{{Name: "gd", Path: "GD:data/media/tvseries", Weight: 1}},
	}
	if movies.PlacementOverlaps(tv) {
		t.Fatal("sibling movie and TV jobs must not overlap")
	}
	nested := tv
	nested.Source = "/sources/storagebox/movies/new"
	if !movies.PlacementOverlaps(nested) {
		t.Fatal("nested local source was not detected")
	}
	nested = tv
	nested.Destinations[0].Path = "GD:data/media/movies/archive"
	if !movies.PlacementOverlaps(nested) {
		t.Fatal("nested remote destination was not detected")
	}
}
