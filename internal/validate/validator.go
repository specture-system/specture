package validate

import (
	"fmt"
	"path/filepath"
	"slices"
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
	Path   string
	Errors []ValidationError
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
	if result.IsValid() {
		return fmt.Sprintf("✓ %s\n", filename)
	}

	output := fmt.Sprintf("✗ %s\n", filename)
	for _, err := range result.Errors {
		output += fmt.Sprintf("  - %s: %s\n", err.Field, err.Message)
	}
	return output
}
