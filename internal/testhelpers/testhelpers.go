package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for tests and returns its path.
// The directory will be automatically cleaned up when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "specture-test-")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// WriteFile writes content to a file in the given directory.
func WriteFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return path
}

// ReadFile reads the contents of a file.
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(content)
}
