package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specture-system/specture/internal/rename"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename a spec file and update cross-references",
	Long: `Rename a spec file and update all markdown links that reference it in the specs directory.

Examples:
  specture rename --spec 3 --slug status-command  # Rename to status-command.md
  specture rename --spec 3                        # Strip numeric prefix
  specture rename --spec 3 --dry-run              # Preview changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRename(cmd, args)
	},
}

func init() {
	renameCmd.Flags().StringP("spec", "s", "", "Spec number to rename (required)")
	renameCmd.Flags().String("slug", "", "Target filename slug (default: strip numeric prefix)")
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
	slug, _ := cmd.Flags().GetString("slug")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Parse spec number
	var specNum int
	if _, err := fmt.Sscanf(specArg, "%d", &specNum); err != nil {
		return fmt.Errorf("invalid spec number: %s", specArg)
	}

	result, err := rename.Plan(specsDir, specNum, slug)
	if err != nil {
		return err
	}

	// Display plan
	cmd.Printf("Rename: %s → %s\n", filepath.Base(result.OldPath), filepath.Base(result.NewPath))
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
