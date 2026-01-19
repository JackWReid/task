package version

import (
	"fmt"
)

// Version is the current version of the application
// This should be set via -ldflags during build:
//   go build -ldflags "-X github.com/jackreid/task/internal/version.Version=v1.0.0"
var Version = "dev"

// BuildDate is the date when the binary was built
// Set via: -ldflags "-X github.com/jackreid/task/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var BuildDate = "unknown"

// GitCommit is the git commit hash
// Set via: -ldflags "-X github.com/jackreid/task/internal/version.GitCommit=$(git rev-parse --short HEAD)"
var GitCommit = "unknown"

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("task version %s (build date: %s, commit: %s)", Version, BuildDate, GitCommit)
}

// Full returns the full version information
func Full() string {
	return fmt.Sprintf("task version %s (build date: %s, commit: %s)", Version, BuildDate, GitCommit)
}
