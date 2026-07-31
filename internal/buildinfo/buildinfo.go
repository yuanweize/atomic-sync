package buildinfo

// These values are replaced at build time through -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
