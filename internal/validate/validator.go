package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	markdownSectionPattern = regexp.MustCompile(`^(#{2,6})\s+(.+)$`)
	numberedSectionPattern = regexp.MustCompile(`^\d+(?:(?:\.\d+)+|[.)]|\s)`)
	specDirPrefixPattern   = regexp.MustCompile(`^(\d+)`)
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains the results of validating a spec
type ValidationResult struct {
	Path     string
	Errors   []ValidationError
	Warnings []ValidationError
}

// IsValid returns true if there are no validation errors
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateSpec validates a spec and returns the validation result
func ValidateSpec(spec *Spec) *ValidationResult {
	result := &ValidationResult{
		Path:   spec.Path,
		Errors: []ValidationError{},
	}

	if fullRefFromPath(spec.Path) == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "path",
			Message: "spec path must encode a numbered ref",
		})
	}

	// Validate frontmatter exists
	if spec.Frontmatter == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "frontmatter",
			Message: "missing frontmatter",
		})
	} else {
		// Validate status field
		if spec.Frontmatter.Status == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "status",
				Message: "missing required field",
			})
		} else if !slices.Contains(ValidStatus, spec.Frontmatter.Status) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "status",
				Message: fmt.Sprintf("invalid value %q (must be one of: draft, approved, in-progress, completed, rejected)", spec.Frontmatter.Status),
			})
		}
	}

	// Validate title (H1 heading) exists
	if spec.Title == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "title",
			Message: "missing H1 heading",
		})
	}

	if numberedHeading, ok := firstNumberedSectionHeading(spec.Source); ok {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "headings",
			Message: fmt.Sprintf("section headers must not be numbered (found %q)", numberedHeading),
		})
	}

	return result
}

func firstNumberedSectionHeading(source []byte) (string, bool) {
	lines := strings.Split(string(source), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		matches := markdownSectionPattern.FindStringSubmatch(trimmed)
		if len(matches) != 3 {
			continue
		}
		title := strings.TrimSpace(matches[2])
		if numberedSectionPattern.MatchString(title) {
			return trimmed, true
		}
	}

	return "", false
}

// ValidateSpecs validates multiple specs, including cross-spec checks like duplicate full refs.
// Returns one ValidationResult per spec.
func ValidateSpecs(specs []*Spec) []*ValidationResult {
	results := make([]*ValidationResult, len(specs))
	for i, spec := range specs {
		results[i] = ValidateSpec(spec)
	}

	// Cross-spec: detect duplicate full refs.
	refToIdx := make(map[string][]int)
	for i, spec := range specs {
		fullRef := fullRefFromPath(spec.Path)
		if fullRef != "" {
			refToIdx[fullRef] = append(refToIdx[fullRef], i)
		}
	}
	for fullRef, indices := range refToIdx {
		if len(indices) > 1 {
			for _, idx := range indices {
				results[idx].Errors = append(results[idx].Errors, ValidationError{
					Field:   "fullref",
					Message: fmt.Sprintf("duplicate ref %s", fullRef),
				})
			}
		}
	}

	return results
}

// ValidateSpecFile parses and validates a spec file
func ValidateSpecFile(path string) (*ValidationResult, error) {
	spec, err := ParseSpec(path)
	if err != nil {
		return nil, err
	}
	return ValidateSpec(spec), nil
}

// FormatValidationResult formats a validation result for display
func FormatValidationResult(result *ValidationResult) string {
	filename := filepath.Base(result.Path)
	if result.IsValid() && len(result.Warnings) == 0 {
		return fmt.Sprintf("✓ %s\n", filename)
	}

	var output string
	if result.IsValid() {
		output = fmt.Sprintf("✓ %s\n", filename)
	} else {
		output = fmt.Sprintf("✗ %s\n", filename)
	}
	for _, err := range result.Errors {
		output += fmt.Sprintf("  - %s: %s\n", err.Field, err.Message)
	}
	for _, w := range result.Warnings {
		output += fmt.Sprintf("  ⚠ %s: %s\n", w.Field, w.Message)
	}
	return output
}

func fullRefFromPath(path string) string {
	cleaned := filepath.Clean(path)
	parts := strings.Split(cleaned, string(filepath.Separator))

	specsIdx := -1
	for i, part := range parts {
		if part == "specs" {
			specsIdx = i
		}
	}
	if specsIdx < 0 || specsIdx+1 >= len(parts)-1 {
		return ""
	}

	var refs []string
	for _, part := range parts[specsIdx+1 : len(parts)-1] {
		number := extractLeadingNumber(part)
		if number < 0 {
			return ""
		}
		refs = append(refs, fmt.Sprintf("%d", number))
	}

	return strings.Join(refs, ".")
}

func extractLeadingNumber(value string) int {
	matches := specDirPrefixPattern.FindStringSubmatch(value)
	if len(matches) != 2 {
		return -1
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}
	return num
}
