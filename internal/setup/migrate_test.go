package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindSpecsNeedingMigration(t *testing.T) {
	t.Run("finds_specs_without_number", func(t *testing.T) {
		dir := t.TempDir()

		// Spec without number in frontmatter
		os.WriteFile(filepath.Join(dir, "003-status-command.md"), []byte("---\nstatus: approved\n---\n\n# Status\n"), 0644)

		// Spec with number already
		os.WriteFile(filepath.Join(dir, "005-list-command.md"), []byte("---\nnumber: 5\nstatus: draft\n---\n\n# List\n"), 0644)

		// Non-spec file
		os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Readme"), 0644)

		// Slug-only file (no prefix, no migration needed)
		os.WriteFile(filepath.Join(dir, "my-feature.md"), []byte("---\nstatus: draft\n---\n"), 0644)

		results, err := FindSpecsNeedingMigration(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 spec needing migration, got %d", len(results))
		}

		if results[0].Number != 3 {
			t.Errorf("expected number 3, got %d", results[0].Number)
		}
		if !strings.HasSuffix(results[0].Path, "003-status-command.md") {
			t.Errorf("expected path ending with 003-status-command.md, got %s", results[0].Path)
		}
	})

	t.Run("empty_when_all_migrated", func(t *testing.T) {
		dir := t.TempDir()

		os.WriteFile(filepath.Join(dir, "003-status.md"), []byte("---\nnumber: 3\nstatus: draft\n---\n"), 0644)

		results, err := FindSpecsNeedingMigration(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("empty_directory", func(t *testing.T) {
		dir := t.TempDir()

		results, err := FindSpecsNeedingMigration(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}

func TestAddNumberToFrontmatter(t *testing.T) {
	t.Run("adds_number_to_existing_frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "003-test.md")

		original := "---\nstatus: approved\nauthor: Test\n---\n\n# Test\n"
		os.WriteFile(path, []byte(original), 0644)

		if err := AddNumberToFrontmatter(path, 3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(path)
		contentStr := string(content)

		if !strings.Contains(contentStr, "number: 3") {
			t.Errorf("expected 'number: 3' in content, got:\n%s", contentStr)
		}

		// Number should be first field after opening ---
		lines := strings.Split(contentStr, "\n")
		if lines[0] != "---" {
			t.Errorf("expected first line to be ---, got %q", lines[0])
		}
		if lines[1] != "number: 3" {
			t.Errorf("expected second line to be 'number: 3', got %q", lines[1])
		}

		// Rest of frontmatter should be preserved
		if !strings.Contains(contentStr, "status: approved") {
			t.Errorf("original frontmatter should be preserved")
		}
		if !strings.Contains(contentStr, "# Test") {
			t.Errorf("body content should be preserved")
		}
	})

	t.Run("number_zero", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "000-test.md")

		original := "---\nstatus: draft\n---\n\n# Test\n"
		os.WriteFile(path, []byte(original), 0644)

		if err := AddNumberToFrontmatter(path, 0); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "number: 0") {
			t.Errorf("expected 'number: 0' in content, got:\n%s", string(content))
		}
	})
}

func TestHasNumberInFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"has number", "---\nnumber: 3\nstatus: draft\n---\n", true},
		{"no number", "---\nstatus: draft\n---\n", false},
		{"no frontmatter", "# Title\n", false},
		{"number in body not frontmatter", "---\nstatus: draft\n---\nnumber: 3\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.md")
			os.WriteFile(path, []byte(tt.content), 0644)

			got, err := hasNumberInFrontmatter(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("hasNumberInFrontmatter() = %v, want %v", got, tt.want)
			}
		})
	}
}
