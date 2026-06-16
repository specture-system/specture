package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runNewInDir(t *testing.T, dir string, flags map[string]string) (string, error) {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	out := &bytes.Buffer{}
	cmd := newCmd
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.Flags().Set("title", "")
	cmd.Flags().Set("parent", "")
	cmd.Flags().Set("spec", "")
	cmd.Flags().Set("plan", "false")
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set %s: %v", name, err)
		}
	}

	err = cmd.RunE(cmd, nil)
	return out.String(), err
}

func TestNewCommandCreatesSpec(t *testing.T) {
	dir := t.TempDir()
	output, err := runNewInDir(t, dir, map[string]string{"title": "My Feature"})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	path := filepath.Join(dir, "specs", "000-my-feature", "SPEC.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected spec file: %v", err)
	}
	if !strings.Contains(string(content), "# My Feature") {
		t.Fatalf("created spec missing title:\n%s", content)
	}
	if !strings.Contains(output, "Creating spec 0: My Feature") || !strings.Contains(output, "File: 000-my-feature/SPEC.md") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestNewCommandCreatesPlan(t *testing.T) {
	dir := t.TempDir()
	output, err := runNewInDir(t, dir, map[string]string{"title": "Implementation Plan", "plan": "true"})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	path := filepath.Join(dir, "specs", "000-implementation-plan", "PLAN.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected plan file: %v", err)
	}
	if !strings.Contains(string(content), "## Tasks") {
		t.Fatalf("expected plan template tasks, got:\n%s", content)
	}
	if !strings.Contains(output, "Creating plan 0: Implementation Plan") {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestNewCommandCreatesPlanBesideExistingSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "specs", "123-existing", "SPEC.md")
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("---\nstatus: draft\n---\n\n# Existing\n"), 0o644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	_, err := runNewInDir(t, dir, map[string]string{"title": "Execution Plan", "spec": "123", "plan": "true"})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	planPath := filepath.Join(dir, "specs", "123-existing", "PLAN.md")
	if _, err := os.Stat(planPath); err != nil {
		t.Fatalf("expected sibling plan file: %v", err)
	}
}

func TestNewCommandExplicitChildRef(t *testing.T) {
	dir := t.TempDir()
	parentPath := filepath.Join(dir, "specs", "123-parent", "SPEC.md")
	if err := os.MkdirAll(filepath.Dir(parentPath), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(parentPath, []byte("---\nstatus: draft\n---\n\n# Parent\n"), 0o644); err != nil {
		t.Fatalf("failed to write parent: %v", err)
	}

	_, err := runNewInDir(t, dir, map[string]string{"title": "Child Feature", "spec": "123.4"})
	if err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	path := filepath.Join(dir, "specs", "123-parent", "004-child-feature", "SPEC.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected explicit child spec: %v", err)
	}
}

func TestNewCommandRequiresTitle(t *testing.T) {
	_, err := runNewInDir(t, t.TempDir(), nil)
	if err == nil || !strings.Contains(err.Error(), "--title is required") {
		t.Fatalf("expected title error, got %v", err)
	}
}
