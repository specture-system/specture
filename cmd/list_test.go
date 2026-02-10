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
func execList(t *testing.T, tmpDir string, flags map[string]string, boolFlags map[string]bool) (string, error) {
	t.Helper()

	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		listCmd.Flags().Set("status", "")
		listCmd.Flags().Set("format", "text")
		listCmd.Flags().Set("tasks", "false")
		listCmd.Flags().Set("incomplete", "false")
		listCmd.Flags().Set("complete", "false")
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	for k, v := range flags {
		cmd.Flags().Set(k, v)
	}
	for k, v := range boolFlags {
		if v {
			cmd.Flags().Set(k, "true")
		} else {
			cmd.Flags().Set(k, "false")
		}
	}

	err := runList(cmd, []string{})
	return out.String(), err
}

// ---- Spec fixtures ----

const listInProgressSpec = `---
status: in-progress
---

# Status Command

Description.

## Task List

### Phase One

- [x] Create SpecInfo struct
- [x] Create Task struct

### Phase Two

- [ ] Write tests
- [ ] Implement feature
`

const listCompletedSpec = `---
status: completed
---

# Setup Command

All done.

## Task List

- [x] Task A
- [x] Task B
- [x] Task C
`

const listDraftSpec = `---
status: draft
---

# Future Feature

Just an idea.

## Task List

- [ ] Think about it
- [ ] Plan it
`

const listApprovedSpec = `---
status: approved
---

# Approved Feature

Ready to go.

## Task List

- [ ] Do the thing
`

// ---- Text output tests ----

func TestListCommand_TextOutput_AllSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":    listCompletedSpec,
		"002-draft.md":    listDraftSpec,
		"003-status.md":   listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should list all three specs in ascending order
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), output)
	}

	// Check each row has expected columns
	expected := []struct {
		number   string
		status   string
		progress string
		name     string
	}{
		{"001", "completed", "3/3", "Setup Command"},
		{"002", "draft", "0/2", "Future Feature"},
		{"003", "in-progress", "2/4", "Status Command"},
	}

	for i, exp := range expected {
		if !strings.Contains(lines[i], exp.number) {
			t.Errorf("line %d: expected number %s, got: %s", i, exp.number, lines[i])
		}
		if !strings.Contains(lines[i], exp.status) {
			t.Errorf("line %d: expected status %s, got: %s", i, exp.status, lines[i])
		}
		if !strings.Contains(lines[i], exp.progress) {
			t.Errorf("line %d: expected progress %s, got: %s", i, exp.progress, lines[i])
		}
		if !strings.Contains(lines[i], exp.name) {
			t.Errorf("line %d: expected name %s, got: %s", i, exp.name, lines[i])
		}
	}
}

func TestListCommand_TextOutput_SortedByNumber(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-third.md":  listInProgressSpec,
		"001-first.md":  listCompletedSpec,
		"002-second.md": listDraftSpec,
	})

	output, err := execList(t, tmpDir, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if !strings.HasPrefix(lines[0], "001") {
		t.Errorf("first line should start with 001, got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "002") {
		t.Errorf("second line should start with 002, got: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "003") {
		t.Errorf("third line should start with 003, got: %s", lines[2])
	}
}

func TestListCommand_TextOutput_NoSpecs(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{})

	output, err := execList(t, tmpDir, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No specs found") {
		t.Errorf("expected 'No specs found', got: %s", output)
	}
}

func TestListCommand_NoSpecsDirectory(t *testing.T) {
	tmpDir := setupListTest(t, nil) // no specs dir created

	_, err := execList(t, tmpDir, nil, nil)
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

	_, err := execList(t, tmpDir, map[string]string{"format": "xml"}, nil)
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

	output, err := execList(t, tmpDir, map[string]string{"format": "json"}, nil)
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

	// Check third spec has full metadata
	spec3 := result[2]
	if spec3["current_task"] != "Write tests" {
		t.Errorf("expected current_task 'Write tests', got %v", spec3["current_task"])
	}
	if spec3["current_task_section"] != "Phase Two" {
		t.Errorf("expected current_task_section 'Phase Two', got %v", spec3["current_task_section"])
	}

	// Validate progress on spec 3
	progress, ok := spec3["progress"].(map[string]any)
	if !ok {
		t.Fatalf("progress is not an object")
	}
	if progress["complete"] != float64(2) {
		t.Errorf("expected progress.complete 2, got %v", progress["complete"])
	}
	if progress["total"] != float64(4) {
		t.Errorf("expected progress.total 4, got %v", progress["total"])
	}

	// Validate complete_tasks and incomplete_tasks are arrays
	completeTasks, ok := spec3["complete_tasks"].([]any)
	if !ok {
		t.Fatalf("complete_tasks is not an array")
	}
	if len(completeTasks) != 2 {
		t.Errorf("expected 2 complete tasks, got %d", len(completeTasks))
	}

	incompleteTasks, ok := spec3["incomplete_tasks"].([]any)
	if !ok {
		t.Fatalf("incomplete_tasks is not an array")
	}
	if len(incompleteTasks) != 2 {
		t.Errorf("expected 2 incomplete tasks, got %d", len(incompleteTasks))
	}
}

func TestListCommand_JSONOutput_EmptyList(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{})

	output, err := execList(t, tmpDir, map[string]string{"format": "json"}, nil)
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

func TestListCommand_JSONOutput_IncludesTaskSections(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"format": "json"}, nil)
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

	// Check that tasks include section info
	completeTasks := result[0]["complete_tasks"].([]any)
	firstTask := completeTasks[0].(map[string]any)
	if firstTask["section"] != "Phase One" {
		t.Errorf("expected section 'Phase One', got %v", firstTask["section"])
	}
}

