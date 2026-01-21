package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

// Note: These tests intentionally do not use t.Parallel() because newCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr calls).
// Parallel execution would cause tests to interfere with each other.

// newTestContext prepares a temporary git repository and changes to it.
// It returns the directory path and registers cleanup.
func newTestContext(t *testing.T) string {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Create initial commit so there's a branch to check out
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalWd)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	return tmpDir
}

// runNewCommand runs the new command with the given title in dry-run mode.
// This avoids stdin complexity and focuses on testing the core logic.
func runNewCommand(t *testing.T, title string) string {
	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Set up stdin for title prompt
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	// Write title only (dry-run won't prompt for more)
	go func() {
		defer w.Close()
		if _, err := w.WriteString(title + "\n"); err != nil {
			t.Logf("failed to write to pipe: %v", err)
		}
	}()

	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	// Set dry-run flag
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	return out.String()
}

func TestNewCommand_DryRunMode(t *testing.T) {
	tmpDir := newTestContext(t)
	output := runNewCommand(t, "My First Feature")

	// Verify spec file was NOT created (dry-run mode)
	specPath := filepath.Join(tmpDir, "specs", "000-my-first-feature.md")
	if _, err := os.Stat(specPath); err == nil {
		t.Error("spec file should not be created in dry-run mode")
	}

	// Verify output contains dry-run message
	if !strings.Contains(output, "[dry-run] No changes made") {
		t.Errorf("output should contain dry-run message, got: %s", output)
	}

	// Verify output shows what would be created
	expectedItems := []string{
		"Creating spec 000",
		"My First Feature",
		"Branch: spec/000-my-first-feature",
		"File: 000-my-first-feature.md",
	}

	for _, item := range expectedItems {
		if !strings.Contains(output, item) {
			t.Errorf("output should contain %q, got: %s", item, output)
		}
	}
}

func TestNewCommand_OutputSummary(t *testing.T) {
	newTestContext(t)
	output := runNewCommand(t, "Test Feature")

	// Verify output shows spec details
	expectedItems := []string{
		"Creating spec 000",
		"Test Feature",
		"Branch: spec/000-test-feature",
		"File: 000-test-feature.md",
		"Author:",
	}

	for _, item := range expectedItems {
		if !strings.Contains(output, item) {
			t.Errorf("output should contain %q, got: %s", item, output)
		}
	}
}

func TestNewCommand_UserConfirmation_DryRun(t *testing.T) {
	tmpDir := newTestContext(t)
	_ = runNewCommand(t, "Cancelled Feature")

	// Verify spec was not created in dry-run
	specPath := filepath.Join(tmpDir, "specs", "000-cancelled-feature.md")
	if _, err := os.Stat(specPath); err == nil {
		t.Error("spec file should not be created in dry-run mode")
	}
}

func TestNewCommand_SpecNumbering_FirstSpec(t *testing.T) {
	// Test that first spec is numbered 000
	newTestContext(t)

	output := runNewCommand(t, "First Feature")
	if !strings.Contains(output, "Creating spec 000") {
		t.Errorf("first spec should be numbered 000, got: %s", output)
	}
	if !strings.Contains(output, "000-first-feature.md") {
		t.Errorf("spec filename should contain 000 prefix, got: %s", output)
	}
}

func TestNewCommand_SpecNumbering_WithExistingSpec(t *testing.T) {
	// Test that numbering correctly identifies next available number
	tmpDir := newTestContext(t)

	// Create an existing spec file and commit it
	specsDir := filepath.Join(tmpDir, "specs")
	os.MkdirAll(specsDir, 0755)
	specFile := filepath.Join(specsDir, "000-existing.md")
	os.WriteFile(specFile, []byte("# Existing"), 0644)

	// Commit the file so working tree is clean
	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Add and commit the file
	gitAddCmd := exec.Command("git", "add", "specs/000-existing.md")
	gitAddCmd.Dir = tmpDir
	gitAddCmd.Run()

	gitCommitCmd := exec.Command("git", "commit", "-m", "Add existing spec")
	gitCommitCmd.Dir = tmpDir
	gitCommitCmd.Run()

	// Now second spec should detect the existing one and increment
	output := runNewCommand(t, "Next Feature")
	if !strings.Contains(output, "Creating spec 001") {
		t.Errorf("second spec should be numbered 001 when spec 000 exists, got: %s", output)
	}
}

func TestNewCommand_BranchNameGeneration(t *testing.T) {
	// Test that branch names are correctly generated from spec title
	newTestContext(t)
	output := runNewCommand(t, "API Authentication")

	expectedBranch := "spec/000-api-authentication"
	if !strings.Contains(output, expectedBranch) {
		t.Errorf("output should contain branch name %q, got: %s", expectedBranch, output)
	}
}

