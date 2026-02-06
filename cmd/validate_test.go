package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/spec"
)

// Note: These tests intentionally do not use t.Parallel() because validateCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr calls).
// Parallel execution would cause tests to interfere with each other.
//
// Note: Tests that would trigger os.Exit(1) for invalid specs cannot be run
// directly as they would terminate the test process. We test the output format
// and that actual errors (like missing files) are returned properly.

func TestValidateCommand_AllSpecsValid(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Create a valid spec
	validSpec := `---
status: draft
author: Test Author
---

# My Feature

Description.

## Task List

- [ ] Task 1
`
	if err := os.WriteFile(filepath.Join(specsDir, "000-test.md"), []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWd) })
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("expected no error for valid spec, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "\u2713") {
		t.Errorf("expected checkmark in output, got: %s", output)
	}
	if !strings.Contains(output, "1 of 1 specs valid") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}

func TestValidateCommand_InvalidSpec(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Create an invalid spec (missing frontmatter)
	invalidSpec := `# My Feature

Description without frontmatter.

## Task List

- [ ] Task 1
`
	if err := os.WriteFile(filepath.Join(specsDir, "000-test.md"), []byte(invalidSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWd) })
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	invalidCount, err := runValidate(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if invalidCount != 1 {
		t.Errorf("expected 1 invalid spec, got: %d", invalidCount)
	}

	output := out.String()
	if !strings.Contains(output, "\u2717") {
		t.Errorf("expected X mark in output, got: %s", output)
	}
	if !strings.Contains(output, "0 of 1 specs valid") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}

func TestValidateCommand_ByNumber(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Create two specs
	validSpec := `---
status: draft
---

# Spec One

## Task List

- [ ] Task
`
	if err := os.WriteFile(filepath.Join(specsDir, "000-first.md"), []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "001-second.md"), []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		validateCmd.Flags().Set("spec", "") // Reset flag
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Set flag and run directly using --spec flag with short number
	cmd.Flags().Set("spec", "1")
	invalidCount, err := runValidate(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if invalidCount != 0 {
		t.Errorf("expected 0 invalid specs, got: %d", invalidCount)
	}

	output := out.String()
	if !strings.Contains(output, "001-second.md") {
		t.Errorf("expected 001-second.md in output, got: %s", output)
	}
	if strings.Contains(output, "000-first.md") {
		t.Errorf("did not expect 000-first.md in output, got: %s", output)
	}
	if !strings.Contains(output, "1 of 1 specs valid") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}

func TestValidateCommand_ByPath(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	validSpec := `---
status: approved
---

# Test

## Task List

- [x] Done
`
	specPath := filepath.Join(specsDir, "000-test.md")
	if err := os.WriteFile(specPath, []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		validateCmd.Flags().Set("spec", "") // Reset flag
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Set flag and run directly using --spec flag with path
	cmd.Flags().Set("spec", specPath)
	invalidCount, err := runValidate(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if invalidCount != 0 {
		t.Errorf("expected 0 invalid specs, got: %d", invalidCount)
	}

	output := out.String()
	if !strings.Contains(output, "000-test.md") {
		t.Errorf("expected 000-test.md in output, got: %s", output)
	}
}

func TestValidateCommand_NoSpecs(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Create README.md (not a spec)
	if err := os.WriteFile(filepath.Join(specsDir, "README.md"), []byte("# Specs"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWd) })
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No specs found") {
		t.Errorf("expected 'No specs found' message, got: %s", output)
	}
}

func TestValidateCommand_SpecNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() {
		os.Chdir(originalWd)
		validateCmd.Flags().Set("spec", "") // Reset flag
	})
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Set flag and run directly
	cmd.Flags().Set("spec", "999")
	_, err := runValidate(cmd, []string{})
	if err == nil {
		t.Error("expected error for nonexistent spec")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestValidateCommand_NoSpecsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (no specs dir)
	originalWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWd) })
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for missing specs directory")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestValidateCommand_MixedResults(t *testing.T) {
	tmpDir := t.TempDir()
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	// Create valid spec
	validSpec := `---
status: draft
---

# Valid

## Task List

- [ ] Task
`
	if err := os.WriteFile(filepath.Join(specsDir, "000-valid.md"), []byte(validSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Create invalid spec
	invalidSpec := `# Invalid

No frontmatter.
`
	if err := os.WriteFile(filepath.Join(specsDir, "001-invalid.md"), []byte(invalidSpec), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalWd) })
	os.Chdir(tmpDir)

	out := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Use runValidate directly to avoid os.Exit(1) in production code
	invalidCount, err := runValidate(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that we got 1 invalid spec
	if invalidCount != 1 {
		t.Errorf("expected 1 invalid spec, got: %d", invalidCount)
	}

	output := out.String()
	if !strings.Contains(output, "1 of 2 specs valid") {
		t.Errorf("expected '1 of 2 specs valid' in output, got: %s", output)
	}
}

func TestFindAllSpecs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various files
	files := []string{
		"000-first.md",
		"001-second.md",
		"README.md",
		"notes.txt",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	paths, err := spec.FindAll(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 specs, got %d: %v", len(paths), paths)
	}
}

func TestResolveSpecPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a spec file
	specPath := filepath.Join(tmpDir, "000-test.md")
	if err := os.WriteFile(specPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"by three-digit number", "000", false},
		{"by two-digit number", "00", false},
		{"by single-digit number", "0", false},
		{"by full path", specPath, false},
		{"nonexistent number", "999", true},
		{"invalid format", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := spec.ResolvePath(tmpDir, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolvePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
