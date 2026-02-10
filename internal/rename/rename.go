// Package rename handles renaming spec files and updating cross-references.
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
	// Find the spec file by number
	oldPath, err := specpkg.ResolvePath(specsDir, fmt.Sprintf("%d", specNumber))
	if err != nil {
		return nil, err
	}

	oldFilename := filepath.Base(oldPath)

	// Determine new slug if not provided
	if newSlug == "" {
		// Strip numeric prefix from current filename
		newSlug = stripNumericPrefix(oldFilename)
	}

	newFilename := newSlug
	if !strings.HasSuffix(newFilename, ".md") {
		newFilename += ".md"
	}
	newPath := filepath.Join(specsDir, newFilename)

	// Don't rename if already the same
	if oldPath == newPath {
		return nil, fmt.Errorf("file is already named %s", newFilename)
	}

	// Check target doesn't already exist
	if _, err := os.Stat(newPath); err == nil {
		return nil, fmt.Errorf("target file already exists: %s", newFilename)
	}

	// Find all markdown link references to the old filename in specs/
	linkUpdates, err := findLinkReferences(specsDir, oldFilename, newFilename)
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
	// Rename the file
	if err := os.Rename(result.OldPath, result.NewPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
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

// findLinkReferences scans all markdown files in specsDir for links referencing the old filename
// and returns the necessary updates.
func findLinkReferences(specsDir, oldFilename, newFilename string) ([]LinkUpdate, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}

	// Patterns to match: [text](/specs/old.md) or [text](old.md)
	oldLinkAbs := "/specs/" + oldFilename
	newLinkAbs := "/specs/" + newFilename

	var updates []LinkUpdate
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(specsDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Check for absolute path references: [text](/specs/old.md)
		if strings.Contains(contentStr, oldLinkAbs) {
			updates = append(updates, LinkUpdate{
				File:    filePath,
				OldLink: oldLinkAbs,
				NewLink: newLinkAbs,
			})
		}

		// Check for relative references: [text](old.md)
		if strings.Contains(contentStr, "("+oldFilename+")") {
			updates = append(updates, LinkUpdate{
				File:    filePath,
				OldLink: "(" + oldFilename + ")",
				NewLink: "(" + newFilename + ")",
			})
		}
	}

	return updates, nil
}
