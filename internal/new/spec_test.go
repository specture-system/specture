package new

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Title", "simple-title"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"under_score_title", "under-score-title"},
		{"Mixed Case_With Spaces", "mixed-case-with-spaces"},
		{"Already-kebab-case", "already-kebab-case"},
		{"Special!@#$Chars", "specialchars"},
		{"Numbers123", "numbers123"},
		{"UPPERCASE", "uppercase"},
		{"---multiple---hyphens---", "multiple-hyphens"},
		{"-leading and trailing-", "leading-and-trailing"},
		{"Title With Dots.", "title-with-dots"},
		{"", ""},
		{"123", "123"},
		{"Title\nWith\nNewlines", "titlewithnewlines"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToSlug(tt.input)
			if result != tt.expected {
				t.Errorf("ToSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindNextSpecNumber(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string]string // filename -> content
		expected    int
		expectError bool
	}{
		{
			name:       "empty_directory",
			setupFiles: map[string]string{},
			expected:   0,
		},
		{
			name: "single_spec_with_number",
			setupFiles: map[string]string{
				"first.md": "---\nnumber: 0\n---\n\n# First\n\n## Task List\n",
			},
			expected: 1,
		},
		{
			name: "multiple_specs",
			setupFiles: map[string]string{
				"000-first.md":  "---\nnumber: 0\n---\n\n# First\n\n## Task List\n",
				"001-second.md": "---\nnumber: 1\n---\n\n# Second\n\n## Task List\n",
				"002-third.md":  "---\nnumber: 2\n---\n\n# Third\n\n## Task List\n",
			},
			expected: 3,
		},
		{
			name: "non_sequential_numbers",
			setupFiles: map[string]string{
				"first.md": "---\nnumber: 0\n---\n\n# First\n\n## Task List\n",
				"fifth.md": "---\nnumber: 5\n---\n\n# Fifth\n\n## Task List\n",
				"third.md": "---\nnumber: 2\n---\n\n# Third\n\n## Task List\n",
			},
			expected: 6,
		},
		{
			name: "ignores_files_without_number",
			setupFiles: map[string]string{
				"README.md":    "# Readme\n",
				"000-spec.md":  "---\nnumber: 0\n---\n\n# Spec\n\n## Task List\n",
				"notes.txt":    "some notes",
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create setup files
			for file, content := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, file)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			result, err := FindNextSpecNumber(tmpDir)
			if (err != nil) != tt.expectError {
				t.Errorf("FindNextSpecNumber() error = %v, want error = %v", err, tt.expectError)
			}
			if err == nil && result != tt.expected {
				t.Errorf("FindNextSpecNumber() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestRenderSpec(t *testing.T) {
	t.Run("renders_with_title_and_author", func(t *testing.T) {
		result, err := RenderSpec("Test Feature", "Test Author", 5)
		if err != nil {
			t.Fatalf("RenderSpec() error = %v", err)
		}

		// Check that the template was rendered with the values
		if !strings.Contains(result, "# Test Feature") {
			t.Errorf("rendered spec doesn't contain title")
		}
		if !strings.Contains(result, "author: Test Author") {
			t.Errorf("rendered spec doesn't contain author")
		}
		if !strings.Contains(result, "status: draft") {
			t.Errorf("rendered spec doesn't contain status")
		}
		if !strings.Contains(result, "number: 5") {
			t.Errorf("rendered spec doesn't contain number")
		}
	})

	t.Run("includes_creation_date", func(t *testing.T) {
		result, err := RenderSpec("Test", "Author", 0)
		if err != nil {
			t.Fatalf("RenderSpec() error = %v", err)
		}

		if !strings.Contains(result, "creation_date:") {
			t.Errorf("rendered spec doesn't contain creation_date")
		}
	})
}

func TestGenerateFrontmatter(t *testing.T) {
	t.Run("generates_valid_frontmatter", func(t *testing.T) {
		frontmatter, err := GenerateFrontmatter("Test Spec", "Test Author", 5)
		if err != nil {
			t.Fatalf("GenerateFrontmatter() error = %v", err)
		}

		if !strings.Contains(frontmatter, "---") {
			t.Errorf("frontmatter missing YAML delimiters")
		}
		if !strings.Contains(frontmatter, "number: 5") {
			t.Errorf("frontmatter missing number")
		}
		if !strings.Contains(frontmatter, "status: draft") {
			t.Errorf("frontmatter missing status")
		}
		if !strings.Contains(frontmatter, "author: Test Author") {
			t.Errorf("frontmatter missing author")
		}
		if !strings.Contains(frontmatter, "creation_date:") {
			t.Errorf("frontmatter missing creation_date")
		}
	})

	t.Run("number_zero", func(t *testing.T) {
		frontmatter, err := GenerateFrontmatter("Test", "Author", 0)
		if err != nil {
			t.Fatalf("GenerateFrontmatter() error = %v", err)
		}
		if !strings.Contains(frontmatter, "number: 0") {
			t.Errorf("frontmatter should contain 'number: 0', got:\n%s", frontmatter)
		}
	})
}

func TestRenderDefaultBody(t *testing.T) {
	t.Run("renders_default_body_template", func(t *testing.T) {
		body, err := RenderDefaultBody("Test Spec")
		if err != nil {
			t.Fatalf("RenderDefaultBody() error = %v", err)
		}

		// Should contain the title heading
		if !strings.Contains(body, "# Test Spec") {
			t.Errorf("body doesn't contain title heading")
		}

		// Should not contain frontmatter
		if strings.Contains(body, "status: draft") {
			t.Errorf("body shouldn't contain frontmatter")
		}
	})
}

func TestJoinSpecContent(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter string
		body        string
		check       func(string) bool
	}{
		{
			name: "joins_with_blank_line",
			frontmatter: `---
status: draft
author: Test
creation_date: 2025-01-21
---`,
			body: "# Content\n\nBody text",
			check: func(result string) bool {
				return strings.Contains(result, "---\n\n# Content")
			},
		},
		{
			name:        "preserves_both_parts",
			frontmatter: "---\nstatus: draft\n---",
			body:        "Body content",
			check: func(result string) bool {
				return strings.Contains(result, "status: draft") &&
					strings.Contains(result, "Body content")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinSpecContent(tt.frontmatter, tt.body)

			if !tt.check(result) {
				t.Errorf("JoinSpecContent() result doesn't meet check:\n%s", result)
			}
		})
	}
}
