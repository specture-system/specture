package new

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/specture-system/specture/internal/fs"
	gitpkg "github.com/specture-system/specture/internal/git"
	specpkg "github.com/specture-system/specture/internal/spec"
)

// NewCommandContext holds information needed to create a new spec.
type NewCommandContext struct {
	WorkDir        string
	SpecsDir       string
	ParentRef      string
	ParentPath     string
	Title          string
	Author         string
	Number         int
	BranchName     string
	FileName       string
	RelativePath   string
	FilePath       string
	OriginalBranch string
}

// NewContext creates a new NewCommandContext for spec creation.
// It validates that the current directory is a git repository and returns an error if not.
func NewContext(workDir, title, parentRef string) (*NewCommandContext, error) {
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

	parentRef = strings.TrimSpace(parentRef)

	var parentPath string
	var parentInfo *specpkg.SpecInfo
	if parentRef != "" {
		var err error
		parentPath, err = specpkg.ResolvePath(specsDir, parentRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parent spec %q: %w", parentRef, err)
		}
		if filepath.Base(parentPath) != "SPEC.md" {
			return nil, fmt.Errorf("parent spec %q must be a SPEC.md spec", parentRef)
		}

		parentInfo, err = specpkg.Parse(parentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse parent spec: %w", err)
		}
	}

	// Find next spec number in the selected scope.
	number, err := FindNextSpecNumber(specsDir, parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find next spec number: %w", err)
	}

	// Get current branch
	currentBranch, err := gitpkg.GetCurrentBranch(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Convert title to slug
	slug := ToSlug(title)

	// Create branch name with date suffix
	date := time.Now().Format("2006-01-02")
	var branchName, fileName, relativePath, filePath string
	dirName := fmt.Sprintf("%03d-%s", number, slug)
	if parentPath == "" {
		branchName = fmt.Sprintf("spec/%s-%s", dirName, date)
		fileName = "SPEC.md"
		relativePath = filepath.Join(dirName, fileName)
		filePath = filepath.Join(specsDir, relativePath)
	} else {
		fullRef := parentInfo.FullRef + "." + strconv.Itoa(number)
		branchName = fmt.Sprintf("spec/%s-%s-%s", strings.ReplaceAll(fullRef, ".", "-"), slug, date)
		fileName = "SPEC.md"
		relativePath = filepath.Join(dirName, fileName)
		filePath = filepath.Join(filepath.Dir(parentPath), relativePath)
	}

	return &NewCommandContext{
		WorkDir:        workDir,
		SpecsDir:       specsDir,
		ParentRef:      parentRef,
		ParentPath:     parentPath,
		Title:          title,
		Author:         author,
		Number:         number,
		BranchName:     branchName,
		FileName:       fileName,
		RelativePath:   relativePath,
		FilePath:       filePath,
		OriginalBranch: currentBranch,
	}, nil
}

// CreateSpec creates the spec file. If body is provided, it's combined with generated frontmatter.
// If body is empty, the template's default content is used (frontmatter + placeholder).
// If noBranch is true, no git branch is created for the spec.
func (c *NewCommandContext) CreateSpec(dryRun bool, body string, noBranch bool) error {
	// Always generate frontmatter
	frontmatter := GenerateFrontmatter(c.Author)

	// Render body (either provided or default from template)
	if body == "" {
		var err error
		body, err = RenderDefaultBody(c.Title)
		if err != nil {
			return fmt.Errorf("failed to render body: %w", err)
		}
	}

	// Join frontmatter and body
	content := JoinSpecContent(frontmatter, body)

	if dryRun {
		fmt.Printf("Would create file: %s\n", c.FilePath)
		if !noBranch {
			fmt.Printf("Would create branch: %s\n", c.BranchName)
		}
		return nil
	}

	// Create the spec file using SafeWriteFile to prevent overwrites
	if err := fs.SafeWriteFile(c.FilePath, content); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	// Create branch unless --no-branch is set
	if !noBranch {
		if err := gitpkg.CreateBranch(c.WorkDir, c.BranchName); err != nil {
			// Clean up the created file if branch creation fails.
			if cleanupErr := os.Remove(c.FilePath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
				return fmt.Errorf("branch creation failed: %w (cleanup error: %v)", err, cleanupErr)
			}
			return err
		}
	}

	return nil
}

// Cleanup removes the spec file and deletes the branch, reverting to the original branch.
// This is called if the editor exits with a non-zero code (user cancellation).
// It handles both branch-based and non-branch specs.
func (c *NewCommandContext) Cleanup() error {
	// Remove the spec file. Nested-spec directories are left in place; repo
	// ignore rules keep stray scaffolding out of the working tree.
	if err := os.Remove(c.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove spec file: %w", err)
	}

	// Check if branch exists before trying to delete it
	// (it won't exist if --no-branch was used)
	branchExists, err := gitpkg.BranchExists(c.WorkDir, c.BranchName)
	if err != nil {
		return fmt.Errorf("failed to inspect cleanup branch %q: %w", c.BranchName, err)
	}

	if branchExists {
		// Checkout back to original branch
		if err := gitpkg.CheckoutBranch(c.WorkDir, c.OriginalBranch); err != nil {
			return fmt.Errorf("failed to checkout original branch: %w", err)
		}

		// Delete the spec branch
		if err := gitpkg.DeleteBranch(c.WorkDir, c.BranchName); err != nil {
			return fmt.Errorf("failed to delete branch: %w", err)
		}
	}

	return nil
}
