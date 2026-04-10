package cmd

import "fmt"

func SetVersion(version, commit string) {
	rootCmd.Version = fmt.Sprintf("%s (%s)", version, commit)
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func init() {
	SetVersion("dev", "unknown")
}
