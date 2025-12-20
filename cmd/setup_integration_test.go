package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

// Note: These tests intentionally do not use t.Parallel() because setupCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr calls).
// Parallel execution would cause tests to interfere with each other.

// setupTestContext prepares a temporary git repository and changes to it.
// It returns the directory path and registers cleanup.
func setupTestContext(t *testing.T) string {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

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

// runSetupCommand runs the setup command and returns output.
// If dryRun is true, the command runs in dry-run mode.
func runSetupCommand(t *testing.T, dryRun bool) string {
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	if dryRun {
		if err := cmd.Flags().Set("dry-run", "true"); err != nil {
			t.Fatalf("failed to set dry-run flag: %v", err)
		}
	}

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	return out.String()
}

func TestSetupCommand_CompleteWorkflow_DryRun(t *testing.T) {
	tmpDir := setupTestContext(t)
	output := runSetupCommand(t, true)

	// Verify specs directory was NOT created (dry-run mode)
	specsDir := filepath.Join(tmpDir, "specs")
	if _, err := os.Stat(specsDir); err == nil {
		t.Error("specs directory should not be created in dry-run mode")
	}

	// Verify output contains setup summary
	if !strings.Contains(output, "Detected forge") {
		t.Errorf("output should contain forge detection, got: %s", output)
	}
	if !strings.Contains(output, "Create specs/ directory") {
		t.Errorf("output should list what will be created, got: %s", output)
	}
}

func TestSetupCommand_OutputSummary(t *testing.T) {
	setupTestContext(t)
	output := runSetupCommand(t, true)

	// Verify output contains expected summary items
	if !strings.Contains(output, "Setup will:") {
		t.Errorf("output should contain 'Setup will:' summary, got: %s", output)
	}
	if !strings.Contains(output, "Create specs/ directory") {
		t.Errorf("output should list specs directory creation, got: %s", output)
	}
	if !strings.Contains(output, "Create specs/README.md") {
		t.Errorf("output should list README.md creation, got: %s", output)
	}
}

func TestSetupCommand_DryRunPreviewsAllActions(t *testing.T) {
	setupTestContext(t)
	output := runSetupCommand(t, true)

	// Verify dry-run shows all actions that would be performed
	expectedItems := []string{
		"Detected forge",
		"Setup will:",
		"Create specs/ directory",
		"Create specs/README.md",
	}

	for _, item := range expectedItems {
		if !strings.Contains(output, item) {
			t.Errorf("output should contain %q, got: %s", item, output)
		}
	}
}
