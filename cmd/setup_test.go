package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestSetupCommand_NotGitRepo(t *testing.T) {
	// Create a temporary directory that is not a git repo
	tmpDir := t.TempDir()

	// Change to the temporary directory
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
		t.Fatal("expected error when running setup in non-git repo, got nil")
	}
	if err.Error() != "not a git repository" {
		t.Errorf("expected 'not a git repository' error, got: %v", err)
	}
}

func TestSetupCommand_DirtyWorkingTree(t *testing.T) {
	// Create a temporary git repository with uncommitted changes
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Create an uncommitted file
	testFile := filepath.Join(tmpDir, "uncommitted.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted content"), 0644); err != nil {
		t.Fatalf("failed to create uncommitted file: %v", err)
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
		t.Fatal("expected error when running setup with uncommitted changes, got nil")
	}
	if err.Error() != "working tree has uncommitted changes" {
		t.Errorf("expected 'working tree has uncommitted changes' error, got: %v", err)
	}
}

func TestSetupCommand_ValidGitRepo_CleanWorkingTree(t *testing.T) {
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
	cmd.SetArgs([]string{"--dry-run"})

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected no error with clean repo and dry-run flag, got: %v", err)
	}
}
