package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateSpecsLayout(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "002-status-command.md"), []byte("---\nnumber: 2\nstatus: completed\n---\n\n# Status Command\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "004-list-command.md"), []byte("---\nnumber: 4\nstatus: draft\n---\n\n# List Command\n\nSee [status](/specs/002-status-command.md).\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	migrated, err := MigrateSpecsLayout(dir, false)
	if err != nil {
		t.Fatalf("MigrateSpecsLayout failed: %v", err)
	}
	if !migrated {
		t.Fatal("expected migration to occur")
	}

	statusPath := filepath.Join(dir, "002-status-command", "SPEC.md")
	if _, err := os.Stat(statusPath); err != nil {
		t.Fatalf("expected migrated spec at %s: %v", statusPath, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "002-status-command.md")); !os.IsNotExist(err) {
		t.Fatalf("old status spec should be removed, got: %v", err)
	}

	listPath := filepath.Join(dir, "004-list-command", "SPEC.md")
	content, err := os.ReadFile(listPath)
	if err != nil {
		t.Fatalf("failed to read migrated list spec: %v", err)
	}
	if strings.Contains(string(content), "/specs/002-status-command.md") {
		t.Fatalf("old spec link should have been rewritten, got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "specs/002-status-command/SPEC.md") {
		t.Fatalf("expected rewritten link, got:\n%s", string(content))
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignore, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read specs .gitignore: %v", err)
	}
	if string(gitignore) != specsGitignoreContent {
		t.Fatalf("unexpected .gitignore content:\n%s", string(gitignore))
	}
}

func TestMigrateSkillsDir_OldExistsNewDoesNot(t *testing.T) {
	tmpDir := t.TempDir()

	// Create old .skills/specture/ with files
	oldDir := filepath.Join(tmpDir, ".skills", "specture")
	oldRefsDir := filepath.Join(oldDir, "references")
	if err := os.MkdirAll(oldRefsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "SKILL.md"), []byte("skill content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldRefsDir, "spec-format.md"), []byte("ref content"), 0644); err != nil {
		t.Fatal(err)
	}

	migrated, err := MigrateSkillsDir(tmpDir, false)
	if err != nil {
		t.Fatalf("MigrateSkillsDir failed: %v", err)
	}
	if !migrated {
		t.Error("expected migration to occur")
	}

	// New location should have files
	newSkill := filepath.Join(tmpDir, ".agents", "skills", "specture", "SKILL.md")
	content, err := os.ReadFile(newSkill)
	if err != nil {
		t.Fatalf("new SKILL.md not found: %v", err)
	}
	if string(content) != "skill content" {
		t.Errorf("expected 'skill content', got %q", string(content))
	}

	newRef := filepath.Join(tmpDir, ".agents", "skills", "specture", "references", "spec-format.md")
	content, err = os.ReadFile(newRef)
	if err != nil {
		t.Fatalf("new spec-format.md not found: %v", err)
	}
	if string(content) != "ref content" {
		t.Errorf("expected 'ref content', got %q", string(content))
	}

	// Old .skills/ should be removed (was empty after move)
	if _, err := os.Stat(filepath.Join(tmpDir, ".skills")); !os.IsNotExist(err) {
		t.Error(".skills/ should have been removed")
	}
}

func TestMigrateSkillsDir_BothExist(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir := filepath.Join(tmpDir, ".skills", "specture")
	newDir := filepath.Join(tmpDir, ".agents", "skills", "specture")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "SKILL.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "SKILL.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	migrated, err := MigrateSkillsDir(tmpDir, false)
	if err != nil {
		t.Fatalf("MigrateSkillsDir failed: %v", err)
	}
	if migrated {
		t.Error("expected no migration when both exist")
	}

	content, _ := os.ReadFile(filepath.Join(newDir, "SKILL.md"))
	if string(content) != "new" {
		t.Error("new SKILL.md should not be modified")
	}
}

func TestMigrateSkillsDir_OldDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	migrated, err := MigrateSkillsDir(tmpDir, false)
	if err != nil {
		t.Fatalf("MigrateSkillsDir failed: %v", err)
	}
	if migrated {
		t.Error("expected no migration when old doesn't exist")
	}
}

func TestMigrateSkillsDir_NonEmptySkillsDirAfterMove(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir := filepath.Join(tmpDir, ".skills", "specture")
	otherDir := filepath.Join(tmpDir, ".skills", "other-skill")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "SKILL.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(otherDir, "SKILL.md"), []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}

	migrated, err := MigrateSkillsDir(tmpDir, false)
	if err != nil {
		t.Fatalf("MigrateSkillsDir failed: %v", err)
	}
	if !migrated {
		t.Error("expected migration to occur")
	}

	// .skills/ should still exist (has other-skill/)
	if _, err := os.Stat(filepath.Join(tmpDir, ".skills", "other-skill", "SKILL.md")); err != nil {
		t.Error(".skills/other-skill/ should still exist")
	}
}

func TestMigrateSkillsDir_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir := filepath.Join(tmpDir, ".skills", "specture")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "SKILL.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	migrated, err := MigrateSkillsDir(tmpDir, true)
	if err != nil {
		t.Fatalf("MigrateSkillsDir dry-run failed: %v", err)
	}
	if !migrated {
		t.Error("expected dry-run to report migration would occur")
	}

	// Old should still exist
	if _, err := os.Stat(filepath.Join(oldDir, "SKILL.md")); err != nil {
		t.Error("old SKILL.md should still exist in dry-run")
	}

	// New should not exist
	newDir := filepath.Join(tmpDir, ".agents", "skills", "specture")
	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Error(".agents/skills/specture/ should not exist in dry-run")
	}
}
