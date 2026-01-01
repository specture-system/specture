package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSpec_Valid(t *testing.T) {
	content := []byte(`---
status: draft
author: Test Author
---

# My Feature

This is a description.

## Task List

- [ ] Task 1
- [x] Task 2
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if !result.IsValid() {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidateSpec_MissingFrontmatter(t *testing.T) {
	content := []byte(`# My Feature

Description.

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
	}

	if result.Errors[0].Field != "frontmatter" {
		t.Errorf("expected frontmatter error, got: %v", result.Errors[0])
	}
}

func TestValidateSpec_MissingStatus(t *testing.T) {
	content := []byte(`---
author: Test Author
---

# My Feature

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	hasStatusError := false
	for _, e := range result.Errors {
		if e.Field == "status" {
			hasStatusError = true
			break
		}
	}
	if !hasStatusError {
		t.Errorf("expected status error, got: %v", result.Errors)
	}
}

func TestValidateSpec_InvalidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"empty status", ""},
		{"invalid status", "invalid"},
		{"typo in status", "draf"},
		{"uppercase status", "DRAFT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte(`---
status: ` + tt.status + `
---

# My Feature

- [ ] Task 1
`)

			spec, err := ParseSpecContent("test.md", content)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			result := ValidateSpec(spec)
			if result.IsValid() {
				t.Error("expected validation to fail")
			}

			hasStatusError := false
			for _, e := range result.Errors {
				if e.Field == "status" {
					hasStatusError = true
					break
				}
			}
			if !hasStatusError {
				t.Errorf("expected status error, got: %v", result.Errors)
			}
		})
	}
}

func TestValidateSpec_ValidStatuses(t *testing.T) {
	for _, status := range ValidStatus {
		t.Run(status, func(t *testing.T) {
			content := []byte(`---
status: ` + status + `
---

# My Feature

- [ ] Task 1
`)

			spec, err := ParseSpecContent("test.md", content)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			result := ValidateSpec(spec)
			if !result.IsValid() {
				t.Errorf("expected valid status %q, got errors: %v", status, result.Errors)
			}
		})
	}
}

func TestValidateSpec_MissingTitle(t *testing.T) {
	content := []byte(`---
status: draft
---

## Not a title (H2)

Description.

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	hasTitleError := false
	for _, e := range result.Errors {
		if e.Field == "title" {
			hasTitleError = true
			break
		}
	}
	if !hasTitleError {
		t.Errorf("expected title error, got: %v", result.Errors)
	}
}

func TestValidateSpec_MissingTaskList(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

Description without tasks.

- Regular list item
- Another item
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	hasTaskListError := false
	for _, e := range result.Errors {
		if e.Field == "task list" {
			hasTaskListError = true
			break
		}
	}
	if !hasTaskListError {
		t.Errorf("expected task list error, got: %v", result.Errors)
	}
}

func TestValidateSpec_MultipleErrors(t *testing.T) {
	// Missing everything
	content := []byte(`Just some text without structure.`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	// Should have errors for frontmatter, title, and task list
	if len(result.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateSpecFile(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "000-test.md")

	content := []byte(`---
status: approved
author: File Author
---

# File Test

Description here.

- [ ] A task
`)

	if err := os.WriteFile(specPath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ValidateSpecFile(specPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsValid() {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidateSpecFile_NotFound(t *testing.T) {
	_, err := ValidateSpecFile("/nonexistent/path/to/spec.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFormatValidationResult_Valid(t *testing.T) {
	result := &ValidationResult{
		Path:   "/path/to/000-test.md",
		Errors: []ValidationError{},
	}

	output := FormatValidationResult(result)
	if !strings.Contains(output, "\u2713") {
		t.Errorf("expected checkmark in output, got: %s", output)
	}
	if !strings.Contains(output, "000-test.md") {
		t.Errorf("expected filename in output, got: %s", output)
	}
}

func TestFormatValidationResult_Invalid(t *testing.T) {
	result := &ValidationResult{
		Path: "/path/to/000-test.md",
		Errors: []ValidationError{
			{Field: "status", Message: "missing required field"},
			{Field: "title", Message: "missing H1 heading"},
		},
	}

	output := FormatValidationResult(result)
	if !strings.Contains(output, "\u2717") {
		t.Errorf("expected X mark in output, got: %s", output)
	}
	if !strings.Contains(output, "000-test.md") {
		t.Errorf("expected filename in output, got: %s", output)
	}
	if !strings.Contains(output, "status") {
		t.Errorf("expected status error in output, got: %s", output)
	}
	if !strings.Contains(output, "title") {
		t.Errorf("expected title error in output, got: %s", output)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{Field: "status", Message: "invalid value"}
	expected := "status: invalid value"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
