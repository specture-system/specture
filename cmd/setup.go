package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/prompt"
	"github.com/specture-system/specture/internal/setup"
	"github.com/spf13/cobra"
)

// promptAndShowTemplate prompts user to update a file and displays the template if approved.
func promptAndShowTemplate(cmd *cobra.Command, filename, promptText string, template string) error {
	ok, err := prompt.Confirm(promptText)
	if err != nil {
		return fmt.Errorf("failed to get %s confirmation: %w", filename, err)
	}
	if ok {
		cmd.Println("Copy the following into your " + filename + " file:")
		cmd.Println()
		cmd.Println(template)
	}
	return nil
}

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

		// Check for existing spec files (protection against accidental overwrites)
		if err := ctx.CheckExistingSpecFiles(); err != nil {
			return err
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
				if err := promptAndShowTemplate(cmd, "AGENTS.md",
					"Update AGENTS.md with Specture System information?",
					setup.AgentPromptTemplate); err != nil {
					return err
				}
			}

			if hasClaude {
				cmd.Println()
				if err := promptAndShowTemplate(cmd, "CLAUDE.md",
					"Update CLAUDE.md with Specture System information?",
					setup.ClaudePromptTemplate); err != nil {
					return err
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
