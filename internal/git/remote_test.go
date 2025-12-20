package git

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestGetRemoteURL(t *testing.T) {
	tests := []struct {
		name       string
		remoteName string
		setup      func(dir string) error
		want       string
		wantErr    string // empty string means no error expected
	}{
		{
			name:       "origin remote exists",
			remoteName: "origin",
			setup: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/user/repo.git")
				cmd.Dir = dir
				return cmd.Run()
			},
			want: "https://github.com/user/repo.git",
		},
		{
			name:       "no remotes",
			remoteName: "origin",
			setup: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir
				return cmd.Run()
			},
			want:    "",
			wantErr: "failed to get remote URL",
		},
		{
			name:       "multiple remotes, get origin",
			remoteName: "origin",
			setup: func(dir string) error {
				cmd := exec.Command("git", "init")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/user/repo.git")
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					return err
				}
				cmd = exec.Command("git", "remote", "add", "upstream", "https://github.com/other/repo.git")
				cmd.Dir = dir
				return cmd.Run()
			},
			want: "https://github.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got, err := GetRemoteURL(dir, tt.remoteName)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("GetRemoteURL() expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("GetRemoteURL() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("GetRemoteURL() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTerminology(t *testing.T) {
	tests := []struct {
		name  string
		forge Forge
		want  string
	}{
		{
			name:  "github contribution type",
			forge: ForgeGitHub,
			want:  "pull request",
		},
		{
			name:  "gitlab contribution type",
			forge: ForgeGitLab,
			want:  "merge request",
		},
		{
			name:  "unknown defaults to pull request",
			forge: ForgeUnknown,
			want:  "pull request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTerminology(tt.forge)
			if got != tt.want {
				t.Errorf("GetTerminology() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIdentifyForge(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      Forge
		wantErr   string // empty string means no error expected
	}{
		{
			name:      "github HTTPS",
			remoteURL: "https://github.com/user/repo.git",
			want:      ForgeGitHub,
		},
		{
			name:      "github SSH",
			remoteURL: "git@github.com:user/repo.git",
			want:      ForgeGitHub,
		},
		{
			name:      "gitlab HTTPS",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      ForgeGitLab,
		},
		{
			name:      "gitlab SSH",
			remoteURL: "git@gitlab.com:user/repo.git",
			want:      ForgeGitLab,
		},
		{
			name:      "custom gitlab instance",
			remoteURL: "https://my-gitlab.com/user/repo.git",
			want:      ForgeUnknown,
		},
		{
			name:      "unknown forge",
			remoteURL: "https://example.com/user/repo.git",
			want:      ForgeUnknown,
		},
		{
			name:      "SSH without .git suffix",
			remoteURL: "git@github.com:user/repo",
			want:      ForgeGitHub,
		},
		{
			name:      "invalid SSH URL format",
			remoteURL: "git@github.com",
			want:      ForgeUnknown,
			wantErr:   "invalid SSH URL format",
		},
		{
			name:      "invalid URL",
			remoteURL: "://invalid-url",
			want:      ForgeUnknown,
			wantErr:   "invalid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IdentifyForge(tt.remoteURL)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("IdentifyForge() expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("IdentifyForge() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("IdentifyForge() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("IdentifyForge() = %v, want %v", got, tt.want)
			}
		})
	}
}
