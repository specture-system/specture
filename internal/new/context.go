package new

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/specture-system/specture/internal/fs"
	gitpkg "github.com/specture-system/specture/internal/git"
)

// NewCommandContext holds information needed to create a new spec.
type NewCommandContext struct {
	WorkDir    string
	SpecsDir   string
	Title      string
	Author     string
	Number     int
	BranchName string
	FileName   string
	FilePath   string
}

// NewContext creates a new NewCommandContext for spec creation.
// It validates that the current directory is a git repository and returns an error if not.
func NewContext(workDir, title string) (*NewCommandContext, error) {
	// Validate git repository
	if err := gitpkg.IsGitRepository(workDir); err != nil {
		return nil, err
	}

	// Check for uncommitted changes
	hasDirty, err := gitpkg.HasUncommittedChanges(workDir)
	if err != nil {
		return nil, err
	}
	if hasDirty {
		return nil, fmt.Errorf("repository has uncommitted changes; please commit or stash them first")
	}

	// Get author from git config
	author, err := gitpkg.GetAuthor(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get git author: %w", err)
	}

	// Find specs directory
	specsDir := filepath.Join(workDir, "specs")

	// Ensure specs directory exists
	if err := fs.EnsureDir(specsDir); err != nil {
		return nil, fmt.Errorf("failed to ensure specs directory exists: %w", err)
	}

	// Find next spec number
	number, err := FindNextSpecNumber(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find next spec number: %w", err)
	}

	// Convert title to slug
	slug := ToSlug(title)

	// Create branch name with date suffix
	date := time.Now().Format("2006-01-02")
	branchName := fmt.Sprintf("spec/%03d-%s-%s", number, slug, date)

	// Create file name and path
	fileName := fmt.Sprintf("%03d-%s.md", number, slug)
	filePath := filepath.Join(specsDir, fileName)

	return &NewCommandContext{
		WorkDir:    workDir,
		SpecsDir:   specsDir,
		Title:      title,
		Author:     author,
		Number:     number,
		BranchName: branchName,
		FileName:   fileName,
		FilePath:   filePath,
	}, nil
}

// CreateSpec creates the spec file and opens it in the editor.
func (c *NewCommandContext) CreateSpec(dryRun bool) error {
	// Render spec from template
	content, err := RenderSpec(c.Title, c.Author)
	if err != nil {
		return fmt.Errorf("failed to render spec: %w", err)
	}

	if dryRun {
		fmt.Printf("Would create file: %s\n", c.FilePath)
		fmt.Printf("Would create branch: %s\n", c.BranchName)
		return nil
	}

	// Create the spec file using SafeWriteFile to prevent overwrites
	if err := fs.SafeWriteFile(c.FilePath, content); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	// Create branch
	if err := gitpkg.CreateBranch(c.WorkDir, c.BranchName); err != nil {
		// Clean up the file if branch creation fails
		os.Remove(c.FilePath)
		return err
	}

	return nil
}

// Cleanup removes the spec file and deletes the branch, reverting to the previous branch.
// This is called if the editor exits with a non-zero code (user cancellation).
func (c *NewCommandContext) Cleanup() error {
	// Remove the spec file
	if err := os.Remove(c.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove spec file: %w", err)
	}

	// Delete the branch by checking out main/master and deleting the branch
	// Try main first, then master
	for _, targetBranch := range []string{"main", "master"} {
		checkoutCmd := exec.Command("git", "checkout", targetBranch)
		checkoutCmd.Dir = c.WorkDir
		if err := checkoutCmd.Run(); err == nil {
			// Successfully checked out, now delete the branch
			deleteCmd := exec.Command("git", "branch", "-D", c.BranchName)
			deleteCmd.Dir = c.WorkDir
			if err := deleteCmd.Run(); err != nil {
				return fmt.Errorf("failed to delete branch: %w", err)
			}
			return nil
		}
	}

	// If we can't find main or master, just try to delete the branch anyway
	deleteCmd := exec.Command("git", "branch", "-D", c.BranchName)
	deleteCmd.Dir = c.WorkDir
	if err := deleteCmd.Run(); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}
