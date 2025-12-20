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
)

// SpecData holds the template data for a new spec.
type SpecData struct {
	Title        string
	Author       string
	CreationDate string
}

// ToKebabCase converts a string to kebab-case.
func ToKebabCase(s string) string {
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

// GetSpecTemplate returns the spec file template.
func GetSpecTemplate() (string, error) {
	content, err := templateFiles.ReadFile("templates/spec-template.md")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// RenderSpec renders a spec file from the template with the given data.
func RenderSpec(title, author string) (string, error) {
	tmpl, err := GetSpecTemplate()
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
