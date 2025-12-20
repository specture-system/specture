package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsGitRepository checks if the given directory is a git repository
// by checking for the existence of a .git directory.
func IsGitRepository(dir string) error {
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: %s", dir)
		}
		return fmt.Errorf("failed to check for git repository: %w", err)
	}
	return nil
}
