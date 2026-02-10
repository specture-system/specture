package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Note: These tests intentionally do not use t.Parallel() because statusCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr calls).

// Helper to create a temp directory with a specs subdirectory and spec files.
func setupStatusTest(t *testing.T, specs map[string]string) string {
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

// Helper to run the status command and return the output and error.
func execStatus(t *testing.T, tmpDir string, flags map[string]string) (string, error) {
	t.Helper()

	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		statusCmd.Flags().Set("spec", "")
		statusCmd.Flags().Set("format", "text")
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := statusCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	for k, v := range flags {
		cmd.Flags().Set(k, v)
	}

	err := runStatus(cmd, []string{})
	return out.String(), err
}

const inProgressSpec = `---
status: in-progress
---

# Status Command

Description.

## Task List

### Phase One

- [x] Create SpecInfo struct
- [x] Create Task struct
- [x] Move functions

### Phase Two

- [ ] Write tests for task parsing
- [ ] Write tests for current task
`

// Spec with no explicit status that infers in-progress from mixed tasks
const inferredInProgressSpec = `# Inferred Spec

No frontmatter.

## Task List

### Setup

- [x] Step one

### Work

- [ ] Step two
`

const completedSpec = `---
status: completed
---

# Setup Command

All done.

## Task List

- [x] Task A
- [x] Task B
- [x] Task C
`

const draftSpec = `---
status: draft
---

# Future Feature

Just an idea.

## Task List

- [ ] Think about it
- [ ] Plan it
`

const noTasksSpec = `---
status: draft
---

# Empty Spec

No task list here.
`

func TestStatusCommand_TextOutput_InProgress(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"003-status.md": inProgressSpec,
	})

	output, err := execStatus(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check key elements of text output
	expected := []string{
		"Spec 003: Status Command",
		"Progress: 3/5 tasks complete",
		"Current Task Section: Phase Two",
		"Current Task: Write tests for task parsing",
		"Complete:",
		"\u2713 Create SpecInfo struct",
		"Remaining:",
		"\u2022 Write tests for current task",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got:\n%s", exp, output)
		}
	}
}

func TestStatusCommand_JSONOutput_InProgress(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"003-status.md": inProgressSpec,
	})

	output, err := execStatus(t, tmpDir, map[string]string{"format": "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, output)
	}

	// Validate JSON structure
	if result["number"] != float64(3) {
		t.Errorf("expected number 3, got %v", result["number"])
	}
	if result["name"] != "Status Command" {
		t.Errorf("expected name 'Status Command', got %v", result["name"])
	}
	if result["current_task"] != "Write tests for task parsing" {
		t.Errorf("expected current_task, got %v", result["current_task"])
	}
	if result["current_task_section"] != "Phase Two" {
		t.Errorf("expected current_task_section 'Phase Two', got %v", result["current_task_section"])
	}

	// Validate complete_tasks
	completeTasks, ok := result["complete_tasks"].([]interface{})
	if !ok {
		t.Fatalf("complete_tasks is not an array")
	}
	if len(completeTasks) != 3 {
		t.Errorf("expected 3 complete tasks, got %d", len(completeTasks))
	}

	// Validate incomplete_tasks
	incompleteTasks, ok := result["incomplete_tasks"].([]interface{})
	if !ok {
		t.Fatalf("incomplete_tasks is not an array")
	}
	if len(incompleteTasks) != 2 {
		t.Errorf("expected 2 incomplete tasks, got %d", len(incompleteTasks))
	}

	// Validate progress
	progress, ok := result["progress"].(map[string]interface{})
	if !ok {
		t.Fatalf("progress is not an object")
	}
	if progress["complete"] != float64(3) {
		t.Errorf("expected progress.complete 3, got %v", progress["complete"])
	}
	if progress["total"] != float64(5) {
		t.Errorf("expected progress.total 5, got %v", progress["total"])
	}
}

func TestStatusCommand_SpecFlag(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md":  completedSpec,
		"003-status.md": inProgressSpec,
	})

	// Use --spec to target the completed spec specifically
	output, err := execStatus(t, tmpDir, map[string]string{"spec": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Spec 001: Setup Command") {
		t.Errorf("expected output for spec 001, got:\n%s", output)
	}
	if !strings.Contains(output, "Status: completed") {
		t.Errorf("expected completed status, got:\n%s", output)
	}
}

func TestStatusCommand_NoSpecsDirectory(t *testing.T) {
	// No specs dir at all
	tmpDir := setupStatusTest(t, nil)

	_, err := execStatus(t, tmpDir, nil)
	if err == nil {
		t.Fatal("expected error for missing specs directory")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestStatusCommand_NoInProgressSpec(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md": completedSpec,
	})

	output, err := execStatus(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "No in-progress spec found") {
		t.Errorf("expected helpful message, got:\n%s", output)
	}
}

func TestStatusCommand_SpecNotFound(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md": completedSpec,
	})

	_, err := execStatus(t, tmpDir, map[string]string{"spec": "999"})
	if err == nil {
		t.Fatal("expected error for nonexistent spec")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestStatusCommand_EmptyTaskList(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-empty.md": noTasksSpec,
	})

	output, err := execStatus(t, tmpDir, map[string]string{"spec": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Spec 001: Empty Spec") {
		t.Errorf("expected spec name in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Status: draft") {
		t.Errorf("expected draft status, got:\n%s", output)
	}
	if !strings.Contains(output, "Progress: 0/0 tasks complete") {
		t.Errorf("expected 0/0 progress, got:\n%s", output)
	}
	// Should NOT contain Complete or Remaining sections
	if strings.Contains(output, "Complete:") {
		t.Errorf("should not contain Complete section for empty task list, got:\n%s", output)
	}
	if strings.Contains(output, "Remaining:") {
		t.Errorf("should not contain Remaining section for empty task list, got:\n%s", output)
	}
}

func TestStatusCommand_AllTasksComplete(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md": completedSpec,
	})

	output, err := execStatus(t, tmpDir, map[string]string{"spec": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Status: completed") {
		t.Errorf("expected completed status, got:\n%s", output)
	}
	if !strings.Contains(output, "Progress: 3/3 tasks complete") {
		t.Errorf("expected 3/3 progress, got:\n%s", output)
	}
	// Should have Complete section but no Current Task or Remaining
	if !strings.Contains(output, "Complete:") {
		t.Errorf("expected Complete section, got:\n%s", output)
	}
	if strings.Contains(output, "Current Task:") {
		t.Errorf("should not show Current Task for completed spec, got:\n%s", output)
	}
	if strings.Contains(output, "Remaining:") {
		t.Errorf("should not contain Remaining section, got:\n%s", output)
	}
}

func TestStatusCommand_InvalidFormat(t *testing.T) {
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md": completedSpec,
	})

	_, err := execStatus(t, tmpDir, map[string]string{"format": "xml"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format: xml") {
		t.Errorf("expected 'invalid format: xml' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "must be 'text' or 'json'") {
		t.Errorf("expected usage hint in error, got: %v", err)
	}
}

func TestStatusCommand_DefaultFindsInProgress(t *testing.T) {
	// With multiple specs, default should find the first in-progress one by number
	tmpDir := setupStatusTest(t, map[string]string{
		"001-setup.md":  completedSpec,
		"002-draft.md":  draftSpec,
		"003-status.md": inProgressSpec,
	})

	output, err := execStatus(t, tmpDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find spec 003 which has status: in-progress
	if !strings.Contains(output, "Spec 003: Status Command") {
		t.Errorf("expected to find in-progress spec 003, got:\n%s", output)
	}
}
