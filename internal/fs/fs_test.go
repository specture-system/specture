package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) string
		wantErr bool
	}{
		{
			name: "create new directory",
			setup: func(dir string) string {
				return filepath.Join(dir, "newdir")
			},
			wantErr: false,
		},
		{
			name: "directory already exists",
			setup: func(dir string) string {
				path := filepath.Join(dir, "existing")
				os.Mkdir(path, 0755)
				return path
			},
			wantErr: false,
		},
		{
			name: "create nested directories",
			setup: func(dir string) string {
				return filepath.Join(dir, "a", "b", "c")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			path := tt.setup(dir)
			err := EnsureDir(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("directory was not created: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("path is not a directory")
				}
			}
		})
	}
}

func TestSafeWriteFile(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		content   string
		setup     func(dir string) error
		wantErr   bool
		shouldErr string // error substring
	}{
		{
			name:     "write new file",
			filename: "test.txt",
			content:  "hello world",
			setup:    func(dir string) error { return nil },
			wantErr:  false,
		},
		{
			name:      "file already exists",
			filename:  "existing.txt",
			content:   "new content",
			wantErr:   true,
			shouldErr: "already exists",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("old"), 0644)
			},
		},
		{
			name:     "create nested file",
			filename: "a/b/c.txt",
			content:  "nested",
			setup:    func(dir string) error { return nil },
			wantErr:  false,
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
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeWriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.shouldErr != "" {
				if !contains(err.Error(), tt.shouldErr) {
					t.Errorf("SafeWriteFile() error = %v, want error containing %q", err, tt.shouldErr)
				}
			}
			if err == nil {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("failed to read written file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("file content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(dir string) string
		wantErr bool
	}{
		{
			name: "file exists",
			setup: func(dir string) string {
				path := filepath.Join(dir, "test.txt")
				os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			wantErr: false,
		},
		{
			name: "file does not exist",
			setup: func(dir string) string {
				return filepath.Join(dir, "missing.txt")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := testhelpers.TempDir(t)
			path := tt.setup(dir)
			err := FileExists(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
