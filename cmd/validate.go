package cmd

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v"},
	Short:   "Validate specs",
	Long: `Validate checks that specs follow the Specture System guidelines.

It validates frontmatter, status, descriptions, and task lists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement validate command
		return nil
	},
}
