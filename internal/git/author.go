package git

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
