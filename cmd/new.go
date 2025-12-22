package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/new"
	"github.com/specture-system/specture/internal/prompt"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new spec",
	Long: `New creates a new spec file with the proper numbering,
creates a branch for the spec, and opens the file in your editor.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get dry-run flag
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return fmt.Errorf("failed to get dry-run flag: %w", err)
		}

		// Prompt for spec title
		title, err := prompt.PromptString("Spec title: ")
		if err != nil {
			return fmt.Errorf("failed to read spec title: %w", err)
		}
		if title == "" {
			return fmt.Errorf("spec title cannot be empty")
		}

		// Create context
		ctx, err := new.NewContext(cwd, title)
		if err != nil {
			return err
		}

		// Show what will happen
		cmd.Printf("Creating spec %03d: %s\n", ctx.Number, ctx.Title)
		cmd.Printf("Branch: %s\n", ctx.BranchName)
		cmd.Printf("File: %s\n", ctx.FileName)
		cmd.Printf("Author: %s\n", ctx.Author)

		if dryRun {
			cmd.Println("\n[dry-run] No changes made")
			return nil
		}

		cmd.Println()
		ok, err := prompt.Confirm("Proceed with creating spec?")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Spec creation cancelled.")
			return nil
		}

		// Create spec file and branch
		if err := ctx.CreateSpec(dryRun); err != nil {
			return err
		}

		cmd.Printf("\nOpening %s in your editor...\n", ctx.FileName)

		// Open in editor
		if err := new.OpenEditor(ctx.FilePath); err != nil {
			return err
		}

		cmd.Printf("\nSpec created in branch %s. Commit and push when ready.\n", ctx.BranchName)
		return nil
	},
}

func init() {
	newCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
}
