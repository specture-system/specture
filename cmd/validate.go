package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/specture-system/specture/internal/validate"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:     "validate [spec...]",
	Aliases: []string{"v"},
	Short:   "Validate specs",
	Long: `Validate checks that specs follow the Specture System guidelines.

It validates frontmatter, status, descriptions, and task lists.

Examples:
  specture validate                    # Validate all specs in specs/
  specture validate 000                # Validate spec by number
  specture validate specs/000-mvp.md   # Validate spec by path`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		specsDir := filepath.Join(cwd, "specs")

		// Determine which specs to validate
		var specPaths []string

		if len(args) == 0 {
			// Validate all specs
			paths, err := findAllSpecs(specsDir)
			if err != nil {
				return err
			}
			specPaths = paths
		} else {
			// Validate specific specs
			for _, arg := range args {
				path, err := resolveSpecPath(specsDir, arg)
				if err != nil {
					return err
				}
				specPaths = append(specPaths, path)
			}
		}

		if len(specPaths) == 0 {
			cmd.Println("No specs found to validate")
			return nil
		}

		// Validate each spec
		var validCount, invalidCount int
		var hasErrors bool

		for _, path := range specPaths {
			result, err := validate.ValidateSpecFile(path)
			if err != nil {
				cmd.PrintErrf("Error validating %s: %v\n", filepath.Base(path), err)
				hasErrors = true
				continue
			}

			cmd.Print(validate.FormatValidationResult(result))

			if result.IsValid() {
				validCount++
			} else {
				invalidCount++
				hasErrors = true
			}
		}

		// Print summary
		total := validCount + invalidCount
		cmd.Printf("\n%d of %d specs valid\n", validCount, total)

		if hasErrors {
			return fmt.Errorf("validation failed")
		}

		return nil
	},
}

// findAllSpecs finds all spec files in the specs directory
func findAllSpecs(specsDir string) ([]string, error) {
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("specs directory not found: %s", specsDir)
	}

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var paths []string
	specPattern := regexp.MustCompile(`^\d{3}-.*\.md$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if specPattern.MatchString(entry.Name()) {
			paths = append(paths, filepath.Join(specsDir, entry.Name()))
		}
	}

	return paths, nil
}

// resolveSpecPath resolves a spec argument to a file path
// Accepts:
//   - Full path: specs/000-mvp.md
//   - Just number: 000
//   - Number with name: 000-mvp
func resolveSpecPath(specsDir, arg string) (string, error) {
	// If it's already a path that exists, use it
	if _, err := os.Stat(arg); err == nil {
		return arg, nil
	}

	// Try to find by number prefix
	numberPattern := regexp.MustCompile(`^(\d{3})`)
	matches := numberPattern.FindStringSubmatch(arg)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid spec reference: %s (expected number like 000 or path)", arg)
	}

	number := matches[1]

	// Look for a file starting with that number
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if regexp.MustCompile(`^` + number + `-.*\.md$`).MatchString(entry.Name()) {
			return filepath.Join(specsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("spec not found: %s", arg)
}
