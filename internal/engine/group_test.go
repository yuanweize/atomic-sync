package engine

import "testing"

func TestUnitFor(t *testing.T) {
	cases := []struct {
		strategy string
		depth    int
		want     string
	}{
		{"folder", 0, "Andor"}, {"show", 0, "Andor"},
		{"season", 0, "Andor/Season 01"}, {"depth", 2, "Andor/Season 01"},
	}
	for _, tc := range cases {
		if got := UnitFor("Andor/Season 01/E01.mkv", tc.strategy, tc.depth); got != tc.want {
			t.Fatalf("got %q want %q", got, tc.want)
		}
	}
}

func TestUnitForRejectsTraversalAndShallowFiles(t *testing.T) {
	for _, unsafe := range []string{"../movie/file.mkv", "/absolute/file.mkv", `..\\movie\\file.mkv`} {
		if got := UnitFor(unsafe, "folder", 0); got != "" {
			t.Fatalf("unsafe path %q produced unit %q", unsafe, got)
		}
	}
	if got := UnitFor(`Show\\Season 02\\E01.mkv`, "season", 0); got != "Show/Season 02" {
		t.Fatalf("backslash normalization failed: %q", got)
	}
	for _, test := range []struct {
		path     string
		strategy string
		depth    int
	}{
		{path: "movie.mkv", strategy: "folder"},
		{path: "Show/E01.mkv", strategy: "season"},
		{path: "Show/Season 01", strategy: "depth", depth: 2},
		{path: "Show/E01.mkv", strategy: "unknown"},
	} {
		if got := UnitFor(test.path, test.strategy, test.depth); got != "" {
			t.Fatalf("shallow path %q produced %s unit %q", test.path, test.strategy, got)
		}
	}
}

func TestPickDestinationStable(t *testing.T) {
	a := PickDestination("Movie (2026)", []int{1, 2})
	for i := 0; i < 10; i++ {
		if PickDestination("Movie (2026)", []int{1, 2}) != a {
			t.Fatal("assignment changed")
		}
	}
}
