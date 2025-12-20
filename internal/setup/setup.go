package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specture-system/specture/internal/fs"
	"github.com/specture-system/specture/internal/git"
	"github.com/specture-system/specture/internal/template"
)

// Context holds the setup context for the current repository.
type Context struct {
	WorkDir     string    // Current working directory
	Forge       git.Forge // Detected git forge
	Terminology string    // "pull request" or "merge request"
}

// NewContext creates a new setup context for the current directory.
func NewContext(cwd string) (*Context, error) {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Check if this is a git repository
	if err := git.IsGitRepository(cwd); err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	// Check for uncommitted changes
	hasChanges, err := git.HasUncommittedChanges(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to check git status: %w", err)
	}
	if hasChanges {
		return nil, fmt.Errorf("working tree has uncommitted changes")
	}

	// Detect forge
	var forge git.Forge
	remoteURL, err := git.GetRemoteURL(cwd, "origin")
	if err == nil && remoteURL != "" {
		forge, _ = git.IdentifyForge(remoteURL)
	}

	terminology := git.GetTerminology(forge)

	return &Context{
		WorkDir:     cwd,
		Forge:       forge,
		Terminology: terminology,
	}, nil
}

// CreateSpecsDirectory creates the specs/ directory in the current repository.
func (c *Context) CreateSpecsDirectory(dryRun bool) error {
	specsDir := fmt.Sprintf("%s/specs", c.WorkDir)

	if dryRun {
		fmt.Printf("[dry-run] Would create directory: %s\n", specsDir)
		return nil
	}

	if err := fs.EnsureDir(specsDir); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}

	return nil
}

// CreateSpecsReadme generates the specs/README.md file with forge-appropriate terminology.
func (c *Context) CreateSpecsReadme(dryRun bool) error {
	readmePath := filepath.Join(c.WorkDir, "specs", "README.md")

	// Render template with context
	content, err := template.RenderTemplate(SpecsReadmeTemplate, c)
	if err != nil {
		return fmt.Errorf("failed to render specs README template: %w", err)
	}

	if dryRun {
		fmt.Printf("[dry-run] Would create file: %s\n", readmePath)
		return nil
	}

	// Use WriteFile directly to overwrite if necessary (specs/README.md should always be updated)
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write specs README: %w", err)
	}

	return nil
}
