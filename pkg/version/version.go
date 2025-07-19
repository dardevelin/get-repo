package version

import "fmt"

// Version information
var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
	Author    = "Darcy Brás da Silva"
	Website   = "https://github.com/dardevelin/get-repo"
)

// String returns the version string
func String() string {
	return fmt.Sprintf(`get-repo %s

Author: %s
Source: %s
Commit: %s
Built:  %s`, Version, Author, Website, GitCommit, BuildDate)
}
