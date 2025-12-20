package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestSetupCommand_CompleteWorkflow(t *testing.T) {
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

	// Run setup command
	out := &bytes.Buffer{}
	cmd := setupCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify specs directory was created
	specsDir := filepath.Join(tmpDir, "specs")
	if _, err := os.Stat(specsDir); err != nil {
		t.Errorf("specs directory was not created: %v", err)
	}

	// Verify specs/README.md was created
	readmePath := filepath.Join(specsDir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Errorf("specs/README.md was not created: %v", err)
	}

	// Verify content
	content := testhelpers.ReadFile(t, readmePath)
	if !testhelpers.Contains(content, "Spec Guidelines") {
		t.Error("README.md should contain 'Spec Guidelines'")
	}

	output := out.String()
	if !testhelpers.Contains(output, "Initialized Specture System") {
		t.Errorf("output should contain initialization message, got: %s", output)
	}
}

// Note: The dry-run tests are covered in the setup package unit tests
// (setup_test.go TestCreateSpecsDirectory_DryRun and TestCreateSpecsReadme_DryRun)
// Integration tests with the command-line flag are harder to test due to Cobra's
// flag parsing complexity in unit tests, so we rely on unit tests for dry-run behavior.
