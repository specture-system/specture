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

	// Reset all flags to their defaults
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("yes", "false")
	cmd.Flags().Set("update-agents", "false")
	cmd.Flags().Set("no-update-agents", "false")
	cmd.Flags().Set("update-claude", "false")
	cmd.Flags().Set("no-update-claude", "false")

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

func TestSetupCommand_CreatesFilesWithCorrectContent(t *testing.T) {
	tmpDir := setupTestContext(t)

	// Setup stdin to automatically answer "yes" to confirmation
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()

	if _, err := w.WriteString("yes\n"); err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}
	w.Close()

	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
	})

	// Run setup without dry-run (reset flag from previous tests)
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset dry-run flag to false (it might be set to true from previous test)
	if err := cmd.Flags().Set("dry-run", "false"); err != nil {
		t.Fatalf("failed to reset dry-run flag: %v", err)
	}

	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("setup command failed: %v", err)
	}

	// Verify specs directory was created
	specsDir := filepath.Join(tmpDir, "specs")
	if info, err := os.Stat(specsDir); err != nil || !info.IsDir() {
		t.Errorf("specs directory should be created, got error: %v", err)
	}

	// Verify specs/README.md was created
	readmePath := filepath.Join(tmpDir, "specs", "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Errorf("specs/README.md should be created, got error: %v", err)
	}

	// Verify README content contains expected sections
	contentStr := string(content)
	expectedContent := []string{
		"Spec Guidelines",
		"Spec Scope",
		"Spec File Structure",
		"Workflow",
		"pull request", // ContributionType should be rendered
	}

	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("README.md should contain %q, full content:\n%s", expected, contentStr)
		}
	}
}

func TestSetupCommand_UpdateAgentsFlag_WithoutFile(t *testing.T) {
	tmpDir := setupTestContext(t)

	// Ensure AGENTS.md doesn't exist
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		t.Fatal("AGENTS.md should not exist in test directory")
	}

	// Run setup with --update-agents flag and --yes in dry-run mode to avoid prompts
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset all flags to their defaults
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("yes", "false")
	cmd.Flags().Set("update-agents", "false")
	cmd.Flags().Set("no-update-agents", "false")
	cmd.Flags().Set("update-claude", "false")
	cmd.Flags().Set("no-update-claude", "false")

	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("update-agents", "true")

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("setup command failed: %v", err)
	}

	// Verify output mentions AGENTS.md will be updated
	output := out.String()
	if !strings.Contains(output, "Show update prompt for AGENTS.md") {
		t.Errorf("output should indicate AGENTS.md will be prompted for update, got: %s", output)
	}
}

func TestSetupCommand_UpdateAgentsFlag_WithFile(t *testing.T) {
	tmpDir := setupTestContext(t)

	// Create AGENTS.md file
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("existing agents file"), 0644); err != nil {
		t.Fatalf("failed to create AGENTS.md: %v", err)
	}

	// Commit the file so working tree is clean
	out := &bytes.Buffer{}
	if err := testhelpers.RunGitCommand(tmpDir, []string{"add", agentsPath}); err != nil {
		t.Fatalf("failed to stage AGENTS.md: %v", err)
	}
	if err := testhelpers.RunGitCommand(tmpDir, []string{"commit", "-m", "Add AGENTS.md"}); err != nil {
		t.Fatalf("failed to commit AGENTS.md: %v", err)
	}

	// Run setup with --update-agents flag in dry-run mode to avoid prompts
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset all flags to their defaults
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("yes", "false")
	cmd.Flags().Set("update-agents", "false")
	cmd.Flags().Set("no-update-agents", "false")
	cmd.Flags().Set("update-claude", "false")
	cmd.Flags().Set("no-update-claude", "false")

	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("update-agents", "true")

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("setup command failed: %v", err)
	}

	// Verify output mentions AGENTS.md will be updated
	output := out.String()
	if !strings.Contains(output, "Show update prompt for AGENTS.md") {
		t.Errorf("output should indicate AGENTS.md will be prompted for update, got: %s", output)
	}
}

func TestSetupCommand_NoUpdateAgentsFlag_SkipsPrompt(t *testing.T) {
	tmpDir := setupTestContext(t)

	// Create AGENTS.md file
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("existing agents file"), 0644); err != nil {
		t.Fatalf("failed to create AGENTS.md: %v", err)
	}

	// Commit the file so working tree is clean
	if err := testhelpers.RunGitCommand(tmpDir, []string{"add", agentsPath}); err != nil {
		t.Fatalf("failed to stage AGENTS.md: %v", err)
	}
	if err := testhelpers.RunGitCommand(tmpDir, []string{"commit", "-m", "Add AGENTS.md"}); err != nil {
		t.Fatalf("failed to commit AGENTS.md: %v", err)
	}

	// Run setup with --no-update-agents flag in dry-run mode
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	// Reset all flags to their defaults
	cmd.Flags().Set("dry-run", "false")
	cmd.Flags().Set("yes", "false")
	cmd.Flags().Set("update-agents", "false")
	cmd.Flags().Set("no-update-agents", "false")
	cmd.Flags().Set("update-claude", "false")
	cmd.Flags().Set("no-update-claude", "false")

	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("no-update-agents", "true")

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("setup command failed: %v", err)
	}

	// Verify output does NOT mention AGENTS.md will be updated
	output := out.String()
	if strings.Contains(output, "Show update prompt for AGENTS.md") {
		t.Errorf("output should NOT indicate AGENTS.md will be prompted for update, got: %s", output)
	}
}
