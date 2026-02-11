package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSkill_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()

	if err := InstallSkill(tmpDir, false); err != nil {
		t.Fatalf("InstallSkill failed: %v", err)
	}

	// Check SKILL.md was created with expected content
	skillPath := filepath.Join(tmpDir, ".agents", "skills", "specture", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("SKILL.md not created: %v", err)
	}
	if !strings.Contains(string(content), "specture") {
		t.Error("SKILL.md should contain 'specture'")
	}

	// Check references/spec-format.md was created with expected content
	refPath := filepath.Join(tmpDir, ".agents", "skills", "specture", "references", "spec-format.md")
	content, err = os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("references/spec-format.md not created: %v", err)
	}
	if !strings.Contains(string(content), "Spec File Format") {
		t.Error("spec-format.md should contain 'Spec File Format'")
	}
}

func TestInstallSkill_OverwritesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing files with stale content
	skillDir := filepath.Join(tmpDir, ".agents", "skills", "specture")
	refsDir := filepath.Join(skillDir, "references")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	refPath := filepath.Join(refsDir, "spec-format.md")

	oldContent := []byte("old content")
	if err := os.WriteFile(skillPath, oldContent, 0644); err != nil {
		t.Fatalf("failed to write existing SKILL.md: %v", err)
	}
	if err := os.WriteFile(refPath, oldContent, 0644); err != nil {
		t.Fatalf("failed to write existing spec-format.md: %v", err)
	}

	// Run install — should overwrite
	if err := InstallSkill(tmpDir, false); err != nil {
		t.Fatalf("InstallSkill failed: %v", err)
	}

	// Verify files were overwritten
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	if string(content) == "old content" {
		t.Error("SKILL.md should have been overwritten")
	}

	content, err = os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("failed to read spec-format.md: %v", err)
	}
	if string(content) == "old content" {
		t.Error("spec-format.md should have been overwritten")
	}
}

func TestInstallSkill_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	if err := InstallSkill(tmpDir, true); err != nil {
		t.Fatalf("InstallSkill dry-run failed: %v", err)
	}

	// Verify no files were created
	agentsDir := filepath.Join(tmpDir, ".agents")
	if _, err := os.Stat(agentsDir); err == nil {
		t.Error(".agents directory should not be created in dry-run mode")
	}
}
