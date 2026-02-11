package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSpec_Valid(t *testing.T) {
	content := []byte(`---
number: 0
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

## Task List

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

## Task List

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

## Task List

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
number: 0
status: ` + status + `
---

# My Feature

## Task List

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

## Task List

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

Description without Task List heading.

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
number: 0
status: approved
author: File Author
---

# File Test

Description here.

## Task List

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

// ---------- Number validation tests ----------

func TestValidateSpec_MissingNumber(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail for missing number")
	}

	hasNumberError := false
	for _, e := range result.Errors {
		if e.Field == "number" {
			hasNumberError = true
			break
		}
	}
	if !hasNumberError {
		t.Errorf("expected number error, got: %v", result.Errors)
	}
}

func TestValidateSpec_ValidNumber(t *testing.T) {
	tests := []struct {
		name   string
		number string
	}{
		{"zero", "0"},
		{"positive", "5"},
		{"large", "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte(`---
number: ` + tt.number + `
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

			spec, err := ParseSpecContent("test.md", content)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			result := ValidateSpec(spec)
			hasNumberError := false
			for _, e := range result.Errors {
				if e.Field == "number" {
					hasNumberError = true
					break
				}
			}
			if hasNumberError {
				t.Errorf("expected no number error for %s, got: %v", tt.number, result.Errors)
			}
		})
	}
}

func TestValidateSpec_NegativeNumber(t *testing.T) {
	content := []byte(`---
number: -1
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail for negative number")
	}

	hasNumberError := false
	for _, e := range result.Errors {
		if e.Field == "number" {
			hasNumberError = true
			break
		}
	}
	if !hasNumberError {
		t.Errorf("expected number error, got: %v", result.Errors)
	}
}

func TestValidateSpecs_DuplicateNumbers(t *testing.T) {
	content1 := []byte(`---
number: 3
status: draft
---

# Feature A

## Task List

- [ ] Task 1
`)
	content2 := []byte(`---
number: 3
status: draft
---

# Feature B

## Task List

- [ ] Task 1
`)

	spec1, err := ParseSpecContent("feature-a.md", content1)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	spec2, err := ParseSpecContent("feature-b.md", content2)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	results := ValidateSpecs([]*Spec{spec1, spec2})

	// At least one spec should have a duplicate number error
	foundDuplicateError := false
	for _, result := range results {
		for _, e := range result.Errors {
			if e.Field == "number" && strings.Contains(e.Message, "duplicate") {
				foundDuplicateError = true
				break
			}
		}
	}
	if !foundDuplicateError {
		t.Error("expected duplicate number error")
	}
}

func TestValidateSpecs_NoDuplicates(t *testing.T) {
	content1 := []byte(`---
number: 1
status: draft
---

# Feature A

## Task List

- [ ] Task 1
`)
	content2 := []byte(`---
number: 2
status: draft
---

# Feature B

## Task List

- [ ] Task 1
`)

	spec1, err := ParseSpecContent("feature-a.md", content1)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	spec2, err := ParseSpecContent("feature-b.md", content2)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	results := ValidateSpecs([]*Spec{spec1, spec2})

	for _, result := range results {
		for _, e := range result.Errors {
			if e.Field == "number" && strings.Contains(e.Message, "duplicate") {
				t.Errorf("unexpected duplicate error: %v", e)
			}
		}
	}
}

func TestValidateSpec_NumberFilenameMismatch(t *testing.T) {
	content := []byte(`---
number: 5
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	// Filename says 003 but frontmatter says 5
	spec, err := ParseSpecContent("003-my-feature.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)

	hasWarning := false
	for _, e := range result.Warnings {
		if e.Field == "number" && strings.Contains(e.Message, "mismatch") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("expected number/filename mismatch warning, got warnings: %v, errors: %v", result.Warnings, result.Errors)
	}
}

func TestValidateSpec_NumberFilenameMatch(t *testing.T) {
	content := []byte(`---
number: 3
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	// Filename and frontmatter agree
	spec, err := ParseSpecContent("003-my-feature.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)

	for _, e := range result.Warnings {
		if e.Field == "number" {
			t.Errorf("unexpected number warning: %v", e)
		}
	}
}

func TestValidateSpec_SlugOnlyFilenameNoMismatchWarning(t *testing.T) {
	content := []byte(`---
number: 5
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	// Slug-only filename — no mismatch possible
	spec, err := ParseSpecContent("my-feature.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)

	for _, e := range result.Warnings {
		if e.Field == "number" {
			t.Errorf("unexpected number warning for slug-only filename: %v", e)
		}
	}
}