// ---- Filter tests ----

func TestListCommand_FilterSingleStatus(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":  listCompletedSpec,
		"002-draft.md":  listDraftSpec,
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "completed"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[0], "Setup Command") {
		t.Errorf("expected Setup Command, got: %s", lines[0])
	}
}

func TestListCommand_FilterMultipleStatuses(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md":    listCompletedSpec,
		"002-draft.md":    listDraftSpec,
		"003-status.md":   listInProgressSpec,
		"004-approved.md": listApprovedSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "draft,in-progress"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), output)
	}
	if !strings.Contains(lines[0], "draft") {
		t.Errorf("expected draft in first line, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "in-progress") {
		t.Errorf("expected in-progress in second line, got: %s", lines[1])
	}
}

func TestListCommand_FilterNoMatches(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"001-setup.md": listCompletedSpec,
		"002-draft.md": listDraftSpec,
	})

	output, err := execList(t, tmpDir, map[string]string{"status": "rejected"}, nil)
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

	output, err := execList(t, tmpDir, map[string]string{"format": "json", "status": "in-progress"}, nil)
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

// ---- Task display tests ----

func TestListCommand_TasksFlag(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, map[string]bool{"tasks": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show both complete and incomplete tasks
	if !strings.Contains(output, "\u2713 Create SpecInfo struct") {
		t.Errorf("expected complete task marker, got:\n%s", output)
	}
	if !strings.Contains(output, "\u2022 Write tests") {
		t.Errorf("expected incomplete task marker, got:\n%s", output)
	}
}

func TestListCommand_IncompleteFlag(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, map[string]bool{"incomplete": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show incomplete tasks only
	if !strings.Contains(output, "\u2022 Write tests") {
		t.Errorf("expected incomplete task, got:\n%s", output)
	}
	// Should NOT show complete tasks
	if strings.Contains(output, "\u2713") {
		t.Errorf("should not show complete tasks with --incomplete, got:\n%s", output)
	}
}

func TestListCommand_CompleteFlag(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, map[string]bool{"complete": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show complete tasks only
	if !strings.Contains(output, "\u2713 Create SpecInfo struct") {
		t.Errorf("expected complete task, got:\n%s", output)
	}
	// Should NOT show incomplete tasks
	if strings.Contains(output, "\u2022") {
		t.Errorf("should not show incomplete tasks with --complete, got:\n%s", output)
	}
}

func TestListCommand_BothCompleteAndIncomplete(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, map[string]bool{"complete": true, "incomplete": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show both (equivalent to --tasks)
	if !strings.Contains(output, "\u2713 Create SpecInfo struct") {
		t.Errorf("expected complete task, got:\n%s", output)
	}
	if !strings.Contains(output, "\u2022 Write tests") {
		t.Errorf("expected incomplete task, got:\n%s", output)
	}
}

func TestListCommand_TasksNotShownByDefault(t *testing.T) {
	tmpDir := setupListTest(t, map[string]string{
		"003-status.md": listInProgressSpec,
	})

	output, err := execList(t, tmpDir, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT show task details
	if strings.Contains(output, "\u2713") {
		t.Errorf("should not show tasks by default, got:\n%s", output)
	}
	if strings.Contains(output, "\u2022") {
		t.Errorf("should not show tasks by default, got:\n%s", output)
	}
}
