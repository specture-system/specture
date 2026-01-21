package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/specture-system/specture/internal/new"
	"github.com/specture-system/specture/internal/prompt"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new spec",
	Long: `New creates a new spec file with the proper numbering,
creates a branch for the spec, and opens the file in your editor.

Non-interactive usage:
  - Provide the spec title via --title (or -t) to skip the title prompt and confirmation.
  - Pipe full spec content to stdin to create a spec programmatically. Example:

      cat content.md | specture new --title "My Spec"

    - When piping a full spec body, you must provide --title.
    - Piping content implies --no-editor (the editor will not be opened).
  - If stdin contains a single non-empty line, it will be treated as the title (interactive title input).

Examples:
  - specture new --title "My Spec"  (non-interactive title)
  - cat spec-body.md | specture new --title "My Spec"  (create from piped body)

Note: Use --dry-run to preview what will be created without modifying files.`,
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

			// Heuristic: determine if piped content looks like a full spec body
			trimmed := strings.TrimSpace(pipedContent)
			lines := strings.Split(trimmed, "\n")
			isBody := false
			// Consider it a body if it has more than one non-empty line, or looks like markdown/frontmatter
			nonEmptyLines := 0
			for _, l := range lines {
				if strings.TrimSpace(l) != "" {
					nonEmptyLines++
				}
			}
			if nonEmptyLines > 1 || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") {
				isBody = true
			}

			if isBody {
				// If content looks like a body, require --title to be present
				if title == "" {
					return fmt.Errorf("title is required when piping spec content to stdin")
				}
			} else {
				// Treat single-line piped input as a potential title.
				if title == "" {
					if trimmed == "" {
						// When piped input is empty and no title flag provided, fail early
						return fmt.Errorf("spec title cannot be empty")
					}
					title = trimmed
				}
				// Clear pipedContent so it isn't treated as a body later
				pipedContent = ""
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

		// Show what will happen
		cmd.Printf("Creating spec %03d: %s\n", ctx.Number, ctx.Title)
		cmd.Printf("Branch: %s\n", ctx.BranchName)
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

		// If there was piped content, create spec using that body and imply no-editor
		if pipedContent != "" {
			// Create spec file and branch using provided body
			if err := ctx.CreateSpecWithBody(dryRun, pipedContent); err != nil {
				// Clean up if spec creation fails
				if cleanupErr := ctx.Cleanup(); cleanupErr != nil {
					cmd.Printf("Spec creation failed: %v\n", err)
					cmd.Printf("Cleanup also failed: %v\n", cleanupErr)
					return err
				}
				return err
			}

			cmd.Printf("\nSpec created in branch %s. Commit and push when ready.\n", ctx.BranchName)
			return nil
		}

		// Create spec file and branch (no piped content)
		if err := ctx.CreateSpec(dryRun); err != nil {
			// Clean up if spec creation fails (defensive, since CreateSpec does internal cleanup)
			if cleanupErr := ctx.Cleanup(); cleanupErr != nil {
				cmd.Printf("Spec creation failed: %v\n", err)
				cmd.Printf("Cleanup also failed: %v\n", cleanupErr)
				return err
			}
			return err
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

		cmd.Printf("\nSpec created in branch %s. Commit and push when ready.\n", ctx.BranchName)
		return nil
	},
}

func init() {
	newCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
	newCmd.Flags().StringP("title", "t", "", "Spec title (skips title prompt)")
	newCmd.Flags().Bool("no-editor", false, "Skip opening editor after creating spec")
}
