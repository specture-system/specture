package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a temp directory with a specs subdirectory and spec files.
func setupListTest(t *testing.T, specs map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	if specs != nil {
		specsDir := filepath.Join(tmpDir, "specs")
		if err := os.MkdirAll(specsDir, 0755); err != nil {
			t.Fatalf("failed to create specs dir: %v", err)
		}
		for name, content := range specs {
			if err := os.WriteFile(filepath.Join(specsDir, name), []byte(content), 0644); err != nil {
				t.Fatalf("failed to write spec %s: %v", name, err)
			}
		}
	}
	return tmpDir
}

// Helper to run the list command and return the output and error.
func execList(t *testing.T, tmpDir string, flags map[string]string) (string, error) {
	t.Helper()

	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		listCmd.Flags().Set("status", "")
		listCmd.Flags().Set("format", "text")
		listCmd.Flags().Set("parent", "")
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	for k, v := range flags {
		cmd.Flags().Set(k, v)
	}

	err := runList(cmd, []string{})
	return out.String(), err
}

// ---- Spec fixtures ----

const listInProgressSpec = `---
number: 3
status: in-progress
---

# Status Command
Description.
`

const listCompletedSpec = `---
number: 1
status: completed
---

# Setup Command

All done.
`

const listDraftSpec = `---
number: 2
status: draft
---

# Future Feature

Just an idea.
`

const listApprovedSpec = `---
number: 4
status: approved
---

# Approved Feature

Ready to go.
`

// ---- Text output tests ----

func TestListCommand_TextOutput_AllSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":  listCompletedSpec,
		"002-draft.md":  listDraftSpec,
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// 1 header + 3 data rows
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (header + 3 rows), got %d:\n%s", len(lines), output)
	}

	// Check header
	header := lines[0]
	for _, col := range []string{"NUM", "STATUS", "NAME"} {
		if !strings.Contains(header, col) {
			t.Errorf("header missing %q: %s", col, header)
		}
	}

	// Check each data row has expected columns
	expected := []struct {
		number string
		status string
		name   string
	}{
		{"001", "completed", "Setup Command"},
		{"002", "draft", "Future Feature"},
		{"003", "in-progress", "Status Command"},
	}

	for i, exp := range expected {
		row := lines[i+1] // skip header
		if !strings.Contains(row, exp.number) {
			t.Errorf("row %d: expected number %s, got: %s", i, exp.number, row)
		}
		if !strings.Contains(row, exp.status) {
			t.Errorf("row %d: expected status %s, got: %s", i, exp.status, row)
		}
		if !strings.Contains(row, exp.name) {
			t.Errorf("row %d: expected name %s, got: %s", i, exp.name, row)
		}
	}
}

func TestListCommand_TextOutput_SortedByNumber(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-third.md":  listInProgressSpec,
		"001-first.md":  listCompletedSpec,
		"002-second.md": listDraftSpec,
	})

	output, err := execList(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (header + 3 rows), got %d", len(lines))
	}

	// Skip header (lines[0]), check data rows
	if !strings.HasPrefix(lines[1], "001") {
		t.Errorf("first data row should start with 001, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "002") {
		t.Errorf("second data row should start with 002, got: %s", lines[2])
	}
	if !strings.HasPrefix(lines[3], "003") {
		t.Errorf("third data row should start with 003, got: %s", lines[3])
	}
}

func TestListCommand_TextOutput_NoSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{})

	output, err := execList(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No specs found") {
		t.Errorf("expected 'No specs found', got: %s", output)
	}
}

