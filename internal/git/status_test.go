package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestHasUncommittedChanges(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(dir string) error
		hasUncommitted bool
		wantErr        bool
	}{
		{
			name: "clean working tree",
			setup: func(dir string) error {
				// Initialize git repo and commit a file
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
				if err := cmd.Run(); err != nil {
					return err
				}
				testhelpers.WriteFile(t, dir, "test.txt", "content")
				cmd = exec.Command("git", "add", "test.txt")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "commit", "-m", "initial")
				cmd.Dir = dir
				return cmd.Run()
			},
			hasUncommitted: false,
			wantErr:        false,
		},
		{
			name: "uncommitted changes",
			setup: func(dir string) error {
				// Initialize git repo with a file, then modify it
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
				if err := cmd.Run(); err != nil {
					return err
				}
				testhelpers.WriteFile(t, dir, "test.txt", "content")
				cmd = exec.Command("git", "add", "test.txt")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "commit", "-m", "initial")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				// Now modify the file
				return os.WriteFile(filepath.Join(dir, "test.txt"), []byte("modified"), 0644)
			},
			hasUncommitted: true,
			wantErr:        false,
		},
		{
			name: "untracked files",
			setup: func(dir string) error {
				// Initialize git repo and add an untracked file
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
				if err := cmd.Run(); err != nil {
					return err
				}
				testhelpers.WriteFile(t, dir, "tracked.txt", "content")
				cmd = exec.Command("git", "add", "tracked.txt")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "commit", "-m", "initial")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				// Add untracked file
				_, err := os.Create(filepath.Join(dir, "untracked.txt"))
				return err
			},
			hasUncommitted: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			hasUncommitted, err := HasUncommittedChanges(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasUncommittedChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hasUncommitted != tt.hasUncommitted {
				t.Errorf("HasUncommittedChanges() = %v, want %v", hasUncommitted, tt.hasUncommitted)
			}
		})
	}
}
