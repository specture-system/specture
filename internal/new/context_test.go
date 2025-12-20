package new

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestNewContext_ErrorHandling(t *testing.T) {
	t.Run("fails_for_non_git_repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := NewContext(tmpDir, "Test Spec")
		if err == nil {
			t.Errorf("NewContext() expected error for non-git repo")
		}
	})

	t.Run("fails_for_dirty_working_tree", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create uncommitted changes
		dirtyFile := filepath.Join(tmpDir, "dirty.txt")
		if err := os.WriteFile(dirtyFile, []byte("changes"), 0644); err != nil {
			t.Fatalf("failed to create dirty file: %v", err)
		}

		_, err := NewContext(tmpDir, "Test Spec")
		if err == nil {
			t.Errorf("NewContext() expected error for dirty working tree")
		}
	})

	t.Run("succeeds_with_valid_repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		ctx, err := NewContext(tmpDir, "My First Spec")
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.Number != 0 {
			t.Errorf("NewContext() spec number = %d, want 0", ctx.Number)
		}
		if ctx.FileName != "000-my-first-spec.md" {
			t.Errorf("NewContext() filename = %q, want %q", ctx.FileName, "000-my-first-spec.md")
		}
		if ctx.BranchName != "spec/000-my-first-spec" {
			t.Errorf("NewContext() branch = %q, want %q", ctx.BranchName, "spec/000-my-first-spec")
		}
	})
}
