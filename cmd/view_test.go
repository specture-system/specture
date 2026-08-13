package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewCommand_UsesVisualWithArgumentsAndWaits(t *testing.T) {
	repoDir, childPath := setupViewTest(t)
	editor := writeViewEditor(t)

	t.Setenv("VISUAL", editor+` --visual "two words"`)
	t.Setenv("EDITOR", "should-not-run")

	output, err := execView(t, repoDir, "1.2", "standard input\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{"--visual", "two words", childPath, "standard input"}, "\n") + "\n"
	if output != want {
		t.Fatalf("expected editor output %q, got %q", want, output)
	}
}

func TestViewCommand_FallsBackToEditor(t *testing.T) {
	repoDir, childPath := setupViewTest(t)
	editor := writeViewEditor(t)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor+" --editor")

	output, err := execView(t, repoDir, "001.002", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{"--editor", childPath}, "\n") + "\n"
	if output != want {
		t.Fatalf("expected editor output %q, got %q", want, output)
	}
}

func TestViewCommand_DefaultsToCat(t *testing.T) {
	repoDir, _ := setupViewTest(t)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	output, err := execView(t, repoDir, "1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "# Parent\n" {
		t.Fatalf("expected cat output %q, got %q", "# Parent\n", output)
	}
}

func TestViewCommand_AcceptsReferencesOnly(t *testing.T) {
	repoDir, childPath := setupViewTest(t)
	t.Setenv("VISUAL", writeViewEditor(t))

	_, err := execView(t, repoDir, childPath, "")
	if err == nil || !strings.Contains(err.Error(), "invalid spec reference") {
		t.Fatalf("expected path to be rejected, got %v", err)
	}
}

func TestViewCommand_ReportsEditorFailure(t *testing.T) {
	repoDir, _ := setupViewTest(t)
	t.Setenv("VISUAL", "exit 7")

	_, err := execView(t, repoDir, "1", "")
	if err == nil || !strings.Contains(err.Error(), "editor command failed") {
		t.Fatalf("expected editor failure, got %v", err)
	}
}

func execView(t *testing.T, repoDir, ref, input string) (string, error) {
	t.Helper()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	var output bytes.Buffer
	viewCmd.SetIn(strings.NewReader(input))
	viewCmd.SetOut(&output)
	viewCmd.SetErr(&output)

	err = runView(viewCmd, []string{ref})
	return output.String(), err
}

func setupViewTest(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()
	parentDir := filepath.Join(repoDir, "specs", "001-parent")
	childDir := filepath.Join(parentDir, "002-child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directories: %v", err)
	}

	parentPath := filepath.Join(parentDir, "SPEC.md")
	childPath := filepath.Join(childDir, "SPEC.md")
	if err := os.WriteFile(parentPath, []byte("# Parent\n"), 0o644); err != nil {
		t.Fatalf("failed to write parent spec: %v", err)
	}
	if err := os.WriteFile(childPath, []byte("# Child\n"), 0o644); err != nil {
		t.Fatalf("failed to write child spec: %v", err)
	}
	return repoDir, childPath
}

func writeViewEditor(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor")
	content := "#!/bin/sh\nprintf '%s\\n' \"$@\"\ncat\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write editor: %v", err)
	}
	return path
}
