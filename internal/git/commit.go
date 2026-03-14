package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// StageAll stages all tracked and untracked changes in the repository.
func StageAll(dir string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	return nil
}

// Commit creates a git commit with the provided message.
func Commit(dir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail != "" {
			return fmt.Errorf("failed to create commit: %w: %s", err, detail)
		}
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// PushBranch pushes a branch to origin and sets upstream tracking.
func PushBranch(dir, branchName string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %q: %w", branchName, err)
	}

	return nil
}
