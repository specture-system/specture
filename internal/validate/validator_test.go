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

### Phase 1

- [ ] Task 1
- [x] Task 2
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

### Phase 1

- [ ] Task 1
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

### Phase 1

- [ ] Task 1
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

### Phase 1

- [ ] Task 1
`)

			spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

## Task List

### Phase 1

- [ ] Task 1
`)

			spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
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

func TestValidateSpec_MalformedPath(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

## Task List

- [ ] Task 1
`)

	spec, err := ParseSpecContent("specs/foo/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Fatal("expected validation to fail")
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "path" && strings.Contains(e.Message, "spec path must encode a numbered ref") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected malformed path error, got: %v", result.Errors)
	}
}

func TestValidateSpec_MissingTaskListAllowed(t *testing.T) {
	content := []byte(`---
number: 0
status: draft
---

# My Feature

Description without Task List heading.

- Regular list item
- Another item
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if !result.IsValid() {
		t.Errorf("expected missing Task List to be allowed, got errors: %v", result.Errors)
	}
}

func TestValidateSpec_TopLevelTaskCheckboxesAllowedWithoutSection(t *testing.T) {
	content := []byte(`---
number: 7
status: draft
---

# My Feature

## Task List

- [ ] Top-level task without section

### Proper Section

- [ ] Properly sectioned task
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if !result.IsValid() {
		t.Fatalf("expected top-level checkboxes without sections to be allowed, got errors: %v", result.Errors)
	}
}

func TestValidateSpec_AllTopLevelTaskCheckboxesSectioned(t *testing.T) {
	content := []byte(`---
number: 7
status: draft
---

# My Feature

## Task List

### Task Structure and Validation

- [ ] Parent task
  - [ ] Nested checkbox

### CLI Polish

- [ ] Another parent task
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	for _, e := range result.Errors {
		if e.Field == "task list" && strings.Contains(e.Message, "must be organized into '###' sections") {
			t.Fatalf("unexpected sectioning error: %v", e)
		}
	}
}

func TestValidateSpec_NumberedSectionHeadersAreInvalid(t *testing.T) {
	content := []byte(`---
number: 7
status: draft
---

# My Feature

## 1. Overview

Description.

## Task List

### Foundation

- [ ] Implement parser change
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Fatal("expected validation to fail")
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "headings" && strings.Contains(e.Message, "must not be numbered") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected numbered heading validation error, got: %v", result.Errors)
	}
}

func TestValidateSpec_UnnumberedSectionHeadersAreValid(t *testing.T) {
	content := []byte(`---
number: 7
status: draft
---

# My Feature

## Overview

Description.

## Task List

### Foundation

- [ ] Implement parser change
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	for _, e := range result.Errors {
		if e.Field == "headings" && strings.Contains(e.Message, "must not be numbered") {
			t.Fatalf("unexpected numbered heading error: %v", e)
		}
	}
}

func TestValidateSpec_GenericSpecLinkLabelIsAllowed(t *testing.T) {
	content := []byte(`---
number: 8
status: draft
---

# My Feature

See [spec 12](status-command.md) for background.

## Task List

### Foundation

- [ ] Implement parser change
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if !result.IsValid() {
		t.Fatalf("expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestValidateSpec_SpecHashLinkLabelIsAllowed(t *testing.T) {
	content := []byte(`---
number: 8
status: draft
---

# My Feature

See [spec #12](status-command.md) for background.

## Task List

### Foundation

- [ ] Implement parser change
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if !result.IsValid() {
		t.Fatalf("expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestValidateSpec_SpecTitleLinkLabelIsValid(t *testing.T) {
	content := []byte(`---
number: 8
status: draft
---

# My Feature

See [Status command](status-command.md) for background.

## Task List

### Foundation

- [ ] Implement parser change
`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	for _, e := range result.Errors {
		if e.Field == "links" && strings.Contains(e.Message, "must use the referenced spec title") {
			t.Fatalf("unexpected spec link label error: %v", e)
		}
	}
}

func TestValidateSpec_MultipleErrors(t *testing.T) {
	// Missing everything
	content := []byte(`Just some text without structure.`)

	spec, err := ParseSpecContent("specs/001-test/SPEC.md", content)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	result := ValidateSpec(spec)
	if result.IsValid() {
		t.Error("expected validation to fail")
	}

	// Should have errors for frontmatter and title
	if len(result.Errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateSpecFile(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "specs", "000-test", "SPEC.md")

	content := []byte(`---
status: approved
author: File Author
---

# File Test

Description here.

## Task List

### Phase 1

- [ ] A task
`)

	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
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

func TestValidateSpecs_DuplicateFullRefs(t *testing.T) {
	content1 := []byte(`---
status: draft
---

# Feature A

## Task List

- [ ] Task 1
`)
	content2 := []byte(`---
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

	spec1.Path = "specs/001-feature-a/SPEC.md"
	spec2.Path = "specs/001-feature-b/SPEC.md"

	results := ValidateSpecs([]*Spec{spec1, spec2})

	// At least one spec should have a duplicate ref error.
	foundDuplicateError := false
	for _, result := range results {
		for _, e := range result.Errors {
			if e.Field == "fullref" && strings.Contains(e.Message, "duplicate ref") {
				foundDuplicateError = true
				break
			}
		}
	}
	if !foundDuplicateError {
		t.Error("expected duplicate ref error")
	}
}

func TestValidateSpecs_AllowDuplicateFullRefsAcrossScopes(t *testing.T) {
	content1 := []byte(`---
status: draft
---

# Feature A

## Task List

- [ ] Task 1
`)
	content2 := []byte(`---
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

	spec1.Path = "specs/000-mvp/001-feature-a/SPEC.md"
	spec2.Path = "specs/001-platform/001-feature-b/SPEC.md"

	results := ValidateSpecs([]*Spec{spec1, spec2})

	for _, result := range results {
		for _, e := range result.Errors {
			if e.Field == "fullref" && strings.Contains(e.Message, "duplicate ref") {
				t.Errorf("unexpected duplicate ref error: %v", e)
			}
		}
	}
}

func TestValidateSpecs_NoDuplicates(t *testing.T) {
	content1 := []byte(`---
status: draft
---

# Feature A

## Task List

- [ ] Task 1
`)
	content2 := []byte(`---
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

	spec1.Path = "specs/001-feature-a/SPEC.md"
	spec2.Path = "specs/002-feature-b/SPEC.md"

	results := ValidateSpecs([]*Spec{spec1, spec2})

	for _, result := range results {
		for _, e := range result.Errors {
			if e.Field == "fullref" && strings.Contains(e.Message, "duplicate ref") {
				t.Errorf("unexpected duplicate ref error: %v", e)
			}
		}
	}
}
