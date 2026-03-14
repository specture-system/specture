package git

import (
	"os/exec"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestBranchExists(t *testing.T) {
	tmpDir := t.TempDir()
	testhelpers.InitGitRepo(t, tmpDir)

	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	if err := CreateBranch(tmpDir, "feature/test"); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	checkoutMain := exec.Command("git", "checkout", "main")
	checkoutMain.Dir = tmpDir
	if err := checkoutMain.Run(); err != nil {
		checkoutMaster := exec.Command("git", "checkout", "master")
		checkoutMaster.Dir = tmpDir
		if err := checkoutMaster.Run(); err != nil {
			t.Fatalf("failed to checkout default branch: %v", err)
		}
	}

	exists, err := BranchExists(tmpDir, "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected branch to exist")
	}

	notExists, err := BranchExists(tmpDir, "feature/missing")
	if err != nil {
		t.Fatalf("unexpected error for missing branch: %v", err)
	}
	if notExists {
		t.Fatal("expected missing branch to return false")
	}
}
