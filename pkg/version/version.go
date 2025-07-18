package version

import "fmt"

// Version information
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// String returns the version string
func String() string {
	return fmt.Sprintf("get-repo version %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}