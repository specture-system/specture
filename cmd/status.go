package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var statusSpecFlag string
var statusFormatFlag string

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"s"},
	Short:   "Show status of current or specified spec",
	Long: `Show the status of the current in-progress spec, or a specific spec by number.

By default, finds the first in-progress spec and displays its status.
Use --spec to target a specific spec by number.
Use --format to choose between human-readable text and JSON output.

Examples:
  specture status              # Show current in-progress spec
  specture status --spec 3     # Show status of spec 003
  specture status -f json      # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd, args)
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusSpecFlag, "spec", "s", "", "Spec number to target (e.g., 0, 00, or 000)")
	statusCmd.Flags().StringVarP(&statusFormatFlag, "format", "f", "text", "Output format: text or json")
}

// runStatus performs the status command logic.
// Separated from the command for testability.
func runStatus(cmd *cobra.Command, args []string) error {
	// Validate format flag
	format, _ := cmd.Flags().GetString("format")
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", format)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	specsDir := filepath.Join(cwd, "specs")

	// Get the target spec
	var info *specpkg.SpecInfo

	specArg, _ := cmd.Flags().GetString("spec")
	if specArg != "" {
		// Target a specific spec by number
		path, err := specpkg.ResolvePath(specsDir, specArg)
		if err != nil {
			return err
		}
		info, err = specpkg.Parse(path)
		if err != nil {
			return err
		}
	} else {
		// Find current in-progress spec
		specs, err := specpkg.ParseAll(specsDir)
		if err != nil {
			return err
		}
		info = specpkg.FindCurrent(specs)
		if info == nil {
			cmd.Println("No in-progress spec found")
			return nil
		}
	}

	// Format and display output
	if format == "json" {
		return formatJSON(cmd, info)
	}
	return formatText(cmd, info)
}

// formatText outputs the spec status in human-readable text format.
func formatText(cmd *cobra.Command, info *specpkg.SpecInfo) error {
	totalTasks := len(info.CompleteTasks) + len(info.IncompleteTasks)

	cmd.Printf("Spec %03d: %s\n", info.Number, info.Name)
	cmd.Printf("Status: %s\n", info.Status)
	cmd.Printf("Progress: %d/%d tasks complete\n", len(info.CompleteTasks), totalTasks)

	// Current task info (only if there is one)
	if info.CurrentTask != "" {
		cmd.Println()
		if info.CurrentTaskSection != "" {
			cmd.Printf("Current Task Section: %s\n", info.CurrentTaskSection)
		}
		cmd.Printf("Current Task: %s\n", info.CurrentTask)
	}

	// Complete tasks
	if len(info.CompleteTasks) > 0 {
		cmd.Println()
		cmd.Println("Complete:")
		for _, task := range info.CompleteTasks {
			cmd.Printf("  ✓ %s\n", task.Text)
		}
	}

	// Remaining tasks
	if len(info.IncompleteTasks) > 0 {
		cmd.Println()
		cmd.Println("Remaining:")
		for _, task := range info.IncompleteTasks {
			cmd.Printf("  • %s\n", task.Text)
		}
	}

	return nil
}

// jsonOutput represents the JSON structure for the status command.
type jsonOutput struct {
	Number          int          `json:"number"`
	Name            string       `json:"name"`
	Status          string       `json:"status"`
	CurrentTask     string       `json:"current_task"`
	CurrentTaskSect string       `json:"current_task_section"`
	CompleteTasks   []jsonTask   `json:"complete_tasks"`
	IncompleteTasks []jsonTask   `json:"incomplete_tasks"`
	Progress        jsonProgress `json:"progress"`
}

type jsonTask struct {
	Text    string `json:"text"`
	Section string `json:"section"`
}

type jsonProgress struct {
	Complete int `json:"complete"`
	Total    int `json:"total"`
}

// formatJSON outputs the spec status in JSON format.
func formatJSON(cmd *cobra.Command, info *specpkg.SpecInfo) error {
	completeTasks := make([]jsonTask, 0, len(info.CompleteTasks))
	for _, t := range info.CompleteTasks {
		completeTasks = append(completeTasks, jsonTask{Text: t.Text, Section: t.Section})
	}

	incompleteTasks := make([]jsonTask, 0, len(info.IncompleteTasks))
	for _, t := range info.IncompleteTasks {
		incompleteTasks = append(incompleteTasks, jsonTask{Text: t.Text, Section: t.Section})
	}

	output := jsonOutput{
		Number:          info.Number,
		Name:            info.Name,
		Status:          info.Status,
		CurrentTask:     info.CurrentTask,
		CurrentTaskSect: info.CurrentTaskSection,
		CompleteTasks:   completeTasks,
		IncompleteTasks: incompleteTasks,
		Progress: jsonProgress{
			Complete: len(info.CompleteTasks),
			Total:    len(info.CompleteTasks) + len(info.IncompleteTasks),
		},
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd.Println(string(data))
	return nil
}
