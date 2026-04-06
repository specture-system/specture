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
var listParentFlag string

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List specs with filtering and output options",
	Long: `List top-level specs or the children of a specific spec with optional filtering by status.

By default, shows a compact table with Ref, Name, Status, and Path for top-level specs.
Use --parent to show the immediate children of a parent spec.
Use --format json for machine-readable output with ref, name, status, and path.

Examples:
  specture list                          # List top-level specs
  specture list --parent 1.4             # List the children of spec 1.4
  specture list --status in-progress     # Filter by status
  specture list --status draft,approved  # Multiple statuses
  specture list -f json                  # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(cmd, args)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listStatusFilter, "status", "s", "", "Filter by status (comma-separated for multiple)")
	listCmd.Flags().StringVarP(&listFormatFlag, "format", "f", "text", "Output format: text or json")
	listCmd.Flags().StringVarP(&listParentFlag, "parent", "p", "", "Parent spec reference to list children for")
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

	parentRef, _ := cmd.Flags().GetString("parent")
	var parentPath string
	if strings.TrimSpace(parentRef) != "" {
		parentPath, err = specpkg.ResolvePath(specsDir, parentRef)
		if err != nil {
			return err
		}
	}

	specs, err := specpkg.FindSpecsInScope(specsDir, parentPath)
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

	// Calculate column widths from data.
	refWidth := len("REF")
	statusWidth := len("STATUS")
	nameWidth := len("NAME")
	pathWidth := len("PATH")
	for _, spec := range specs {
		if len(spec.FullRef) > refWidth {
			refWidth = len(spec.FullRef)
		}
		if len(spec.Status) > statusWidth {
			statusWidth = len(spec.Status)
		}
		if len(spec.Name) > nameWidth {
			nameWidth = len(spec.Name)
		}
		if len(spec.Path) > pathWidth {
			pathWidth = len(spec.Path)
		}
	}

	rowFmt := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds\n", refWidth, nameWidth, statusWidth, pathWidth)
	headerFmt := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds\n", refWidth, nameWidth, statusWidth, pathWidth)

	cmd.Printf(headerFmt, "REF", "NAME", "STATUS", "PATH")

	for _, spec := range specs {
		cmd.Printf(rowFmt, spec.FullRef, spec.Name, spec.Status, spec.Path)
	}

	return nil
}

// listJSONOutput represents a single spec in the JSON array output.
type listJSONOutput struct {
	Ref    string `json:"ref"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

// formatListJSON outputs specs as a JSON array with full metadata.
func formatListJSON(cmd *cobra.Command, specs []*specpkg.SpecInfo) error {
	output := make([]listJSONOutput, 0, len(specs))

	for _, spec := range specs {
		output = append(output, listJSONOutput{
			Ref:    spec.FullRef,
			Name:   spec.Name,
			Status: spec.Status,
			Path:   spec.Path,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd.Println(string(data))
	return nil
}
