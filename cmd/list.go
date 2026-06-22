package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var listStatusFilter string
var listFormatFlag string
var listParentFlag string
var listDepthFlag string

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List specs with filtering and output options",
	Long: `List specs with optional filtering, parent scoping, and depth control.

By default, shows a compact table with Ref, Name, Status, and Path for top-level specs.
Use --parent to scope to the children of a specific parent spec.
Use --depth to control how deep to recurse into the spec hierarchy.
Use --format json for machine-readable output with ref, name, status, and path.

Examples:
  specture list                          # List top-level specs
  specture list --parent 1.4             # List specs under spec 1.4 (unlimited depth)
  specture list --parent 1.4 --depth 1   # List immediate children of spec 1.4
  specture list --depth 2                # List top-level and immediate children
  specture list --depth all              # List all specs recursively
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
	listCmd.Flags().StringVarP(&listDepthFlag, "depth", "d", "1", "Recursion depth (1 = immediate scope, 0 or all = unlimited)")
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

	depth, err := parseDepth(cmd, parentPath != "")
	if err != nil {
		return err
	}

	specs, err := specpkg.FindSpecsInScopeDepth(specsDir, parentPath, depth)
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

// parseDepth converts the --depth flag string to an int.
// When --parent is set and --depth was not explicitly given, defaults to all.
// "all" and "0" mean unlimited.
func parseDepth(cmd *cobra.Command, hasParent bool) (int, error) {
	raw := listDepthFlag

	// When --parent narrows the output significantly, --depth defaults to
	// "all" so users see the full subtree without an extra flag. Without
	// --parent, --depth defaults to 1 (top-level only).  Changed() detects
	// whether the user supplied --depth; if not and --parent is set, we
	// switch the effective default.
	if hasParent && !cmd.Flags().Changed("depth") {
		raw = "all"
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "all", "0":
		return 0, nil
	default:
		d, err := strconv.Atoi(raw)
		if err != nil || d < 0 {
			return 0, fmt.Errorf("invalid depth: %s (must be a positive integer or 'all')", raw)
		}
		return d, nil
	}
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
