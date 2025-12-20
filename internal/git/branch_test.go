package git

import (
	"os/exec"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestCreateBranch(t *testing.T) {
	t.Run("creates_new_branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit so we can create a branch
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		err := CreateBranch(tmpDir, "test-branch")
		if err != nil {
			t.Fatalf("CreateBranch() error = %v", err)
		}

		// Verify branch was created by checking current branch
		cmd = exec.Command("git", "branch", "--show-current")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to get current branch: %v", err)
		}

		if string(output) != "test-branch\n" {
			t.Errorf("CreateBranch() failed to create branch, got current branch: %s", string(output))
		}
	})

	t.Run("fails_for_non_git_repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CreateBranch(tmpDir, "test-branch")
		if err == nil {
			t.Errorf("CreateBranch() expected error for non-git repo")
		}
	})
}

func TestPushBranch(t *testing.T) {
	t.Run("fails_without_remote", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		// Create branch
		if err := CreateBranch(tmpDir, "test-branch"); err != nil {
			t.Fatalf("failed to create branch: %v", err)
		}

		// Try to push without remote
		err := PushBranch(tmpDir, "origin", "test-branch")
		if err == nil {
			t.Errorf("PushBranch() expected error when remote doesn't exist")
		}
	})
}
