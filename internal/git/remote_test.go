package git

import (
	"os/exec"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestGetRemoteURL(t *testing.T) {
	tests := []struct {
		name       string
		remoteName string
		setup      func(dir string) error
		want       string
		wantErr    bool
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
			want:    "https://github.com/user/repo.git",
			wantErr: false,
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
			wantErr: true,
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
			want:    "https://github.com/user/repo.git",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got, err := GetRemoteURL(dir, tt.remoteName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentifyForge(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      Forge
		wantErr   bool
	}{
		{
			name:      "github HTTPS",
			remoteURL: "https://github.com/user/repo.git",
			want:      ForgeGitHub,
			wantErr:   false,
		},
		{
			name:      "github SSH",
			remoteURL: "git@github.com:user/repo.git",
			want:      ForgeGitHub,
			wantErr:   false,
		},
		{
			name:      "gitlab HTTPS",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      ForgeGitLab,
			wantErr:   false,
		},
		{
			name:      "gitlab SSH",
			remoteURL: "git@gitlab.com:user/repo.git",
			want:      ForgeGitLab,
			wantErr:   false,
		},
		{
			name:      "custom gitlab instance",
			remoteURL: "https://my-gitlab.com/user/repo.git",
			want:      ForgeUnknown,
			wantErr:   false,
		},
		{
			name:      "unknown forge",
			remoteURL: "https://example.com/user/repo.git",
			want:      ForgeUnknown,
			wantErr:   false,
		},
		{
			name:      "SSH without .git suffix",
			remoteURL: "git@github.com:user/repo",
			want:      ForgeGitHub,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IdentifyForge(tt.remoteURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyForge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IdentifyForge() = %v, want %v", got, tt.want)
			}
		})
	}
}
