package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestCommitChanges(t *testing.T) {
	t.Run("commits_staged_changes", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		// Create a test file
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Commit the file
		err := CommitChanges(tmpDir, "test: add test file", testFile)
		if err != nil {
			t.Fatalf("CommitChanges() error = %v", err)
		}

		// Verify commit was created by checking log
		cmd = exec.Command("git", "log", "--oneline", "-n", "1")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to check git log: %v", err)
		}

		if len(output) == 0 {
			t.Errorf("CommitChanges() failed to create commit")
		}
	})

	t.Run("fails_for_non_git_repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		err := CommitChanges(tmpDir, "test commit", testFile)
		if err == nil {
			t.Errorf("CommitChanges() expected error for non-git repo")
		}
	})

	t.Run("commits_multiple_files", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		// Create initial commit
		cmd := exec.Command("git", "commit", "--allow-empty", "-m", "initial")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create initial commit: %v", err)
		}

		// Create test files
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")
		if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
			t.Fatalf("failed to create file1: %v", err)
		}
		if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
			t.Fatalf("failed to create file2: %v", err)
		}

		// Commit both files
		err := CommitChanges(tmpDir, "test: add multiple files", file1, file2)
		if err != nil {
			t.Fatalf("CommitChanges() error = %v", err)
		}

		// Verify both files are tracked
		cmd = exec.Command("git", "ls-tree", "HEAD")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to check git tree: %v", err)
		}

		if len(output) == 0 {
			t.Errorf("CommitChanges() failed to commit files")
		}
	})
}
