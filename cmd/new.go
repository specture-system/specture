package cmd

import (
	"fmt"
	"os"

	"github.com/specture-system/specture/internal/new"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n", "add", "a"},
	Short:   "Create a new spec or plan file",
	Long: `Create a new SPEC.md or PLAN.md file with deterministic numbering.

Examples:
  specture new --title "My Spec"
  specture new --title "My Spec" --parent 1.4
  specture new --title "My Plan" --plan
  specture new --title "Issue Spec" --spec 123
  specture new --title "Child Spec" --spec 123.4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		title, err := cmd.Flags().GetString("title")
		if err != nil {
			return fmt.Errorf("failed to get title flag: %w", err)
		}
		if title == "" {
			return fmt.Errorf("--title is required")
		}

		parentRef, err := cmd.Flags().GetString("parent")
		if err != nil {
			return fmt.Errorf("failed to get parent flag: %w", err)
		}
		specRef, err := cmd.Flags().GetString("spec")
		if err != nil {
			return fmt.Errorf("failed to get spec flag: %w", err)
		}
		plan, err := cmd.Flags().GetBool("plan")
		if err != nil {
			return fmt.Errorf("failed to get plan flag: %w", err)
		}

		ctx, err := new.NewContext(cwd, new.Options{
			Title:     title,
			ParentRef: parentRef,
			SpecRef:   specRef,
			Plan:      plan,
		})
		if err != nil {
			return err
		}

		cmd.Printf("Creating %s %s: %s\n", ctx.Kind, ctx.FullRef, ctx.Title)
		cmd.Printf("File: %s\n", ctx.RelativePath)
		cmd.Printf("Author: %s\n", ctx.Author)

		if err := ctx.CreateFile(); err != nil {
			return err
		}
		cmd.Printf("\nCreated %s.\n", ctx.FilePath)
		return nil
	},
}

func init() {
	newCmd.Flags().StringP("title", "t", "", "Spec or plan title")
	newCmd.Flags().String("parent", "", "Parent spec reference for a child spec (e.g., 1.4)")
	newCmd.Flags().StringP("spec", "s", "", "Explicit spec reference to create (e.g., 123 or 123.4)")
	newCmd.Flags().Bool("plan", false, "Create PLAN.md instead of SPEC.md")
}
