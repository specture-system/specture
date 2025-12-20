package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/git"
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

		// Check if this is a git repository
		if err := git.IsGitRepository(cwd); err != nil {
			return fmt.Errorf("not a git repository")
		}

		// Check for uncommitted changes
		hasChanges, err := git.HasUncommittedChanges(cwd)
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if hasChanges {
			return fmt.Errorf("working tree has uncommitted changes")
		}

		// Detect forge
		remoteURL, err := git.GetRemoteURL(cwd, "origin")
		if err != nil {
			// No remote configured, prompt user
			cmd.Println("No 'origin' remote configured. Using default (pull request) terminology.")
		} else if remoteURL == "" {
			// No remotes at all
			cmd.Println("No git remotes configured. Using default (pull request) terminology.")
		}

		var forge git.Forge
		if remoteURL != "" {
			forge, err = git.IdentifyForge(remoteURL)
			if err != nil {
				cmd.Printf("Warning: Could not identify forge from remote URL: %v\n", err)
			}
		}

		terminology := git.GetTerminology(forge)
		cmd.Printf("Detected forge: %s (%s)\n", forge, terminology)

		cmd.Println("Setup would initialize the Specture System here")
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
}