func TestListCommand_TextOutput_IgnoresNestedSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md": listCompletedSpec,
	})

	parentDir := filepath.Join(tmpDir, "specs", "1-root")
	nestedDir := filepath.Join(parentDir, "2-child")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	parentSpec := `---
number: 1
status: approved
---

# Root
`
	nestedSpec := `---
number: 2
status: draft
---

# Nested
`
	if err := os.WriteFile(filepath.Join(parentDir, "SPEC.md"), []byte(parentSpec), 0o644); err != nil {
		t.Fatalf("failed to write parent spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "SPEC.md"), []byte(nestedSpec), 0o644); err != nil {
		t.Fatalf("failed to write nested spec: %v", err)
	}

	output, err := execList(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Setup Command") {
		t.Fatalf("expected top-level spec in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Root") {
		t.Fatalf("expected top-level nested-dir spec in output, got:\n%s", output)
	}
	if strings.Contains(output, "Nested") {
		t.Fatalf("did not expect nested spec in top-level list output, got:\n%s", output)
	}
}

func TestListCommand_TextOutput_ParentScope(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md": listCompletedSpec,
	})

	parentDir := filepath.Join(tmpDir, "specs", "0-parent")
	childDir := filepath.Join(parentDir, "000-child")
	grandchildDir := filepath.Join(childDir, "000-grandchild")
	for _, dir := range []string{parentDir, childDir, grandchildDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	parentSpec := `---
number: 0
status: approved
---

# Parent
`
	childSpec := `---
number: 0
status: draft
---

# Child
`
	grandchildSpec := `---
number: 0
status: draft
---

# Grandchild
`
	if err := os.WriteFile(filepath.Join(parentDir, "SPEC.md"), []byte(parentSpec), 0o644); err != nil {
		t.Fatalf("failed to write parent spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(childDir, "SPEC.md"), []byte(childSpec), 0o644); err != nil {
		t.Fatalf("failed to write child spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(grandchildDir, "SPEC.md"), []byte(grandchildSpec), 0o644); err != nil {
		t.Fatalf("failed to write grandchild spec: %v", err)
	}

	output, err := execList(t, tmpDir, map[string]string{"parent": "0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 child), got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[1], "Child") {
		t.Fatalf("expected direct child spec in output, got:\n%s", output)
	}
	if strings.Contains(output, "Grandchild") {
		t.Fatalf("did not expect grandchild in parent-scoped output, got:\n%s", output)
	}
}

func TestListCommand_NoSpecsDirectory(t *testing.T) {
	tmpDir := setupListTest(t, nil) // no specs dir created

	_, err := execList(t, tmpDir, nil)
	if err == nil {
		t.Fatal("expected error for missing specs directory")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestListCommand_InvalidFormat(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md": listCompletedSpec,
	})

	_, err := execList(t, tmpDir, map[string]string{"format": "xml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format: xml") {
		t.Errorf("expected 'invalid format: xml' in error, got: %v", err)
	}
}

// ---- JSON output tests ----

func TestListCommand_JSONOutput_AllSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":  listCompletedSpec,
		"002-draft.md":  listDraftSpec,
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"format": "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, output)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 specs in JSON, got %d", len(result))
	}

	// Check first spec (001)
	if result[0]["number"] != float64(1) {
		t.Errorf("expected first spec number 1, got %v", result[0]["number"])
	}
	if result[0]["name"] != "Setup Command" {
		t.Errorf("expected name 'Setup Command', got %v", result[0]["name"])
	}
	if result[0]["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", result[0]["status"])
	}

	// Check second spec (002)
	if result[1]["status"] != "draft" {
		t.Errorf("expected status 'draft', got %v", result[1]["status"])
	}

	// Check third spec has the expected metadata fields
	spec3 := result[2]
	if spec3["full_ref"] != "3" {
		t.Errorf("expected full_ref '3', got %v", spec3["full_ref"])
	}
	if path, ok := spec3["path"].(string); !ok || !strings.HasSuffix(path, filepath.Join("specs", "003-status.md")) {
		t.Errorf("expected path to end with %q, got %v", filepath.Join("specs", "003-status.md"), spec3["path"])
	}
}

func TestListCommand_JSONOutput_EmptyList(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{})

	output, err := execList(t, tmpDir, map[string]string{"format": "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

// ---- Filter tests ----

func TestListCommand_FilterSingleStatus(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":  listCompletedSpec,
		"002-draft.md":  listDraftSpec,
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "completed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// 1 header + 1 data row
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[1], "Setup Command") {
		t.Errorf("expected Setup Command, got: %s", lines[1])
	}
}

func TestListCommand_FilterMultipleStatuses(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":    listCompletedSpec,
		"002-draft.md":    listDraftSpec,
		"003-status.md":   listInProgressSpec,
		"004-approved.md": listApprovedSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "draft,in-progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	// 1 header + 2 data rows
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[1], "draft") {
		t.Errorf("expected draft in first data row, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "in-progress") {
		t.Errorf("expected in-progress in second data row, got: %s", lines[2])
	}
}

func TestListCommand_FilterNoMatches(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md": listCompletedSpec,
		"002-draft.md": listDraftSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "rejected"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No specs found") {
		t.Errorf("expected 'No specs found', got: %s", output)
	}
}

func TestListCommand_FilterJSON(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":  listCompletedSpec,
		"002-draft.md":  listDraftSpec,
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"format": "json", "status": "in-progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(result))
	}
	if result[0]["status"] != "in-progress" {
		t.Errorf("expected status 'in-progress', got %v", result[0]["status"])
	}
}
