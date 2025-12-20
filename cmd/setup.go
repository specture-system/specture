package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/setup"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"update"},
	Short:   "Initialize the Specture System in a repository",
	Long: `Setup initializes the Specture System in the current git repository.

It creates the specs/ directory and specs/README.md with guidelines,
and optionally updates AGENTS.md and CLAUDE.md.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Create setup context
		ctx, err := setup.NewContext(cwd)
		if err != nil {
			return err
		}

		// Get dry-run flag
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("failed to get dry-run flag: %w", err)
		}

		// Create specs directory
		if err := ctx.CreateSpecsDirectory(dryRun); err != nil {
			return err
		}

		// Create specs/README.md
		if err := ctx.CreateSpecsReadme(dryRun); err != nil {
			return err
		}

		cmd.Printf("Detected forge: %s (%s)\n", ctx.Forge, ctx.Terminology)
		cmd.Println("Initialized Specture System in this repository")
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
}
