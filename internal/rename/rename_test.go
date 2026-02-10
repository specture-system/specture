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
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func TestPlan_BasicRename(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, 3, "status-command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result.OldPath) != "003-status-command.md" {
		t.Errorf("expected old path to be 003-status-command.md, got %s", filepath.Base(result.OldPath))
	}
	if filepath.Base(result.NewPath) != "status-command.md" {
		t.Errorf("expected new path to be status-command.md, got %s", filepath.Base(result.NewPath))
	}
}

func TestPlan_DefaultStripsPrefix(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	// No slug provided — should strip numeric prefix
	result, err := Plan(dir, 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result.NewPath) != "status-command.md" {
		t.Errorf("expected new path status-command.md, got %s", filepath.Base(result.NewPath))
	}
}

func TestPlan_CustomSlug(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, 3, "spec-status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(result.NewPath) != "spec-status.md" {
		t.Errorf("expected new path spec-status.md, got %s", filepath.Base(result.NewPath))
	}
}

func TestPlan_FindsLinkReferences(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"005-list-command.md": "---\nnumber: 5\n---\n\n# List Command\n\nSee [status](/specs/003-status-command.md).\n\n## Task List\n",
	})

	result, err := Plan(dir, 3, "status-command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LinkUpdates) == 0 {
		t.Fatal("expected link updates, got none")
	}

	found := false
	for _, u := range result.LinkUpdates {
		if strings.Contains(u.OldLink, "003-status-command.md") {
			found = true
			if !strings.Contains(u.NewLink, "status-command.md") {
				t.Errorf("expected new link to contain status-command.md, got %s", u.NewLink)
			}
		}
	}
	if !found {
		t.Error("expected to find link update for 003-status-command.md")
	}
}

func TestPlan_SameNameError(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	_, err := Plan(dir, 3, "status-command")
	if err == nil {
		t.Fatal("expected error for same name")
	}
}

func TestPlan_TargetExistsError(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"status-command.md":     "---\nnumber: 99\n---\n\n# Other\n\n## Task List\n",
	})

	_, err := Plan(dir, 3, "status-command")
	if err == nil {
		t.Fatal("expected error when target exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

func TestExecute_RenamesFile(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	result, err := Plan(dir, 3, "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	if err := Execute(result); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(filepath.Join(dir, "003-status-command.md")); err == nil {
		t.Error("old file should not exist after rename")
	}

	// New file should exist
	if _, err := os.Stat(filepath.Join(dir, "status-command.md")); err != nil {
		t.Error("new file should exist after rename")
	}
}

func TestExecute_UpdatesLinks(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
		"005-list-command.md":   "---\nnumber: 5\n---\n\n# List\n\nSee [status](/specs/003-status-command.md).\n\n## Task List\n",
	})

	result, err := Plan(dir, 3, "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	if err := Execute(result); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Check that the link was updated
	content, err := os.ReadFile(filepath.Join(dir, "005-list-command.md"))
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	if strings.Contains(string(content), "003-status-command.md") {
		t.Error("old link should be updated")
	}
	if !strings.Contains(string(content), "/specs/status-command.md") {
		t.Errorf("expected new link, got: %s", string(content))
	}
}

func TestPlan_DryRunDoesNotModify(t *testing.T) {
	dir := setupSpecsDir(t, map[string]string{
		"003-status-command.md": "---\nnumber: 3\n---\n\n# Status Command\n\n## Task List\n",
	})

	// Plan only, don't execute
	_, err := Plan(dir, 3, "status-command")
	if err != nil {
		t.Fatalf("Plan error: %v", err)
	}

	// File should still exist with old name
	if _, err := os.Stat(filepath.Join(dir, "003-status-command.md")); err != nil {
		t.Error("file should still exist after Plan (no Execute)")
	}
}

func TestStripNumericPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"003-status-command.md", "status-command.md"},
		{"000-basic-cli.md", "basic-cli.md"},
		{"status-command.md", "status-command.md"}, // no prefix, unchanged
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
