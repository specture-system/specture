package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/specture-system/specture/internal/validate"
	"github.com/spf13/cobra"
)

var specFlag string

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v"},
	Short:   "Validate specs",
	Long: `Validate checks that specs follow the Specture System guidelines.

It validates frontmatter, status, descriptions, and task lists.

Examples:
  specture validate              # Validate all specs in specs/
  specture validate --spec 0     # Validate spec 000-*.md by number
  specture validate -s 42        # Short form, validates spec 042-*.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		invalidCount, err := runValidate(cmd, args)
		if err != nil {
			return err
		}

		// Exit with non-zero status if any specs failed validation
		if invalidCount > 0 {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&specFlag, "spec", "s", "", "Spec number to validate (e.g., 0, 00, or 000)")
}

// runValidate performs validation and returns the count of invalid specs.
// Separated from the command for testability.
func runValidate(cmd *cobra.Command, args []string) (invalidCount int, err error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return 0, fmt.Errorf("failed to get current directory: %w", err)
	}

	specsDir := filepath.Join(cwd, "specs")

	// Get spec flag value
	spec, _ := cmd.Flags().GetString("spec")

	// Determine which specs to validate
	var specPaths []string

	if spec == "" {
		// Validate all specs
		paths, err := specpkg.FindAll(specsDir)
		if err != nil {
			return 0, err
		}
		specPaths = paths
	} else {
		// Validate specific spec
		path, err := specpkg.ResolvePath(specsDir, spec)
		if err != nil {
			return 0, err
		}
		specPaths = append(specPaths, path)
	}

	if len(specPaths) == 0 {
		cmd.Println("No specs found to validate")
		return 0, nil
	}

	// Validate each spec
	var validCount int

	for _, path := range specPaths {
		result, err := validate.ValidateSpecFile(path)
		if err != nil {
			cmd.PrintErrf("Error reading %s: %v\n", filepath.Base(path), err)
			invalidCount++
			continue
		}

		cmd.Print(validate.FormatValidationResult(result))

		if result.IsValid() {
			validCount++
		} else {
			invalidCount++
		}
	}

	// Print summary
	total := validCount + invalidCount
	cmd.Printf("\n%d of %d specs valid\n", validCount, total)

	return invalidCount, nil
}
