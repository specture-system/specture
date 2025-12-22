package git

import (
	"fmt"
	"os/exec"
)

// CreateBranch creates a new git branch with the given name.
func CreateBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}
