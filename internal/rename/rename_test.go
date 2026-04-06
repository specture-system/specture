package rename

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func specWithNumber(number int) string {
	return "---\nnumber: " + strings.Replace("N", "N", itoa(number), 1) + "\n---\n\n# Test\n\n## Task List\n"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func setupSpecsDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create parent dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func TestPlan_BasicRename(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, "3", "status-command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result.OldPath) != "SPEC.md" {
		t.Errorf("expected old path to be SPEC.md, got %s", filepath.Base(result.OldPath))
	}
	if filepath.Base(filepath.Dir(result.NewPath)) != "003-status-command" {
		t.Errorf("expected new spec directory to be 003-status-command, got %s", filepath.Base(filepath.Dir(result.NewPath)))
	}
}

func TestPlan_EmptySlugError(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	_, err := Plan(dir, "3", "")
	if err == nil {
		t.Fatal("expected error for empty slug")
	}
}

func TestPlan_CustomSlug(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, "3", "spec-status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(filepath.Dir(result.NewPath)) != "003-spec-status" {
		t.Errorf("expected new path directory 003-spec-status, got %s", filepath.Base(filepath.Dir(result.NewPath)))
	}
}

func TestPlan_FindsLinkReferences(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md":     "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"005-list-command/SPEC.md": "---\nnumber: 5\n---\n\n# List Command\n\nSee [status](/specs/003-old-name/SPEC.md).\n\n## Task List\n",
	})

	result, err := Plan(dir, "3", "status-command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LinkUpdates) == 0 {
		t.Fatal("expected link updates, got none")
	}

	found := false
	for _, u := range result.LinkUpdates {
		if strings.Contains(u.OldLink, "003-old-name/SPEC.md") {
			found = true
			if !strings.Contains(u.NewLink, "003-status-command/SPEC.md") {
				t.Errorf("expected new link to contain updated spec path, got %s", u.NewLink)
			}
		}
	}
	if !found {
		t.Error("expected to find link update for 003-old-name/SPEC.md")
	}
}

func TestPlan_SameNameError(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	_, err := Plan(dir, "3", "status-command")
	if err == nil {
		t.Fatal("expected error for same name")
	}
}

func TestPlan_TargetExistsError(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md":       "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"003-status-command/SPEC.md": "---\nnumber: 99\n---\n\n# Other\n\n## Task List\n",
	})

	_, err := Plan(dir, "3", "status-command")
	if err == nil {
		t.Fatal("expected error when target exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

func TestExecute_RenamesFile(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, "3", "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	if err := Execute(result); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Old directory should be gone
	if _, err := os.Stat(filepath.Join(dir, "003-old-name")); err == nil {
		t.Error("old directory should not exist after rename")
	}

	// New directory should exist
	if _, err := os.Stat(filepath.Join(dir, "003-status-command", "SPEC.md")); err != nil {
		t.Error("new spec should exist after rename")
	}
}

func TestExecute_UpdatesLinks(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md":     "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"005-list-command/SPEC.md": "---\nnumber: 5\n---\n\n# List\n\nSee [status](/specs/003-old-name/SPEC.md).\n\n## Task List\n",
	})

	result, err := Plan(dir, "3", "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	if err := Execute(result); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Check that the link was updated
	content, err := os.ReadFile(filepath.Join(dir, "005-list-command", "SPEC.md"))
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	if strings.Contains(string(content), "003-old-name/SPEC.md") {
		t.Error("old link should be updated")
	}
	if !strings.Contains(string(content), "/specs/003-status-command/SPEC.md") {
		t.Errorf("expected new link, got: %s", string(content))
	}
}

func TestPlan_DryRunDoesNotModify(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-old-name/SPEC.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	// Plan only, don't execute
	_, err := Plan(dir, "3", "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	// File should still exist with old name
	if _, err := os.Stat(filepath.Join(dir, "003-old-name", "SPEC.md")); err != nil {
		t.Error("file should still exist after Plan (no Execute)")
	}
}

func TestPlan_DottedRef(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"000-root/SPEC.md":                    "---\nnumber: 0\n---\n\n# Root\n",
		"000-root/001-child/SPEC.md":          "---\nnumber: 1\n---\n\n# Child\n",
		"000-root/001-child/002-leaf/SPEC.md": "---\nnumber: 2\n---\n\n# Leaf\n",
	})

	result, err := Plan(dir, "0.1.2", "leaf-renamed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(filepath.Dir(result.OldPath)) != "002-leaf" {
		t.Fatalf("expected old path to resolve to 002-leaf, got %s", filepath.Base(filepath.Dir(result.OldPath)))
	}
	if filepath.Base(filepath.Dir(result.NewPath)) != "002-leaf-renamed" {
		t.Fatalf("expected new path to be 002-leaf-renamed, got %s", filepath.Base(filepath.Dir(result.NewPath)))
	}
}

func TestStripNumericPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"003-status-command/SPEC.md", "status-command/SPEC.md"},
		{"000-basic-cli/SPEC.md", "basic-cli/SPEC.md"},
		{"status-command/SPEC.md", "status-command/SPEC.md"}, // no prefix, unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripNumericPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("stripNumericPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
