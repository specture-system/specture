package cmd

import (
	"fmt"
	"strings"
)

func SetVersion(version, commit string) {
	version = normalizeVersion(version)
	commit = normalizeCommit(commit)

	if commit == "" {
		rootCmd.Version = version
	} else {
		rootCmd.Version = fmt.Sprintf("%s (%s)", version, commit)
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

// normalizeVersion keeps release output stable by restoring the canonical
// v-prefixed semver format when build metadata strips it.
func normalizeVersion(version string) string {
	if version == "" || version == "dev" || strings.HasPrefix(version, "v") {
		return version
	}

	return "v" + version
}

// normalizeCommit shortens long git SHAs so --version stays compact.
func normalizeCommit(commit string) string {
	if commit == "" || len(commit) <= 7 {
		return commit
	}

	return commit[:7]
}

func init() {
	SetVersion("dev", "")
}
