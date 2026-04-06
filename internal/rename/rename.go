// Package rename handles renaming spec directories and updating cross-references.
package rename

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	specpkg "github.com/specture-system/specture/internal/spec"
)

// RenameResult describes what a rename operation will do.
type RenameResult struct {
	OldPath     string
	NewPath     string
	LinkUpdates []LinkUpdate
}

// LinkUpdate describes a markdown link update in a file.
type LinkUpdate struct {
	File    string
	OldLink string
	NewLink string
}

// Plan creates a rename plan for a spec without executing it.
func Plan(specsDir string, specNumber int, newSlug string) (*RenameResult, error) {
	// Find the spec directory by number.
	oldPath, err := specpkg.ResolvePath(specsDir, fmt.Sprintf("%d", specNumber))
	if err != nil {
		return nil, err
	}
	if filepath.Base(oldPath) != "SPEC.md" {
		return nil, fmt.Errorf("spec %d must resolve to a SPEC.md spec", specNumber)
	}

	if newSlug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	oldDir := filepath.Dir(oldPath)
	parentDir := filepath.Dir(oldDir)
	newDirName := fmt.Sprintf("%03d-%s", specNumber, newSlug)
	newDir := filepath.Join(parentDir, newDirName)
	newPath := filepath.Join(newDir, "SPEC.md")

	// Don't rename if already the same.
	if oldPath == newPath {
		return nil, fmt.Errorf("spec is already named %s", newDirName)
	}

	// Check target directory doesn't already exist.
	if _, err := os.Stat(newDir); err == nil {
		return nil, fmt.Errorf("target spec directory already exists: %s", newDirName)
	}

	// Find all markdown link references to the old spec path in specs/.
	linkUpdates, err := findLinkReferences(specsDir, oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for link references: %w", err)
	}

	return &RenameResult{
		OldPath:     oldPath,
		NewPath:     newPath,
		LinkUpdates: linkUpdates,
	}, nil
}

// Execute performs a rename operation described by the result.
func Execute(result *RenameResult) error {
	oldDir := filepath.Dir(result.OldPath)
	newDir := filepath.Dir(result.NewPath)

	// Rename the spec directory.
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename spec directory: %w", err)
	}

	// Update links in other files
	for _, update := range result.LinkUpdates {
		content, err := os.ReadFile(update.File)
		if err != nil {
			return fmt.Errorf("failed to read %s for link update: %w", update.File, err)
		}

		updated := strings.ReplaceAll(string(content), update.OldLink, update.NewLink)
		if err := os.WriteFile(update.File, []byte(updated), 0644); err != nil {
			return fmt.Errorf("failed to write %s for link update: %w", update.File, err)
		}
	}

	return nil
}

// stripNumericPrefix removes the NNN- prefix from a filename, keeping the .md extension.
// Returns just the slug with .md extension.
func stripNumericPrefix(filename string) string {
	re := regexp.MustCompile(`^\d{3}-`)
	return re.ReplaceAllString(filename, "")
}

// findLinkReferences scans all markdown files in specsDir for links referencing the old spec path
// and returns the necessary updates.
func findLinkReferences(specsDir, oldPath, newPath string) ([]LinkUpdate, error) {
	oldRel, err := filepath.Rel(specsDir, oldPath)
	if err != nil {
		return nil, err
	}
	newRel, err := filepath.Rel(specsDir, newPath)
	if err != nil {
		return nil, err
	}
	oldRel = filepath.ToSlash(oldRel)
	newRel = filepath.ToSlash(newRel)

	oldLinks := []string{
		oldRel,
		"specs/" + oldRel,
		"/specs/" + oldRel,
	}
	newLinks := []string{
		newRel,
		"specs/" + newRel,
		"/specs/" + newRel,
	}

	var updates []LinkUpdate
	if err := filepath.WalkDir(specsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)

		for i, oldLink := range oldLinks {
			if strings.Contains(contentStr, oldLink) {
				updates = append(updates, LinkUpdate{
					File:    path,
					OldLink: oldLink,
					NewLink: newLinks[i],
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return updates, nil
}
