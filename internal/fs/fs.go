package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates a directory if it doesn't already exist.
// It creates all necessary parent directories.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// SafeWriteFile writes content to a file, but fails if the file already exists.
// It creates necessary parent directories.
func SafeWriteFile(path, content string) error {
	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check file: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FileExists returns nil if the file exists, or an error if it doesn't.
func FileExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to check file: %w", err)
	}
	return nil
}
