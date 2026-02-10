package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "specture",
	Short: "A spec-driven software architecture system",
	Long: `Specture is a spec-driven software architecture system. It provides a lightweight, document-driven approach to project planning.

Spec numbers are stored in YAML frontmatter (number field). New specs use slug-only filenames.
Use 'specture setup' to migrate existing NNN-slug.md specs to include the number field.`,
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
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(renameCmd)
}
