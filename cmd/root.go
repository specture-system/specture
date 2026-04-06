package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "specture",
	Short: "A spec-driven software architecture system",
	Long: `Specture is a spec-driven software architecture system. It provides a lightweight, document-driven approach to project planning.

Spec numbers are stored in YAML frontmatter (number field). Specs live in directories with SPEC.md files and may nest to any number of levels.
Use 'specture setup' to migrate older flat specs into the directory-based layout.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error and usage, so we just exit
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(implementCmd)
}
