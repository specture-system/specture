package new

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetAuthor retrieves the git user name from git config.
func GetAuthor(dir string) (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user name: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateBranch creates a new git branch with the given name.
func CreateBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

// CommitChanges stages and commits files with the given message.
func CommitChanges(dir, message string, files ...string) error {
	// Stage files
	addCmd := exec.Command("git", append([]string{"add"}, files...)...)
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

// PushChanges pushes the current branch to origin.
func PushChanges(dir, branchName string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}
	return nil
}
