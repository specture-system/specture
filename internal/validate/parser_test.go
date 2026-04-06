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

	if err := os.WriteFile(specPath, content, 0o644); err != nil {
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
}
