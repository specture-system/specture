package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestNewContext_ValidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	ctx, err := NewContext(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if ctx.WorkDir != tmpDir {
		t.Errorf("expected WorkDir %s, got %s", tmpDir, ctx.WorkDir)
	}

	if ctx.Terminology == "" {
		t.Error("expected Terminology to be set")
	}
}

func TestNewContext_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	ctx, err := NewContext(tmpDir)
	if err == nil {
		t.Fatalf("expected error for non-git repo, got nil context: %v", ctx)
	}

	if err.Error() != "not a git repository" {
		t.Errorf("expected 'not a git repository' error, got: %v", err)
	}
}

func TestNewContext_DirtyWorkingTree(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	// Create an uncommitted file
	uncommittedFile := filepath.Join(tmpDir, "uncommitted.txt")
	if err := os.WriteFile(uncommittedFile, []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("failed to create uncommitted file: %v", err)
	}

	ctx, err := NewContext(tmpDir)
	if err == nil {
		t.Fatalf("expected error for dirty working tree, got nil context: %v", ctx)
	}

	if err.Error() != "working tree has uncommitted changes" {
		t.Errorf("expected 'working tree has uncommitted changes' error, got: %v", err)
	}
}

func TestCreateSpecsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	ctx, err := NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Create specs directory
	if err := ctx.CreateSpecsDirectory(false); err != nil {
		t.Fatalf("failed to create specs directory: %v", err)
	}

	// Check that directory was created
	specsDir := filepath.Join(tmpDir, "specs")
	if _, err := os.Stat(specsDir); err != nil {
		t.Errorf("specs directory was not created: %v", err)
	}
}

func TestCreateSpecsDirectory_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	ctx, err := NewContext(tmpDir)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// Create specs directory in dry-run mode
	if err := ctx.CreateSpecsDirectory(true); err != nil {
		t.Fatalf("failed in dry-run mode: %v", err)
	}

	// Check that directory was NOT created
	specsDir := filepath.Join(tmpDir, "specs")
	if _, err := os.Stat(specsDir); err == nil {
		t.Error("specs directory should not be created in dry-run mode")
	}
}
