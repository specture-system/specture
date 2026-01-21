package new

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/specture-system/specture/internal/template"
	"github.com/specture-system/specture/internal/templates"
)

// SpecData holds the template data for a new spec.
type SpecData struct {
	Title        string
	Author       string
	CreationDate string
}

// ToSlug converts a string to a URL-safe slug (kebab-case with special characters removed).
func ToSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.NewReplacer(" ", "-", "_", "-").Replace(s)

	// Remove any non-alphanumeric characters except hyphens
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	s = re.ReplaceAllString(s, "")

	// Replace multiple consecutive hyphens with a single hyphen
	re = regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	// Trim leading and trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// FindNextSpecNumber returns the next available spec number (e.g., 001 for a new spec).
// It searches the specs directory for existing specs and returns the next number.
func FindNextSpecNumber(specsDir string) (int, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		// If directory doesn't exist, start with 0
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Extract numbers from spec filenames (e.g., "000-name.md" -> 0)
	var numbers []int
	specFileRegex := regexp.MustCompile(`^(\d+)-`)

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			matches := specFileRegex.FindStringSubmatch(entry.Name())
			if len(matches) > 1 {
				num, err := strconv.Atoi(matches[1])
				if err == nil {
					numbers = append(numbers, num)
				}
			}
		}
	}

	// If no specs found, return 0
	if len(numbers) == 0 {
		return 0, nil
	}

	// Find the maximum number and return next
	sort.Ints(numbers)
	return numbers[len(numbers)-1] + 1, nil
}

// GenerateFrontmatter generates the YAML frontmatter for a spec.
func GenerateFrontmatter(title, author string) (string, error) {
	data := SpecData{
		Title:        title,
		Author:       author,
		CreationDate: time.Now().Format("2006-01-02"),
	}

	frontmatter := fmt.Sprintf(`---
status: draft
author: %s
creation_date: %s
---`, data.Author, data.CreationDate)

	return frontmatter, nil
}

// RenderDefaultBody renders the default body template from the spec template.
// It returns just the body portion (title, description, and task sections).
func RenderDefaultBody(title string) (string, error) {
	// The default body is everything after the frontmatter in the template.
	// We render the full template and extract just the body part.
	tmpl, err := templates.GetSpecTemplate()
	if err != nil {
		return "", err
	}

	data := SpecData{
		Title:        title,
		Author:       "", // Not needed for body-only rendering
		CreationDate: "", // Not needed for body-only rendering
	}

	content, err := template.RenderTemplate(tmpl, data)
	if err != nil {
		return "", err
	}

	// Extract body: everything after the frontmatter (second "---")
	lines := strings.Split(content, "\n")
	var bodyStart int
	var foundEnd bool

	for i, line := range lines {
		if line == "---" && i > 0 { // Skip first "---"
			bodyStart = i + 1
			foundEnd = true
			break
		}
	}

	if !foundEnd || bodyStart >= len(lines) {
		// If we can't find frontmatter, return the whole content
		return content, nil
	}

	// Join body lines and trim leading/trailing whitespace
	body := strings.Join(lines[bodyStart:], "\n")
	return strings.TrimSpace(body), nil
}

// RenderSpec renders a complete spec file from the template (frontmatter + default body).
// Kept for backward compatibility.
func RenderSpec(title, author string) (string, error) {
	tmpl, err := templates.GetSpecTemplate()
	if err != nil {
		return "", err
	}

	data := SpecData{
		Title:        title,
		Author:       author,
		CreationDate: time.Now().Format("2006-01-02"),
	}

	return template.RenderTemplate(tmpl, data)
}

// JoinSpecContent joins frontmatter and body into a complete spec.
func JoinSpecContent(frontmatter, body string) string {
	return frontmatter + "\n\n" + body
}
