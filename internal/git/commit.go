package git

import (
	"fmt"
	"os/exec"
)

// CommitChanges stages and commits files with the given message.
func CommitChanges(dir, message string, files ...string) error {
	// Stage files
	addArgs := append([]string{"add"}, files...)
	addCmd := exec.Command("git", addArgs...)
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = dir
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
