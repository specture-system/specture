package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestIsGitRepository(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) error
		wantErr bool
	}{
		{
			name: "valid git repo",
			setup: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, ".git"), 0755)
			},
			wantErr: false,
		},
		{
			name: "not a git repo",
			setup: func(dir string) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "git directory is a file",
			setup: func(dir string) error {
				_, err := os.Create(filepath.Join(dir, ".git"))
				return err
			},
			wantErr: false, // We're lenient - if .git exists (even as a file), consider it a git repo
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			err := IsGitRepository(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsGitRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsGitRepositoryNotFound(t *testing.T) {
	err := IsGitRepository("/nonexistent/path")
	if err == nil {
		t.Errorf("IsGitRepository() expected error for nonexistent path")
	}
}
