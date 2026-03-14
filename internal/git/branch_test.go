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

func TestCheckoutBranch(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	if err := CreateBranch(tmpDir, "feature/checkout-test"); err != nil {
		t.Fatalf("failed to create test branch: %v", err)
	}

	if err := CheckoutBranch(tmpDir, "main"); err != nil {
		if err := CheckoutBranch(tmpDir, "master"); err != nil {
			t.Fatalf("failed to checkout default branch: %v", err)
		}
	}

	if err := CheckoutBranch(tmpDir, "feature/checkout-test"); err != nil {
		t.Fatalf("CheckoutBranch() error = %v", err)
	}

	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	if string(output) != "feature/checkout-test\n" {
		t.Fatalf("expected to be on feature/checkout-test, got %q", string(output))
	}
}

func TestDeleteBranch(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	if err := CreateBranch(tmpDir, "feature/delete-test"); err != nil {
		t.Fatalf("failed to create test branch: %v", err)
	}

	if err := CheckoutBranch(tmpDir, "main"); err != nil {
		if err := CheckoutBranch(tmpDir, "master"); err != nil {
			t.Fatalf("failed to checkout default branch: %v", err)
		}
	}

	if err := DeleteBranch(tmpDir, "feature/delete-test"); err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	exists, err := BranchExists(tmpDir, "feature/delete-test")
	if err != nil {
		t.Fatalf("failed to check branch existence: %v", err)
	}
	if exists {
		t.Fatal("expected branch to be deleted")
	}
}
