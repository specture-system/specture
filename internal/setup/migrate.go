package setup

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// MigrationResult describes a spec that needs number added to frontmatter.
type MigrationResult struct {
	Path   string
	Number int
}

// FindSpecsNeedingMigration scans the specs directory for files matching NNN-slug.md
// that don't already have a number field in frontmatter.
func FindSpecsNeedingMigration(specsDir string) ([]MigrationResult, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	numericPrefix := regexp.MustCompile(`^(\d{3})-.*\.md$`)
	var results []MigrationResult

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := numericPrefix.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}

		num, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		path := filepath.Join(specsDir, entry.Name())
		hasNumber, err := hasNumberInFrontmatter(path)
		if err != nil {
			continue
		}

		if !hasNumber {
			results = append(results, MigrationResult{Path: path, Number: num})
		}
	}

	return results, nil
}

// AddNumberToFrontmatter adds a `number: N` field to the frontmatter of a spec file.
func AddNumberToFrontmatter(path string, number int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	updated, err := insertNumberInFrontmatter(content, number)
	if err != nil {
		return fmt.Errorf("failed to insert number: %w", err)
	}

	if err := os.WriteFile(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// hasNumberInFrontmatter checks if the file already has a number field in its frontmatter.
func hasNumberInFrontmatter(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if inFrontmatter {
				break // End of frontmatter
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "number:") {
			return true, nil
		}
	}

	return false, nil
}

// insertNumberInFrontmatter inserts `number: N` as the first field in the YAML frontmatter.
func insertNumberInFrontmatter(content []byte, number int) ([]byte, error) {
	lines := strings.Split(string(content), "\n")

	// Find first --- line
	for i, line := range lines {
		if line == "---" {
			// Insert number after the opening ---
			numberLine := fmt.Sprintf("number: %d", number)
			newLines := make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:i+1]...)
			newLines = append(newLines, numberLine)
			newLines = append(newLines, lines[i+1:]...)
			return []byte(strings.Join(newLines, "\n")), nil
		}
	}

	return nil, fmt.Errorf("no frontmatter found")
}