func TestNewCommand_KebabCaseConversion(t *testing.T) {
	// Test that spec titles are converted to kebab-case
	newTestContext(t)
	output := runNewCommand(t, "My Complex_Feature Title")

	expectedFilename := "000-my-complex-feature-title.md"
	if !strings.Contains(output, expectedFilename) {
		t.Errorf("output should contain kebab-case filename %q, got: %s", expectedFilename, output)
	}
}

func TestNewCommand_EmptyTitle(t *testing.T) {
	newTestContext(t)

	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Setup stdin with empty title
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()

	if _, err := w.WriteString("\n"); err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}
	w.Close()

	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
	})

	// Command should fail with empty title
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("new command should fail with empty title")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("error should mention empty title, got: %v", err)
	}
}

func TestNewCommand_GetsAuthorFromGit(t *testing.T) {
	// Test that the command correctly retrieves author from git config
	newTestContext(t)
	output := runNewCommand(t, "Feature with Author")

	// Should contain "Author: Test User" (from testhelpers.InitGitRepo)
	if !strings.Contains(output, "Author:") {
		t.Errorf("output should contain Author field, got: %s", output)
	}
	if !strings.Contains(output, "Test User") {
		t.Errorf("output should contain author name from git config, got: %s", output)
	}
}

func TestNewCommand_ValidateContextCreation(t *testing.T) {
	// Test that NewContext can be created successfully in a git repo
	newTestContext(t)
	output := runNewCommand(t, "Valid Spec")

	// Should show spec creation summary
	if !strings.Contains(output, "Creating spec") {
		t.Errorf("output should show successful spec summary, got: %s", output)
	}
	if strings.Contains(output, "Error") || strings.Contains(output, "error") {
		t.Errorf("output should not contain errors, got: %s", output)
	}
}

func TestNewCommand_SpecsDirectoryCreated(t *testing.T) {
	// Test that specs directory is created if it doesn't exist
	tmpDir := newTestContext(t)

	// specs directory should not exist yet (NewContext creates it)
	specsDir := filepath.Join(tmpDir, "specs")

	_ = runNewCommand(t, "First Spec")

	// After running the command (even in dry-run), specs dir should exist
	if _, err := os.Stat(specsDir); err != nil {
		t.Errorf("specs directory should be created, got error: %v", err)
	}

	// But spec file should NOT exist (dry-run mode)
	specFile := filepath.Join(specsDir, "000-first-spec.md")
	if _, err := os.Stat(specFile); err == nil {
		t.Error("spec file should not be created in dry-run mode")
	}
}

func TestNewCommand_TitleFlag(t *testing.T) {
	newTestContext(t)

	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset flags
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("title", "")
	cmd.Flags().Set("no-editor", "false")

	// Set title and dry-run flags
	if err := cmd.Flags().Set("title", "Feature from Flag"); err != nil {
		t.Fatalf("failed to set title flag: %v", err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	// Should not prompt for stdin since title is provided
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Feature from Flag") {
		t.Errorf("output should contain provided title, got: %s", output)
	}
	if !strings.Contains(output, "Creating spec 000") {
		t.Errorf("output should show spec creation, got: %s", output)
	}
}

func TestNewCommand_TitleFlagSkipsConfirmation(t *testing.T) {
	newTestContext(t)

	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset flags
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("title", "")
	cmd.Flags().Set("no-editor", "false")

	// Set title and dry-run flags (dry-run to avoid needing stdin for confirmation)
	if err := cmd.Flags().Set("title", "Feature from Flag"); err != nil {
		t.Fatalf("failed to set title flag: %v", err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	output := out.String()
	// When title is provided, confirmation prompt should be skipped
	if strings.Contains(output, "Proceed with creating spec?") {
		t.Errorf("output should NOT contain confirmation prompt when title is provided, got: %s", output)
	}
}

func TestNewCommand_NoEditorFlag(t *testing.T) {
	newTestContext(t)

	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset flags
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("title", "")
	cmd.Flags().Set("no-editor", "false")

	// Set flags
	if err := cmd.Flags().Set("title", "Feature without Editor"); err != nil {
		t.Fatalf("failed to set title flag: %v", err)
	}
	if err := cmd.Flags().Set("no-editor", "true"); err != nil {
		t.Fatalf("failed to set no-editor flag: %v", err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	output := out.String()
	// When no-editor is set, should not try to open editor
	if strings.Contains(output, "Opening") {
		t.Errorf("output should NOT contain 'Opening' editor message when --no-editor is set, got: %s", output)
	}
}
