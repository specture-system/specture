package new

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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

		// Create initial commit so there's a branch to check out
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

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

		// Branch name should include date suffix (YYYY-MM-DD)
		today := time.Now().Format("2006-01-02")
		expectedBranchPattern := "spec/000-my-first-spec-" + regexp.QuoteMeta(today)
		if !regexp.MustCompile(expectedBranchPattern).MatchString(ctx.BranchName) {
			t.Errorf("NewContext() branch = %q, want pattern %q", ctx.BranchName, expectedBranchPattern)
		}
	})
}

func TestCleanup(t *testing.T) {
	t.Run("removes_spec_file_and_deletes_branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		// Create context and spec
		ctx, err := NewContext(tmpDir, "Test Spec")
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		// Manually create the file and branch
		if err := os.WriteFile(ctx.FilePath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create spec file: %v", err)
		}

		createBranchCmd := exec.Command("git", "checkout", "-b", ctx.BranchName)
		createBranchCmd.Dir = tmpDir
		if err := createBranchCmd.Run(); err != nil {
			t.Fatalf("failed to create branch: %v", err)
		}

		// Verify file and branch exist
		if _, err := os.Stat(ctx.FilePath); err != nil {
			t.Errorf("spec file should exist before cleanup")
		}

		// Run cleanup
		if err := ctx.Cleanup(); err != nil {
			t.Fatalf("Cleanup() error = %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(ctx.FilePath); err == nil {
			t.Errorf("spec file should be removed after cleanup")
		}

		// Verify we're back on the original branch
		currentBranchCmd := exec.Command("git", "branch", "--show-current")
		currentBranchCmd.Dir = tmpDir
		output, err := currentBranchCmd.Output()
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}

		currentBranch := strings.TrimSpace(string(output))
		if currentBranch != ctx.OriginalBranch {
			t.Errorf("should be on original branch %q, got %q", ctx.OriginalBranch, currentBranch)
		}

		// Verify spec branch is deleted
		checkBranchCmd := exec.Command("git", "rev-parse", "--verify", ctx.BranchName)
		checkBranchCmd.Dir = tmpDir
		if err := checkBranchCmd.Run(); err == nil {
			t.Errorf("spec branch should be deleted after cleanup")
		}
	})
}
