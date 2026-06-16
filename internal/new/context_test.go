package new

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewContext(t *testing.T) {
	t.Run("creates top level spec without git preconditions", func(t *testing.T) {
		workDir := t.TempDir()

		ctx, err := NewContext(workDir, Options{Title: "My First Spec"})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.Number != 0 {
			t.Fatalf("Number = %d, want 0", ctx.Number)
		}
		if ctx.FullRef != "0" {
			t.Fatalf("FullRef = %q, want 0", ctx.FullRef)
		}
		if ctx.FileName != "SPEC.md" {
			t.Fatalf("FileName = %q, want SPEC.md", ctx.FileName)
		}
		wantPath := filepath.Join(workDir, "specs", "000-my-first-spec", "SPEC.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("creates plan file", func(t *testing.T) {
		workDir := t.TempDir()

		ctx, err := NewContext(workDir, Options{Title: "Execution Plan", Plan: true})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.Kind != "plan" {
			t.Fatalf("Kind = %q, want plan", ctx.Kind)
		}
		if ctx.FileName != "PLAN.md" {
			t.Fatalf("FileName = %q, want PLAN.md", ctx.FileName)
		}
		wantPath := filepath.Join(workDir, "specs", "000-execution-plan", "PLAN.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("allocates next child under parent", func(t *testing.T) {
		workDir := t.TempDir()
		parentDir := filepath.Join(workDir, "specs", "011-parent")
		writeFile(t, filepath.Join(parentDir, "SPEC.md"), "---\nstatus: draft\n---\n\n# Parent\n")
		writeFile(t, filepath.Join(parentDir, "000-existing", "SPEC.md"), "---\nstatus: draft\n---\n\n# Existing\n")

		ctx, err := NewContext(workDir, Options{Title: "Child Spec", ParentRef: "11"})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.FullRef != "11.1" {
			t.Fatalf("FullRef = %q, want 11.1", ctx.FullRef)
		}
		wantPath := filepath.Join(parentDir, "001-child-spec", "SPEC.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("uses explicit top level ref", func(t *testing.T) {
		workDir := t.TempDir()

		ctx, err := NewContext(workDir, Options{Title: "Issue Spec", SpecRef: "123"})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.FullRef != "123" {
			t.Fatalf("FullRef = %q, want 123", ctx.FullRef)
		}
		wantPath := filepath.Join(workDir, "specs", "123-issue-spec", "SPEC.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("uses explicit child ref", func(t *testing.T) {
		workDir := t.TempDir()
		parentDir := filepath.Join(workDir, "specs", "123-parent")
		writeFile(t, filepath.Join(parentDir, "SPEC.md"), "---\nstatus: draft\n---\n\n# Parent\n")

		ctx, err := NewContext(workDir, Options{Title: "Child Spec", SpecRef: "123.4"})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.FullRef != "123.4" {
			t.Fatalf("FullRef = %q, want 123.4", ctx.FullRef)
		}
		wantPath := filepath.Join(parentDir, "004-child-spec", "SPEC.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("explicit ref can target sibling plan in existing spec directory", func(t *testing.T) {
		workDir := t.TempDir()
		existingDir := filepath.Join(workDir, "specs", "123-existing")
		writeFile(t, filepath.Join(existingDir, "SPEC.md"), "---\nstatus: draft\n---\n\n# Existing\n")

		ctx, err := NewContext(workDir, Options{Title: "Execution Plan", SpecRef: "123", Plan: true})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		if ctx.FullRef != "123" {
			t.Fatalf("FullRef = %q, want 123", ctx.FullRef)
		}
		wantPath := filepath.Join(existingDir, "PLAN.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("explicit ref can target sibling spec in existing plan directory", func(t *testing.T) {
		workDir := t.TempDir()
		existingDir := filepath.Join(workDir, "specs", "123-existing")
		writeFile(t, filepath.Join(existingDir, "PLAN.md"), "---\nstatus: draft\n---\n\n# Existing\n")

		ctx, err := NewContext(workDir, Options{Title: "Durable Spec", SpecRef: "123"})
		if err != nil {
			t.Fatalf("NewContext() error = %v", err)
		}

		wantPath := filepath.Join(existingDir, "SPEC.md")
		if ctx.FilePath != wantPath {
			t.Fatalf("FilePath = %q, want %q", ctx.FilePath, wantPath)
		}
	})

	t.Run("rejects spec and parent together", func(t *testing.T) {
		_, err := NewContext(t.TempDir(), Options{Title: "Bad", ParentRef: "1", SpecRef: "2"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreateFile(t *testing.T) {
	workDir := t.TempDir()
	ctx, err := NewContext(workDir, Options{Title: "Created Spec"})
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	if err := ctx.CreateFile(); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	content, err := os.ReadFile(ctx.FilePath)
	if err != nil {
		t.Fatalf("created file missing: %v", err)
	}
	if string(content) == "" {
		t.Fatal("created file is empty")
	}

	if err := ctx.CreateFile(); err == nil {
		t.Fatal("expected duplicate file creation to fail")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
