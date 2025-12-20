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
	// If the repository has no remote or remote URL detection fails, forge remains Unknown.
	// This is intentional: we gracefully default to "pull request" terminology rather than
	// prompting the user. This keeps setup non-interactive and suitable for automation.
	var forge git.Forge
	remoteURL, err := git.GetRemoteURL(cwd, "origin")
	if err == nil && remoteURL != "" {
		// IdentifyForge returns Unknown if URL doesn't match known forges; that's OK.
		// We'll use generic terminology in that case.
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
	specsDir := filepath.Join(c.WorkDir, "specs")

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

// FindExistingFiles finds AGENTS.md and CLAUDE.md files in the repository.
func (c *Context) FindExistingFiles() (hasAgentsFile, hasClaudeFile bool) {
	agentsPath := filepath.Join(c.WorkDir, "AGENTS.md")
	claudePath := filepath.Join(c.WorkDir, "CLAUDE.md")

	if _, err := os.Stat(agentsPath); err == nil {
		hasAgentsFile = true
	}
	if _, err := os.Stat(claudePath); err == nil {
		hasClaudeFile = true
	}

	return hasAgentsFile, hasClaudeFile
}

// RenderAgentPromptTemplate renders the agent prompt template with context.
func RenderAgentPromptTemplate(isClaudeFile bool) (string, error) {
	return template.RenderTemplate(AgentPromptTemplate, map[string]bool{"IsClaudeFile": isClaudeFile})
}
