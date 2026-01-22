package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/specture-system/specture/internal/new"
	"github.com/specture-system/specture/internal/prompt"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n", "add", "a"},
	Short:   "Create a new spec",
	Long: "Create a new spec file with proper numbering and branch.\n\n" +
		"Interactive mode (default):\n" +
		"  Prompts for title, shows preview, opens editor\n\n" +
		"Non-interactive mode:\n" +
		"  specture new --title \"My Spec\"           # Provide title via flag\n" +
		"  cat body.md | specture new --title \"...\" # Pipe content (requires --title)",
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

		// Get title from flag if provided
		title, err := cmd.Flags().GetString("title")
		if err != nil {
			return fmt.Errorf("failed to get title flag: %w", err)
		}

		// Detect piped stdin
		stdinStat, _ := os.Stdin.Stat()
		piped := (stdinStat.Mode() & os.ModeCharDevice) == 0

		var pipedContent string
		if piped {
			// If stdin is piped, read the entire content
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read piped stdin: %w", err)
			}
			pipedContent = string(data)

			// When piping content, --title is required
			if title == "" {
				return fmt.Errorf("title is required when piping spec content to stdin")
			}
		}

		// If title not provided via flag or piped input, prompt the user (normal interactive mode)
		if title == "" {
			// Prompt for spec title
			title, err = prompt.PromptString("Spec title: ")
			if err != nil {
				return fmt.Errorf("failed to read spec title: %w", err)
			}
		}

		if title == "" {
			return fmt.Errorf("spec title cannot be empty")
		}

		// Create context
		ctx, err := new.NewContext(cwd, title)
		if err != nil {
			return err
		}

		// Get no-branch flag
		noBranch, err := cmd.Flags().GetBool("no-branch")
		if err != nil {
			return fmt.Errorf("failed to get no-branch flag: %w", err)
		}

		// Show what will happen
		cmd.Printf("Creating spec %03d: %s\n", ctx.Number, ctx.Title)
		if !noBranch {
			cmd.Printf("Branch: %s\n", ctx.BranchName)
		}
		cmd.Printf("File: %s\n", ctx.FileName)
		cmd.Printf("Author: %s\n", ctx.Author)

		if dryRun {
			cmd.Println("\n[dry-run] No changes made")
			return nil
		}

		// Skip confirmation if title was provided via flag
		titleFlag, _ := cmd.Flags().GetString("title")
		if titleFlag == "" {
			cmd.Println()
			ok, err := prompt.Confirm("Proceed with creating spec?")
			if err != nil {
				return fmt.Errorf("failed to get confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Spec creation cancelled.")
				return nil
			}
		}

		// Create spec file and branch (with optional piped content)
		if err := ctx.CreateSpec(dryRun, pipedContent, noBranch); err != nil {
			// Clean up if spec creation fails
			if cleanupErr := ctx.Cleanup(); cleanupErr != nil {
				cmd.Printf("Spec creation failed: %v\n", err)
				cmd.Printf("Cleanup also failed: %v\n", cleanupErr)
				return err
			}
			return err
		}

		// If there was piped content, skip the editor
		if pipedContent != "" {
			if !noBranch {
				cmd.Printf("\nSpec created in branch %s. Commit and push when ready.\n", ctx.BranchName)
			} else {
				cmd.Printf("\nSpec created at %s. Commit and push when ready.\n", ctx.FilePath)
			}
			return nil
		}

		// Open editor unless --no-editor flag is set
		noEditor, err := cmd.Flags().GetBool("no-editor")
		if err != nil {
			return fmt.Errorf("failed to get no-editor flag: %w", err)
		}

		if !noEditor {
			cmd.Printf("\nOpening %s in your editor...\n", ctx.FileName)

			// Open in editor
			if err := new.OpenEditor(ctx.FilePath); err != nil {
				// Editor exited with error, clean up
				cmd.Printf("\nCancelling spec creation...\n")
				if cleanupErr := ctx.Cleanup(); cleanupErr != nil {
					return fmt.Errorf("spec creation cancelled, but cleanup failed: %w", cleanupErr)
				}
				cmd.Println("Spec and branch removed.")
				return nil
			}
		}

		if !noBranch {
			cmd.Printf("\nSpec created in branch %s. Commit and push when ready.\n", ctx.BranchName)
		} else {
			cmd.Printf("\nSpec created at %s. Commit and push when ready.\n", ctx.FilePath)
		}
		return nil
	},
}

func init() {
	newCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
	newCmd.Flags().StringP("title", "t", "", "Spec title (skips title prompt)")
	newCmd.Flags().Bool("no-editor", false, "Skip opening editor after creating spec")
	newCmd.Flags().Bool("no-branch", false, "Skip creating git branch for spec")
}
