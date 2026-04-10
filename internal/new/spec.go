package new

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	specpkg "github.com/specture-system/specture/internal/spec"
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

// FindNextSpecNumber returns the next available spec number in a scope.
// With no parentPath, it allocates from the top-level specs under specsDir.
// With a parentPath, it allocates from that parent spec's immediate children.
func FindNextSpecNumber(specsDir, parentPath string) (int, error) {
	// Check if directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return 0, nil
	}

	specs, err := specpkg.FindSpecsInScope(specsDir, parentPath)
	if err != nil {
		return 0, err
	}

	maxNumber := -1
	for _, info := range specs {
		if info.Number > maxNumber {
			maxNumber = info.Number
		}
	}

	if maxNumber < 0 {
		return 0, nil
	}

	return maxNumber + 1, nil
}

// GenerateFrontmatter generates the YAML frontmatter for a spec.
func GenerateFrontmatter(author string) string {
	frontmatter := fmt.Sprintf(`---
status: draft
author: %s
creation_date: %s
---`, author, time.Now().Format("2006-01-02"))

	return frontmatter
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
