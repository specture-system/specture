package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// HasUncommittedChanges checks if the git repository in the given directory
// has uncommitted changes (modified files or untracked files).
func HasUncommittedChanges(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return len(output) > 0, nil
}

// ChangedFiles returns the set of changed files in the working tree based on
// `git status --porcelain`, including modified, staged, and untracked files.
func ChangedFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list changed files: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	seen := map[string]struct{}{}
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		pathField := strings.TrimSpace(line[3:])
		if pathField == "" {
			continue
		}

		// Rename/copy records use "old -> new"; prefer the destination path.
		if idx := strings.LastIndex(pathField, " -> "); idx >= 0 {
			pathField = strings.TrimSpace(pathField[idx+4:])
		}

		if _, ok := seen[pathField]; ok {
			continue
		}
		seen[pathField] = struct{}{}
		files = append(files, pathField)
	}

	sort.Strings(files)
	return files, nil
}
