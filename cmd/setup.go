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
	Long: `Initialize the Specture System in a repository and manage AI agent config files.

Actions:
  • Create specs/ directory and specs/README.md
  • Optionally show prompts for updating AGENTS.md and CLAUDE.md`,
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
		cmd.Println("  • Create specs/ directory")
		cmd.Println("  • Create specs/README.md with Specture System guidelines")
		cmd.Println("  • Install Specture skill files into .agents/skills/")

		// Get update flags
		updateAgents, err := cmd.Flags().GetBool("update-agents")
		if err != nil {
			return fmt.Errorf("failed to get update-agents flag: %w", err)
		}
		noUpdateAgents, err := cmd.Flags().GetBool("no-update-agents")
		if err != nil {
			return fmt.Errorf("failed to get no-update-agents flag: %w", err)
		}
		updateClaude, err := cmd.Flags().GetBool("update-claude")
		if err != nil {
			return fmt.Errorf("failed to get update-claude flag: %w", err)
		}
		noUpdateClaude, err := cmd.Flags().GetBool("no-update-claude")
		if err != nil {
			return fmt.Errorf("failed to get no-update-claude flag: %w", err)
		}

		// Check for existing AGENTS.md and CLAUDE.md
		hasAgentsFile, hasClaudeFile := ctx.FindExistingFiles()
		shouldPromptAgents := (hasAgentsFile || updateAgents) && !noUpdateAgents
		shouldPromptClaude := (hasClaudeFile || updateClaude) && !noUpdateClaude

		if shouldPromptAgents {
			cmd.Println("  • Show update prompt for AGENTS.md")
		}
		if shouldPromptClaude {
			cmd.Println("  • Show update prompt for CLAUDE.md")
		}

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

		// Migrate specs: add number to frontmatter of NNN-slug.md files
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

		// Handle AGENTS.md and CLAUDE.md update prompts (skip in dry-run mode)
		if !dryRun {
			if shouldPromptAgents {
				if err := promptForAiAgentFileUpdate(cmd, "AGENTS.md", false); err != nil {
					return err
				}
			}

			if shouldPromptClaude {
				if err := promptForAiAgentFileUpdate(cmd, "CLAUDE.md", true); err != nil {
					return err
				}
			}
		}

		cmd.Println("\nInitialized Specture System in this repository")
		return nil
	},
}

func promptForAiAgentFileUpdate(cmd *cobra.Command, filename string, isClaudeFile bool) error {
	cmd.Println()
	confirmed, err := prompt.Confirm(fmt.Sprintf("Update %s with Specture System information?", filename))
	if err != nil {
		return fmt.Errorf("failed to get %s confirmation: %w", filename, err)
	}

	if confirmed {
		cmd.Println()
		cmd.Println("Start a new session with your AI agent and prompt it with the following:")
		cmd.Println()

		renderedTemplate, err := setup.RenderAgentPromptTemplate(isClaudeFile)
		if err != nil {
			return fmt.Errorf("failed to render agent prompt template: %w", err)
		}

		cmd.Println(prompt.Yellow(renderedTemplate))
	}

	return nil
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
	setupCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	setupCmd.Flags().Bool("update-agents", false, "Show update prompt for AGENTS.md (even if file doesn't exist)")
	setupCmd.Flags().Bool("no-update-agents", false, "Skip AGENTS.md update prompt")
	setupCmd.Flags().Bool("update-claude", false, "Show update prompt for CLAUDE.md (even if file doesn't exist)")
	setupCmd.Flags().Bool("no-update-claude", false, "Skip CLAUDE.md update prompt")
}
