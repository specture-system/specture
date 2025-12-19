package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name  string
		setup func(dir string) string
	}{
		{
			name: "create new directory",
			setup: func(dir string) string {
				return filepath.Join(dir, "newdir")
			},
		},
		{
			name: "directory already exists",
			setup: func(dir string) string {
				path := filepath.Join(dir, "existing")
				os.Mkdir(path, 0755)
				return path
			},
		},
		{
			name: "create nested directories",
			setup: func(dir string) string {
				return filepath.Join(dir, "a", "b", "c")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			path := tt.setup(dir)
			err := EnsureDir(path)
			if err != nil {
				t.Errorf("EnsureDir() unexpected error: %v", err)
				return
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Errorf("directory was not created: %v", err)
			}
			if !info.IsDir() {
				t.Errorf("path is not a directory")
			}
		})
	}
}

func TestSafeWriteFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		setup    func(dir string) error
		wantErr  string // empty string means no error expected
	}{
		{
			name:     "write new file",
			filename: "test.txt",
			content:  "hello world",
			setup:    func(dir string) error { return nil },
		},
		{
			name:     "file already exists",
			filename: "existing.txt",
			content:  "new content",
			wantErr:  "file already exists",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("old"), 0644)
			},
		},
		{
			name:     "create nested file",
			filename: "a/b/c.txt",
			content:  "nested",
			setup:    func(dir string) error { return nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			if err := tt.setup(dir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			path := filepath.Join(dir, tt.filename)
			err := SafeWriteFile(path, tt.content)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("SafeWriteFile() expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("SafeWriteFile() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("SafeWriteFile() unexpected error: %v", err)
				return
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("failed to read written file: %v", err)
			}
			if string(content) != tt.content {
				t.Errorf("file content mismatch: got %q, want %q", string(content), tt.content)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) string
		wantErr string // empty string means no error expected
	}{
		{
			name: "file exists",
			setup: func(dir string) string {
				path := filepath.Join(dir, "test.txt")
				os.WriteFile(path, []byte("content"), 0644)
				return path
			},
		},
		{
			name: "file does not exist",
			setup: func(dir string) string {
				return filepath.Join(dir, "missing.txt")
			},
			wantErr: "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			path := tt.setup(dir)
			err := FileExists(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("FileExists() expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("FileExists() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("FileExists() unexpected error: %v", err)
			}
		})
	}
}
