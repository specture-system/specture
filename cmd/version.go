package cmd

import (
	"fmt"
	"strings"
)

func SetVersion(version, commit string) {
	version = normalizeVersion(version)
	commit = normalizeCommit(commit)

	if commit == "" || commit == "unknown" {
		rootCmd.Version = version
	} else {
		rootCmd.Version = fmt.Sprintf("%s (%s)", version, commit)
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func normalizeVersion(version string) string {
	if version == "" || version == "dev" || strings.HasPrefix(version, "v") {
		return version
	}

	return "v" + version
}

func normalizeCommit(commit string) string {
	if commit == "" || commit == "unknown" || len(commit) <= 7 {
		return commit
	}

	return commit[:7]
}

func init() {
	SetVersion("dev", "unknown")
}
