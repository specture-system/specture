package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	implementpkg "github.com/specture-system/specture/internal/implement"
	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var implementLookPath = exec.LookPath
var implementExecutePlan = implementpkg.ExecutePlan

var implementCmd = &cobra.Command{
	Use:   "implement",
	Short: "Plan and orchestrate implementation of a spec",
	Long: `Plan and orchestrate implementation of an approved or in-progress spec.

The implement command is an agent orchestrator. It validates inputs, checks
spec eligibility, detects the agent backend, computes remaining section/task
planning, and runs the worker/review orchestration loop for each remaining
task.

Examples:
  specture implement --spec 7
  specture implement --spec 7 --agent codex`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImplement(cmd, args)
	},
}

func init() {
	implementCmd.Flags().StringP("spec", "s", "", "Spec number to implement (required)")
	implementCmd.Flags().String("agent", "", "Agent backend override: opencode or codex")
	implementCmd.Flags().Bool("dry-run", false, "Print execution plan and exit before making changes")
	implementCmd.MarkFlagRequired("spec")
}

func runImplement(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specArg, _ := cmd.Flags().GetString("spec")
	if strings.TrimSpace(specArg) == "" {
		return fmt.Errorf("spec flag is required")
	}

	agentOverride, _ := cmd.Flags().GetString("agent")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	specsDir := filepath.Join(cwd, "specs")
	specPath, err := specpkg.ResolvePath(specsDir, specArg)
	if err != nil {
		return err
	}

	info, err := specpkg.Parse(specPath)
	if err != nil {
		return err
	}

	if err := implementpkg.ValidateSpecStatus(info.Status); err != nil {
		return fmt.Errorf("spec %03d is not eligible for implement: %w", info.Number, err)
	}

	backend, err := implementpkg.SelectBackend(agentOverride, implementLookPath)
	if err != nil {
		return err
	}

	plan := implementpkg.PlanRemaining(info)

	cmd.Printf("Spec %03d: %s\n", info.Number, info.Name)
	cmd.Printf("Status: %s\n", info.Status)
	cmd.Printf("Spec Path: %s\n", info.Path)
	cmd.Printf("Agent Backend: %s\n", backend)
	cmd.Printf("Remaining Tasks: %d\n", plan.TaskCount)

	if len(plan.Sections) > 0 {
		cmd.Println()
		cmd.Println("Remaining Sections:")
		for _, section := range plan.Sections {
			name := section.Name
			if strings.TrimSpace(name) == "" {
				name = "(unsectioned)"
			}

			cmd.Printf("  • %s (%d tasks)\n", name, len(section.Tasks))
			for _, task := range section.Tasks {
				cmd.Printf("    - %s\n", task.Text)
			}
		}
	}

	if plan.TaskCount == 0 {
		cmd.Println()
		cmd.Println("No remaining tasks found.")
	}

	cmd.Println()
	cmd.Println("Planning complete.")
	if dryRun {
		cmd.Println()
		cmd.Println("Dry run enabled; exiting before implementation execution.")
		return nil
	}

	if plan.TaskCount == 0 {
		return nil
	}

	cmd.Println()
	cmd.Println("Starting implementation execution...")

	if err := implementExecutePlan(cwd, info, plan, backend, cmd.Printf); err != nil {
		return err
	}

	cmd.Println("Implementation execution complete.")

	return nil
}
