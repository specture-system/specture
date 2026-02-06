// Package spec provides shared spec parsing, discovery, and querying.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// SpecInfo represents a parsed spec file with all extracted metadata.
type SpecInfo struct {
	Path               string
	Name               string
	Number             int
	Status             string
	CurrentTask        string
	CurrentTaskSection string
	CompleteTasks      []Task
	IncompleteTasks    []Task
}

// Task represents a single task item from a spec's task list.
type Task struct {
	Text     string
	Complete bool
	Section  string
}

// FindAll finds all spec files in the given specs directory.
// Spec files match the pattern NNN-name.md (3-digit prefix).
func FindAll(specsDir string) ([]string, error) {
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("specs directory not found: %s", specsDir)
	}

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var paths []string
	specPattern := regexp.MustCompile(`^\d{3}-.*\.md$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if specPattern.MatchString(entry.Name()) {
			paths = append(paths, filepath.Join(specsDir, entry.Name()))
		}
	}

	return paths, nil
}

// ResolvePath resolves a spec argument to a file path.
// Accepts:
//   - Full path: specs/000-mvp.md
//   - Just number with or without leading zeros: 0, 00, 000
func ResolvePath(specsDir, arg string) (string, error) {
	// If it's already a path that exists, use it
	if _, err := os.Stat(arg); err == nil {
		return arg, nil
	}

	// Try to find by number - accept 1-3 digit numbers
	numberPattern := regexp.MustCompile(`^(\d{1,3})$`)
	matches := numberPattern.FindStringSubmatch(arg)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid spec reference: %s (expected number like 0, 00, or 000)", arg)
	}

	// Pad number to 3 digits with leading zeros
	number := fmt.Sprintf("%03s", matches[1])

	// Look for a file starting with that number
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if regexp.MustCompile(`^` + number + `-.*\.md$`).MatchString(entry.Name()) {
			return filepath.Join(specsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("spec not found: %s", arg)
}
