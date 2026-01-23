package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpecContent_ValidSpec(t *testing.T) {
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
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Frontmatter == nil {
		t.Fatal("expected frontmatter to be parsed")
	}
	if spec.Frontmatter.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", spec.Frontmatter.Status)
	}
	if spec.Frontmatter.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %q", spec.Frontmatter.Author)
	}
	if spec.Title != "My Feature" {
		t.Errorf("expected title 'My Feature', got %q", spec.Title)
	}
	if !spec.HasTaskList {
		t.Error("expected HasTaskList to be true")
	}
}

func TestParseSpecContent_NoFrontmatter(t *testing.T) {
	content := []byte(`# My Feature

This is a description.
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Frontmatter != nil {
		t.Error("expected frontmatter to be nil")
	}
	if spec.Title != "My Feature" {
		t.Errorf("expected title 'My Feature', got %q", spec.Title)
	}
}

func TestParseSpecContent_NoTitle(t *testing.T) {
	content := []byte(`---
status: draft
---

No heading here, just content.
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Title != "" {
		t.Errorf("expected empty title, got %q", spec.Title)
	}
}

func TestParseSpecContent_NoTaskList(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

This is just a description with no Task List heading.

- Regular list item
- Another item
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.HasTaskList {
		t.Error("expected HasTaskList to be false")
	}
}

func TestParseSpecContent_TaskListHeadingOnly(t *testing.T) {
	// Task List heading without any checkbox items should still be valid
	content := []byte(`---
status: draft
---

# My Feature

This spec is in design phase.

## Task List

Tasks will be added later.
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !spec.HasTaskList {
		t.Error("expected HasTaskList to be true (heading present)")
	}
}

func TestParseSpecContent_TaskListH3Heading_NotValid(t *testing.T) {
	// H3 Task List heading should NOT be valid (must be H2)
	content := []byte(`---
status: draft
---

# My Feature

## Implementation

### Task List

- [ ] Task 1
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.HasTaskList {
		t.Error("expected HasTaskList to be false (H3 is not valid, must be H2)")
	}
}

func TestParseSpecContent_OnlyUncheckedTasks(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

## Task List

- [ ] Task 1
- [ ] Task 2
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !spec.HasTaskList {
		t.Error("expected HasTaskList to be true")
	}
}

func TestParseSpecContent_OnlyCheckedTasks(t *testing.T) {
	content := []byte(`---
status: draft
---

# My Feature

## Task List

- [x] Task 1
- [x] Task 2
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !spec.HasTaskList {
		t.Error("expected HasTaskList to be true")
	}
}

func TestParseSpec_File(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "000-test.md")

	content := []byte(`---
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

	spec, err := ParseSpec(specPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Path != specPath {
		t.Errorf("expected path %q, got %q", specPath, spec.Path)
	}
	if spec.Frontmatter.Status != "approved" {
		t.Errorf("expected status 'approved', got %q", spec.Frontmatter.Status)
	}
	if spec.Title != "File Test" {
		t.Errorf("expected title 'File Test', got %q", spec.Title)
	}
	if !spec.HasTaskList {
		t.Error("expected HasTaskList to be true")
	}
}

func TestParseSpec_FileNotFound(t *testing.T) {
	_, err := ParseSpec("/nonexistent/path/to/spec.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseSpecContent_H2HeadingNotTitle(t *testing.T) {
	// H2 headings should not be treated as the title
	content := []byte(`---
status: draft
---

## Not a Title

Description.
`)

	spec, err := ParseSpecContent("test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Title != "" {
		t.Errorf("expected empty title (H2 is not title), got %q", spec.Title)
	}
}
