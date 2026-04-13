package cmd

import "fmt"

func SetVersion(version, commit string) {
	if commit == "" || commit == "unknown" {
		rootCmd.Version = version
	} else {
		rootCmd.Version = fmt.Sprintf("%s (%s)", version, commit)
	}
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func init() {
	SetVersion("dev", "unknown")
}
