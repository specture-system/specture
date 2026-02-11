package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

	// Validate frontmatter exists
	if spec.Frontmatter == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "frontmatter",
			Message: "missing frontmatter",
		})
	} else {
		// Validate number field
		if spec.Frontmatter.Number == nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "number",
				Message: "missing required field",
			})
		} else if *spec.Frontmatter.Number < 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "number",
				Message: fmt.Sprintf("invalid value %d (must be a non-negative integer)", *spec.Frontmatter.Number),
			})
		} else {
			// Check number/filename mismatch
			filename := filepath.Base(spec.Path)
			fileNumRe := regexp.MustCompile(`^(\d{3})-`)
			matches := fileNumRe.FindStringSubmatch(filename)
			if len(matches) >= 2 {
				fileNum, err := strconv.Atoi(matches[1])
				if err == nil && fileNum != *spec.Frontmatter.Number {
					result.Warnings = append(result.Warnings, ValidationError{
						Field:   "number",
						Message: fmt.Sprintf("mismatch: frontmatter number %d does not match filename prefix %03d", *spec.Frontmatter.Number, fileNum),
					})
				}
			}
		}

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

	// Validate task list heading exists
	if !spec.HasTaskList {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "task list",
			Message: "missing '## Task List' heading",
		})
	}

	return result
}

// ValidateSpecs validates multiple specs, including cross-spec checks like duplicate numbers.
// Returns one ValidationResult per spec.
func ValidateSpecs(specs []*Spec) []*ValidationResult {
	results := make([]*ValidationResult, len(specs))
	for i, spec := range specs {
		results[i] = ValidateSpec(spec)
	}

	// Cross-spec: detect duplicate numbers
	numberToIdx := make(map[int][]int) // number -> indices of specs with that number
	for i, spec := range specs {
		if spec.Frontmatter != nil && spec.Frontmatter.Number != nil && *spec.Frontmatter.Number >= 0 {
			n := *spec.Frontmatter.Number
			numberToIdx[n] = append(numberToIdx[n], i)
		}
	}
	for num, indices := range numberToIdx {
		if len(indices) > 1 {
			for _, idx := range indices {
				results[idx].Errors = append(results[idx].Errors, ValidationError{
					Field:   "number",
					Message: fmt.Sprintf("duplicate number %d", num),
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
