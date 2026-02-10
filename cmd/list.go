package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var listStatusFilter string
var listFormatFlag string
var listTasksFlag bool
var listIncompleteFlag bool
var listCompleteFlag bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List specs with filtering and output options",
	Long: `List all specs with optional filtering by status and task display.

By default, shows a compact table with Number, Status, Progress, and Name.
Use --format json for machine-readable output with full metadata.

Examples:
  specture list                          # List all specs
  specture list --status in-progress     # Filter by status
  specture list --status draft,approved  # Multiple statuses
  specture list --tasks                  # Show all tasks
  specture list --incomplete             # Show only incomplete tasks
  specture list --complete               # Show only complete tasks
  specture list -f json                  # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(cmd, args)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listStatusFilter, "status", "s", "", "Filter by status (comma-separated for multiple)")
	listCmd.Flags().StringVarP(&listFormatFlag, "format", "f", "text", "Output format: text or json")
	listCmd.Flags().BoolVar(&listTasksFlag, "tasks", false, "Show all tasks (complete and incomplete)")
	listCmd.Flags().BoolVar(&listIncompleteFlag, "incomplete", false, "Show only incomplete tasks")
	listCmd.Flags().BoolVar(&listCompleteFlag, "complete", false, "Show only complete tasks")
}

func runList(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specsDir := filepath.Join(cwd, "specs")

	specs, err := specpkg.ParseAll(specsDir)
	if err != nil {
		return err
	}

	// Apply status filter
	statusFilter, _ := cmd.Flags().GetString("status")
	if statusFilter != "" {
		specs = filterByStatus(specs, statusFilter)
	}

	if format == "json" {
		return formatListJSON(cmd, specs)
	}
	return formatListText(cmd, specs)
}

// filterByStatus filters specs by one or more comma-separated status values.
func filterByStatus(specs []*specpkg.SpecInfo, filter string) []*specpkg.SpecInfo {
	statuses := make(map[string]bool)
	for _, s := range strings.Split(filter, ",") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			statuses[trimmed] = true
		}
	}

	var filtered []*specpkg.SpecInfo
	for _, spec := range specs {
		if statuses[spec.Status] {
			filtered = append(filtered, spec)
		}
	}
	return filtered
}

// formatListText outputs specs as a human-readable table with aligned columns.
func formatListText(cmd *cobra.Command, specs []*specpkg.SpecInfo) error {
	if len(specs) == 0 {
		cmd.Println("No specs found")
		return nil
	}

	showTasks, _ := cmd.Flags().GetBool("tasks")
	showIncomplete, _ := cmd.Flags().GetBool("incomplete")
	showComplete, _ := cmd.Flags().GetBool("complete")

	// --incomplete or --complete automatically enables task display
	showTaskDetails := showTasks || showIncomplete || showComplete
	// If both --complete and --incomplete, show all (equivalent to --tasks)
	if showComplete && showIncomplete {
		showComplete = true
		showIncomplete = true
	}
	// If just --tasks, show all
	if showTasks {
		showComplete = true
		showIncomplete = true
	}

	// Calculate column widths from data
	statusWidth := len("STATUS")
	progressWidth := len("PROGRESS")
	for _, spec := range specs {
		if len(spec.Status) > statusWidth {
			statusWidth = len(spec.Status)
		}
		total := len(spec.CompleteTasks) + len(spec.IncompleteTasks)
		p := fmt.Sprintf("%d/%d", len(spec.CompleteTasks), total)
		if len(p) > progressWidth {
			progressWidth = len(p)
		}
	}

	rowFmt := fmt.Sprintf("%%03d  %%-%ds  %%%ds  %%s\n", statusWidth, progressWidth)
	headerFmt := fmt.Sprintf("%%s  %%-%ds  %%%ds  %%s\n", statusWidth, progressWidth)
	indent := "     "

	cmd.Printf(headerFmt, "NUM", "STATUS", "PROGRESS", "NAME")

	for i, spec := range specs {
		totalTasks := len(spec.CompleteTasks) + len(spec.IncompleteTasks)
		progress := fmt.Sprintf("%d/%d", len(spec.CompleteTasks), totalTasks)
		cmd.Printf(rowFmt, spec.Number, spec.Status, progress, spec.Name)

		if showTaskDetails {
			if showComplete {
				for _, task := range spec.CompleteTasks {
					cmd.Printf("%s✓ %s\n", indent, task.Text)
				}
			}
			if showIncomplete {
				for _, task := range spec.IncompleteTasks {
					cmd.Printf("%s• %s\n", indent, task.Text)
				}
			}
			// Add blank line between specs when showing tasks (except after last)
			if i < len(specs)-1 {
				cmd.Println()
			}
		}
	}

	return nil
}

// listJSONOutput represents a single spec in the JSON array output.
type listJSONOutput struct {
	Number          int          `json:"number"`
	Name            string       `json:"name"`
	Status          string       `json:"status"`
	CurrentTask     string       `json:"current_task"`
	CurrentTaskSect string       `json:"current_task_section"`
	CompleteTasks   []jsonTask   `json:"complete_tasks"`
	IncompleteTasks []jsonTask   `json:"incomplete_tasks"`
	Progress        jsonProgress `json:"progress"`
}

// formatListJSON outputs specs as a JSON array with full metadata.
func formatListJSON(cmd *cobra.Command, specs []*specpkg.SpecInfo) error {
	output := make([]listJSONOutput, 0, len(specs))

	for _, spec := range specs {
		completeTasks := make([]jsonTask, 0, len(spec.CompleteTasks))
		for _, t := range spec.CompleteTasks {
			completeTasks = append(completeTasks, jsonTask{Text: t.Text, Section: t.Section})
		}

		incompleteTasks := make([]jsonTask, 0, len(spec.IncompleteTasks))
		for _, t := range spec.IncompleteTasks {
			incompleteTasks = append(incompleteTasks, jsonTask{Text: t.Text, Section: t.Section})
		}

		output = append(output, listJSONOutput{
			Number:          spec.Number,
			Name:            spec.Name,
			Status:          spec.Status,
			CurrentTask:     spec.CurrentTask,
			CurrentTaskSect: spec.CurrentTaskSection,
			CompleteTasks:   completeTasks,
			IncompleteTasks: incompleteTasks,
			Progress: jsonProgress{
				Complete: len(spec.CompleteTasks),
				Total:    len(spec.CompleteTasks) + len(spec.IncompleteTasks),
			},
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd.Println(string(data))
	return nil
}
