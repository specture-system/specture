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

// CheckoutBranch checks out an existing branch.
func CheckoutBranch(dir, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a local branch.
func DeleteBranch(dir, branchName string) error {
	cmd := exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}
