package git

import (
	"os/exec"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestGetCurrentBranch(t *testing.T) {
	t.Run("returns_current_branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		branch, err := GetCurrentBranch(tmpDir)
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}

		// Default branch is main or master
		if branch != "main" && branch != "master" {
			t.Errorf("GetCurrentBranch() = %q, want 'main' or 'master'", branch)
		}
	})

	t.Run("returns_custom_branch", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		// Create and checkout custom branch
		cmd = exec.Command("git", "checkout", "-b", "custom-branch")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create branch: %v", err)
		}

		branch, err := GetCurrentBranch(tmpDir)
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}

		if branch != "custom-branch" {
			t.Errorf("GetCurrentBranch() = %q, want 'custom-branch'", branch)
		}
	})
}
