package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

// Note: These tests intentionally do not use t.Parallel() because setupCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr calls).
// Parallel execution would cause tests to interfere with each other.

func TestSetupCommand_CompleteWorkflow_DryRun(t *testing.T) {
	// Create a temporary clean git repository
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Change to the repository
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

	// Run setup command with dry-run flag
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Parse and set the flag directly
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify specs directory was NOT created (dry-run mode)
	specsDir := filepath.Join(tmpDir, "specs")
	if _, err := os.Stat(specsDir); err == nil {
		t.Error("specs directory should not be created in dry-run mode")
	}

	// Verify output contains setup summary
	output := out.String()
	if !testhelpers.Contains(output, "Detected forge") {
		t.Errorf("output should contain forge detection, got: %s", output)
	}
	if !testhelpers.Contains(output, "Create specs/ directory") {
		t.Errorf("output should list what will be created, got: %s", output)
	}
}

func TestSetupCommand_ExistingSpecFiles_Error(t *testing.T) {
	// Create a temporary git repository with existing spec files
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Create specs directory with an existing spec file
	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	testhelpers.WriteFile(t, specsDir, "000-mvp.md", "# MVP Spec")

	// Commit the file to make working tree clean
	gitCmd := exec.Command("git", "add", "specs/")
	gitCmd.Dir = tmpDir
	if err := gitCmd.Run(); err != nil {
		t.Fatalf("failed to add specs: %v", err)
	}

	gitCmd = exec.Command("git", "commit", "-m", "add spec")
	gitCmd.Dir = tmpDir
	if err := gitCmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Change to the repository
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

	// Run setup command
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when existing spec files found, got nil")
	}
	if !testhelpers.Contains(err.Error(), "cannot be overwritten") {
		t.Errorf("error should mention spec file protection, got: %v", err)
	}
}

func TestSetupCommand_OutputSummary(t *testing.T) {
	// Create a temporary clean git repository
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Change to the repository
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

	// Run setup command with dry-run flag (avoids user confirmation)
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Parse and set the flag directly
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("failed to set dry-run flag: %v", err)
	}

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify output contains expected summary items
	output := out.String()
	if !testhelpers.Contains(output, "Setup will:") {
		t.Errorf("output should contain 'Setup will:' summary, got: %s", output)
	}
	if !testhelpers.Contains(output, "Create specs/ directory") {
		t.Errorf("output should list specs directory creation, got: %s", output)
	}
	if !testhelpers.Contains(output, "Create specs/README.md") {
		t.Errorf("output should list README.md creation, got: %s", output)
	}
}
