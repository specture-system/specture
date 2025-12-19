package git

import (
	"fmt"
	"os/exec"
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
