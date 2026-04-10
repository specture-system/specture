package setup

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	specpkg "github.com/specture-system/specture/internal/spec"
)

type specMovePlan struct {
	oldPath string
	newPath string
	linkOld string
	linkNew string
}

const specsGitignoreContent = "*\n!*/\n!**/SPEC.md\n!README.md\n"

// MigrateSpecsLayout moves flat top-level spec files into numbered spec directories,
// strips legacy frontmatter numbers from the moved files, and ensures specs/.gitignore
// keeps only SPEC.md and README.md files tracked.
func MigrateSpecsLayout(specsDir string, dryRun bool) (bool, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read specs directory: %w", err)
	}
	if os.IsNotExist(err) {
		entries = nil
	}

	var plans []specMovePlan
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if name == "README.md" || name == ".gitignore" || !strings.HasSuffix(name, ".md") {
			continue
		}

		oldPath := filepath.Join(specsDir, name)
		number, err := resolveSpecNumberForMigration(oldPath)
		if err != nil {
			return false, fmt.Errorf("failed to parse %s: %w", oldPath, err)
		}

		slug := specSlugFromFilename(name)
		newDir := fmt.Sprintf("%03d-%s", number, slug)
		newPath := filepath.Join(specsDir, newDir, "SPEC.md")

		if oldPath == newPath {
			continue
		}

		plans = append(plans, specMovePlan{
			oldPath: oldPath,
			newPath: newPath,
			linkOld: "/specs/" + name,
			linkNew: "specs/" + newDir + "/SPEC.md",
		})
	}

	changed := false
	for _, plan := range plans {
		changed = true
		if dryRun {
			fmt.Printf("[dry-run] Would move file: %s -> %s\n", plan.oldPath, plan.newPath)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(plan.newPath), 0o755); err != nil {
			return false, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(plan.newPath), err)
		}
		if err := os.Rename(plan.oldPath, plan.newPath); err != nil {
			return false, fmt.Errorf("failed to move %s to %s: %w", plan.oldPath, plan.newPath, err)
		}
		if err := stripLegacyNumberFromFrontmatter(plan.newPath); err != nil {
			return false, fmt.Errorf("failed to strip legacy number from %s: %w", plan.newPath, err)
		}
	}

	ignorePath := filepath.Join(specsDir, ".gitignore")
	if dryRun {
		fmt.Printf("[dry-run] Would create file: %s\n", ignorePath)
		changed = true
	} else {
		if err := os.WriteFile(ignorePath, []byte(specsGitignoreContent), 0o644); err != nil {
			return false, fmt.Errorf("failed to write specs .gitignore: %w", err)
		}
		if len(plans) > 0 {
			changed = true
		}
	}

	if len(plans) == 0 {
		changed = true
	}

	if dryRun {
		return changed, nil
	}

	if err := rewriteSpecLinks(specsDir, plans); err != nil {
		return false, err
	}

	return changed, nil
}

func specSlugFromFilename(filename string) string {
	slug := strings.TrimSuffix(filename, filepath.Ext(filename))
	re := regexp.MustCompile(`^\d+-`)
	return re.ReplaceAllString(slug, "")
}

func resolveSpecNumberForMigration(path string) (int, error) {
	info, err := specpkg.Parse(path)
	if err != nil {
		return 0, err
	}
	if info.Number >= 0 {
		return info.Number, nil
	}

	return legacyFrontmatterNumber(path)
}

func legacyFrontmatterNumber(path string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	inFrontmatter := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "number:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "number:"))
			number, err := strconv.Atoi(value)
			if err != nil {
				return 0, fmt.Errorf("invalid legacy spec number %q", value)
			}
			return number, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan frontmatter: %w", err)
	}

	return 0, fmt.Errorf("spec does not encode a number in path or frontmatter")
}

func rewriteSpecLinks(specsDir string, plans []specMovePlan) error {
	if len(plans) == 0 {
		return nil
	}

	var markdownFiles []string
	if err := filepath.WalkDir(specsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		markdownFiles = append(markdownFiles, path)
		return nil
	}); err != nil {
		return err
	}

	for _, path := range markdownFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s for link rewrite: %w", path, err)
		}

		updated := string(content)
		for _, plan := range plans {
			updated = strings.ReplaceAll(updated, plan.linkOld, plan.linkNew)
			oldFilename := filepath.Base(plan.oldPath)
			updated = strings.ReplaceAll(updated, "("+oldFilename+")", "("+plan.linkNew+")")
		}

		if updated != string(content) {
			if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("failed to write %s for link rewrite: %w", path, err)
			}
		}
	}

	return nil
}

func stripLegacyNumberFromFrontmatter(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	updated, changed, err := removeLegacyNumberFromFrontmatter(string(content))
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func removeLegacyNumberFromFrontmatter(content string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content, false, nil
	}

	var result []string
	inFrontmatter := false
	removed := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				result = append(result, line)
				result = append(result, lines[i+1:]...)
				break
			}
			inFrontmatter = true
			result = append(result, line)
			continue
		}

		if inFrontmatter && strings.HasPrefix(trimmed, "number:") {
			removed = true
			continue
		}

		result = append(result, line)
	}

	if !removed {
		return content, false, nil
	}

	return strings.Join(result, "\n"), true, nil
}

// MigrateSkillsDir moves .skills/specture/ to .agents/skills/specture/ if the
// old path exists and the new one doesn't. Returns true if migration occurred
// (or would occur in dry-run). After moving, removes .skills/ if empty.
func MigrateSkillsDir(workDir string, dryRun bool) (bool, error) {
	oldDir := filepath.Join(workDir, ".skills", "specture")
	newDir := filepath.Join(workDir, ".agents", "skills", "specture")

	// Check if old directory exists
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return false, nil
	}

	// Skip if new directory already exists
	if _, err := os.Stat(newDir); err == nil {
		return false, nil
	}

	if dryRun {
		fmt.Printf("[dry-run] Would migrate %s to %s\n", oldDir, newDir)
		return true, nil
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(newDir), 0755); err != nil {
		return false, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(newDir), err)
	}

	// Copy files from old to new (can't rename across directories reliably)
	if err := copyDir(oldDir, newDir); err != nil {
		return false, fmt.Errorf("failed to copy skills: %w", err)
	}

	// Remove old specture directory
	if err := os.RemoveAll(oldDir); err != nil {
		return false, fmt.Errorf("failed to remove old directory %s: %w", oldDir, err)
	}

	// Remove .skills/ if empty
	skillsDir := filepath.Join(workDir, ".skills")
	removeIfEmpty(skillsDir)

	return true, nil
}

// copyDir recursively copies src to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// removeIfEmpty removes a directory if it contains no entries.
func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	if len(entries) == 0 {
		os.Remove(dir)
	}
}
