package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/prompt"
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

		// Show summary of what will happen
		cmd.Printf("Detected forge: %s (%s)\n", ctx.Forge, ctx.Terminology)
		cmd.Println("\nSetup will:")
		cmd.Println("  • Create specs/ directory")
		cmd.Println("  • Create specs/README.md with Specture System guidelines")

		// Check for existing AGENTS.md and CLAUDE.md
		hasAgents, hasClaude := ctx.FindExistingFiles()
		if hasAgents {
			cmd.Println("  • Show update prompt for AGENTS.md")
		}
		if hasClaude {
			cmd.Println("  • Show update prompt for CLAUDE.md")
		}

		// Prompt for confirmation unless in dry-run mode
		if !dryRun {
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

		// Handle AGENTS.md and CLAUDE.md update prompts (skip in dry-run mode)
		if !dryRun {
			if hasAgents {
				cmd.Println()
				confirmed, err := prompt.ShowTemplate("Update AGENTS.md with Specture System information?")
				if err != nil {
					return fmt.Errorf("failed to get AGENTS.md confirmation: %w", err)
				}
				if confirmed != "" {
					cmd.Println("Copy the following into your AGENTS.md file:")
					cmd.Println()
					cmd.Println(setup.AgentPromptTemplate)
				}
			}

			if hasClaude {
				cmd.Println()
				confirmed, err := prompt.ShowTemplate("Update CLAUDE.md with Specture System information?")
				if err != nil {
					return fmt.Errorf("failed to get CLAUDE.md confirmation: %w", err)
				}
				if confirmed != "" {
					cmd.Println("Copy the following into your CLAUDE.md file:")
					cmd.Println()
					cmd.Println(setup.ClaudePromptTemplate)
				}
			}
		}

		cmd.Println("\nInitialized Specture System in this repository")
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
}
