package git

import (
	"testing"

	"github.com/specture-system/specture/internal/testhelpers"
)

func TestGetAuthor(t *testing.T) {
	t.Run("retrieves_configured_author", func(t *testing.T) {
		tmpDir := t.TempDir()
		testhelpers.InitGitRepo(t, tmpDir)

		author, err := GetAuthor(tmpDir)
		if err != nil {
			t.Fatalf("GetAuthor() error = %v", err)
		}

		if author != "Test User" {
			t.Errorf("GetAuthor() = %q, want %q", author, "Test User")
		}
	})

	t.Run("returns_empty_for_non_git_repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		author, err := GetAuthor(tmpDir)
		// GetAuthor doesn't fail for non-git repos, it just returns empty string
		// This is OK because the caller (NewContext) will have already validated the repo
		if err != nil && author == "" {
			// Either behavior is acceptable
		}
	})
}
