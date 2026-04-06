package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specture-system/specture/internal/prompt"
	"github.com/specture-system/specture/internal/setup"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"update", "u"},
	Short:   "Initialize the Specture System in a repository",
	Long: `Initialize the Specture System in a repository.

Actions:
  • Create specs/ tree and specs/README.md
  • Migrate the specs tree to nested SPEC.md files`,
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

		// Show summary of what will happen
		cmd.Printf("Detected forge: %s (%s)\n", ctx.Forge, ctx.ContributionType)
		cmd.Println("\nSetup will:")
		cmd.Println("  • Create specs/ tree")
		cmd.Println("  • Create specs/README.md with Specture System guidelines")
		cmd.Println("  • Create specs/.gitignore to keep only SPEC.md and README.md")
		cmd.Println("  • Migrate existing flat specs into numbered SPEC.md directories")
		cmd.Println("  • Install Specture skill files into .agents/skills/")

		// Get yes flag
		yes, err := cmd.Flags().GetBool("yes")
		if err != nil {
			return fmt.Errorf("failed to get yes flag: %w", err)
		}

		// Prompt for confirmation unless in dry-run mode or --yes flag
		if !dryRun && !yes {
			cmd.Println()
			ok, err := prompt.Confirm("Proceed with setup?")
			if err != nil {
				return fmt.Errorf("failed to get user confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Setup cancelled.")
				return nil
			}
		}

		// Create specs directory
		if err := ctx.CreateSpecsDirectory(dryRun); err != nil {
			return err
		}

		// Create specs/README.md
		if err := ctx.CreateSpecsReadme(dryRun); err != nil {
			return err
		}

		// Migrate .skills/specture/ to .agents/skills/specture/ if needed
		migrated, err := setup.MigrateSkillsDir(cwd, dryRun)
		if err != nil {
			return err
		}
		if migrated && !dryRun {
			cmd.Println("Migrated .skills/specture/ → .agents/skills/specture/")
		}

		// Install Specture skill files
		if err := setup.InstallSkill(cwd, dryRun); err != nil {
			return err
		}

		// Migrate existing flat specs to add number to frontmatter
		specsDir := filepath.Join(cwd, "specs")
		migrations, err := setup.FindSpecsNeedingMigration(specsDir)
		if err == nil && len(migrations) > 0 {
			cmd.Printf("\nMigrating %d spec(s) to add number field:\n", len(migrations))
			for _, m := range migrations {
				cmd.Printf("  %s → number: %d\n", filepath.Base(m.Path), m.Number)
				if !dryRun {
					if err := setup.AddNumberToFrontmatter(m.Path, m.Number); err != nil {
						cmd.PrintErrf("  Error migrating %s: %v\n", filepath.Base(m.Path), err)
					}
				}
			}
			if dryRun {
				cmd.Println("  [dry-run] No changes made")
			}
		}

		specsMigrated, err := setup.MigrateSpecsLayout(specsDir, dryRun)
		if err != nil {
			return err
		}
		if specsMigrated && !dryRun {
			cmd.Println("Migrated existing flat specs into numbered SPEC.md directories")
		}

		cmd.Println("\nInitialized Specture System in this repository")
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
	setupCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
