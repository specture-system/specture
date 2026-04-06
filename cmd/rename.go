package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specture-system/specture/internal/rename"
	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename [slug]",
	Args:  cobra.ExactArgs(1),
	Short: "Rename a spec and update cross-references",
	Long: `Rename a spec and update all markdown links that reference it in the specs tree.

Examples:
  specture rename --spec 3 status-command           # Rename to status-command
  specture rename --spec 3 status-command --dry-run  # Preview changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRename(cmd, args)
	},
}

func init() {
	renameCmd.Flags().StringP("spec", "s", "", "Spec number to rename (required)")

	renameCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
	renameCmd.MarkFlagRequired("spec")
}

func runRename(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specsDir := filepath.Join(cwd, "specs")

	specArg, _ := cmd.Flags().GetString("spec")
	slug := args[0]
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Resolve the requested reference first so dotted refs are supported.
	if _, err := specpkg.ResolvePath(specsDir, specArg); err != nil {
		return err
	}

	result, err := rename.Plan(specsDir, specArg, slug)
	if err != nil {
		return err
	}

	oldRelativePath, _ := filepath.Rel(specsDir, result.OldPath)
	newRelativePath, _ := filepath.Rel(specsDir, result.NewPath)

	// Display plan
	cmd.Printf("Rename: %s → %s\n", oldRelativePath, newRelativePath)
	if len(result.LinkUpdates) > 0 {
		cmd.Printf("\nLink updates:\n")
		for _, u := range result.LinkUpdates {
			cmd.Printf("  %s: %s → %s\n", filepath.Base(u.File), u.OldLink, u.NewLink)
		}
	}

	if dryRun {
		cmd.Println("\n[dry-run] No changes made")
		return nil
	}

	if err := rename.Execute(result); err != nil {
		return err
	}

	cmd.Println("\nDone.")
	return nil
}
