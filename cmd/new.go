package cmd

import (
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new spec",
	Long: `New creates a new spec file with the proper numbering,
creates a branch for the spec, and opens the file in your editor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Coming soon: new command is under development")
		return nil
	},
}
