package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestHasUncommittedChanges(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(dir string) error
		testDir        *string
		hasUncommitted bool
		wantErr        string // empty string means no error expected
	}{
		{
			name: "clean working tree",
			setup: func(dir string) error {
				if err := setupGitRepo(t, dir); err != nil {
					return err
				}
				return commitFile(t, dir, "test.txt", "content")
			},
			hasUncommitted: false,
		},
		{
			name: "uncommitted changes",
			setup: func(dir string) error {
				if err := setupGitRepo(t, dir); err != nil {
					return err
				}
				if err := commitFile(t, dir, "test.txt", "content"); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "test.txt"), []byte("modified"), 0644)
			},
			hasUncommitted: true,
		},
		{
			name: "untracked files",
			setup: func(dir string) error {
				if err := setupGitRepo(t, dir); err != nil {
					return err
				}
				if err := commitFile(t, dir, "tracked.txt", "content"); err != nil {
					return err
				}
				_, err := os.Create(filepath.Join(dir, "untracked.txt"))
				return err
			},
			hasUncommitted: true,
		},
		{
			name:           "not a git repository",
			setup:          func(dir string) error { return nil },
			hasUncommitted: false,
			wantErr:        "failed to check git status",
		},
		{
			name:           "nonexistent directory",
			setup:          func(dir string) error { return nil },
			testDir:        strPtr("/nonexistent/path"),
			hasUncommitted: false,
			wantErr:        "failed to check git status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			testDir := dir
			if tt.testDir != nil {
				testDir = *tt.testDir
			}

			hasUncommitted, err := HasUncommittedChanges(testDir)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("HasUncommittedChanges() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("HasUncommittedChanges() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("HasUncommittedChanges() unexpected error = %v", err)
				return
			}
			if hasUncommitted != tt.hasUncommitted {
				t.Errorf("HasUncommittedChanges() = %v, want %v", hasUncommitted, tt.hasUncommitted)
			}
		})
	}
}

// setupGitRepo initializes a git repository with user config.
func setupGitRepo(t *testing.T, dir string) error {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	return cmd.Run()
}

// commitFile creates a file, stages it, and commits it.
func commitFile(t *testing.T, dir, filename, content string) error {
	t.Helper()
	testhelpers.WriteFile(t, dir, filename, content)
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", "add "+filename)
	cmd.Dir = dir
	return cmd.Run()
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
