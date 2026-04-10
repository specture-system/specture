package cmd

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
)

func versionString() string {
	return fmt.Sprintf("%s (%s)", Version, Commit)
}
