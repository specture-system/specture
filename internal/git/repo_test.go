package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestIsGitRepository(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) error
		wantErr string // empty string means no error expected
	}{
		{
			name: "valid git repo",
			setup: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, ".git"), 0755)
			},
		},
		{
			name: "not a git repo",
			setup: func(dir string) error {
				return nil
			},
			wantErr: "not a git repository",
		},
		{
			name: "git directory is a file",
			setup: func(dir string) error {
				_, err := os.Create(filepath.Join(dir, ".git"))
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := IsGitRepository(dir)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("IsGitRepository() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("IsGitRepository() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("IsGitRepository() unexpected error = %v", err)
			}
		})
	}
}

func TestIsGitRepositoryNotFound(t *testing.T) {
	err := IsGitRepository("/nonexistent/path")
	if err == nil {
		t.Errorf("IsGitRepository() expected error for nonexistent path")
		return
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("IsGitRepository() error = %v, want error containing %q", err, "not a git repository")
	}
}
