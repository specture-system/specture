package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	implementpkg "github.com/specture-system/specture/internal/implement"
	specpkg "github.com/specture-system/specture/internal/spec"
)

// Note: These tests intentionally do not use t.Parallel() because implementCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr and flags).

func setupImplementTest(t *testing.T, specs map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	specsDir := filepath.Join(tmpDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("failed to create specs dir: %v", err)
	}

	for name, content := range specs {
		if err := os.WriteFile(filepath.Join(specsDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write spec %s: %v", name, err)
		}
	}

	return tmpDir
}

func execImplement(t *testing.T, tmpDir string, flags map[string]string) (string, error) {
	t.Helper()

	originalWd, _ := os.Getwd()
	originalLookPath := implementLookPath
	originalExecutePlan := implementExecutePlan

	t.Cleanup(func() {
		os.Chdir(originalWd)
		implementLookPath = originalLookPath
		implementExecutePlan = originalExecutePlan
		implementCmd.Flags().Set("spec", "")
		implementCmd.Flags().Set("agent", "")
	})

	implementExecutePlan = func(workDir string, info *specpkg.SpecInfo, plan implementpkg.Plan, backend string, printf implementpkg.PrintfFunc) error {
		return nil
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	out := &bytes.Buffer{}
	cmd := implementCmd
	cmd.SetOut(out)
	cmd.SetErr(out)

	for k, v := range flags {
		if err := cmd.Flags().Set(k, v); err != nil {
			t.Fatalf("failed to set flag %s: %v", k, err)
		}
	}

	err := runImplement(cmd, []string{})
	return out.String(), err
}

const implementApprovedSpec = `---
number: 7
status: approved
---

# Agent-Driven Implement Command

Description.

## Task List

### CLI and Planning

- [x] Existing done task
- [ ] Add failing tests for implement command

### Branch and Task Execution

- [ ] Add section branch tests
`

const implementInProgressSpec = `---
number: 8
status: in-progress
---

# In-Progress Spec

Description.

## Task List

- [x] Already done
- [ ] Remaining task
`

const implementDraftSpec = `---
number: 9
status: draft
---

# Draft Spec

Description.

## Task List

- [ ] Incomplete
`

func TestImplementCommand_RequiresSpecFlag(t *testing.T) {
	tmpDir := setupImplementTest(t, map[string]string{
		"007-agent.md": implementApprovedSpec,
	})

	implementLookPath = func(file string) (string, error) {
		return "/usr/bin/opencode", nil
	}

	_, err := execImplement(t, tmpDir, nil)
	if err == nil {
		t.Fatal("expected error when spec flag is empty")
	}

	if !strings.Contains(err.Error(), "spec flag is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImplementCommand_RejectsInvalidAgentOverride(t *testing.T) {
	tmpDir := setupImplementTest(t, map[string]string{
		"007-agent.md": implementApprovedSpec,
	})

	implementLookPath = func(file string) (string, error) {
		return "", errors.New("missing")
	}

	_, err := execImplement(t, tmpDir, map[string]string{"spec": "7", "agent": "claude"})
	if err == nil {
		t.Fatal("expected error for invalid agent override")
	}

	if !strings.Contains(err.Error(), "invalid agent backend") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImplementCommand_RejectsDisallowedSpecStatus(t *testing.T) {
	tmpDir := setupImplementTest(t, map[string]string{
		"009-draft.md": implementDraftSpec,
	})

	implementLookPath = func(file string) (string, error) {
		return "/usr/bin/opencode", nil
	}

	_, err := execImplement(t, tmpDir, map[string]string{"spec": "9"})
	if err == nil {
		t.Fatal("expected error for disallowed status")
	}

	if !strings.Contains(err.Error(), "must be 'approved' or 'in-progress'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImplementCommand_AllowsApprovedAndInProgressStatuses(t *testing.T) {
	tests := []struct {
		name     string
		specFile string
		specBody string
		specFlag string
	}{
		{name: "approved", specFile: "007-agent.md", specBody: implementApprovedSpec, specFlag: "7"},
		{name: "in-progress", specFile: "008-progress.md", specBody: implementInProgressSpec, specFlag: "8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupImplementTest(t, map[string]string{
				tt.specFile: tt.specBody,
			})

			implementLookPath = func(file string) (string, error) {
				if file == "opencode" {
					return "/usr/bin/opencode", nil
				}
				return "", errors.New("missing")
			}

			_, err := execImplement(t, tmpDir, map[string]string{"spec": tt.specFlag})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestImplementCommand_PlansRemainingSectionsAndTasks(t *testing.T) {
	tmpDir := setupImplementTest(t, map[string]string{
		"007-agent.md": implementApprovedSpec,
	})

	implementLookPath = func(file string) (string, error) {
		if file == "opencode" {
			return "/usr/bin/opencode", nil
		}
		return "", errors.New("missing")
	}

	output, err := execImplement(t, tmpDir, map[string]string{"spec": "7"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Spec 007: Agent-Driven Implement Command",
		"Status: approved",
		"Agent Backend: opencode",
		"Remaining Tasks: 2",
		"Remaining Sections:",
		"CLI and Planning (1 tasks)",
		"Add failing tests for implement command",
		"Branch and Task Execution (1 tasks)",
		"Add section branch tests",
		"Planning complete.",
	}

	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestImplementCommand_AgentOverrideBeatsAutoDetect(t *testing.T) {
	tmpDir := setupImplementTest(t, map[string]string{
		"007-agent.md": implementApprovedSpec,
	})

	implementLookPath = func(file string) (string, error) {
		switch file {
		case "opencode":
			return "/usr/bin/opencode", nil
		case "codex":
			return "/usr/bin/codex", nil
		default:
			return "", errors.New("missing")
		}
	}

	output, err := execImplement(t, tmpDir, map[string]string{"spec": "7", "agent": "codex"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Agent Backend: codex") {
		t.Fatalf("expected codex backend, got:\n%s", output)
	}
}

func TestImplementCommand_HelpMentionsOrchestratorAndExample(t *testing.T) {
	if !strings.Contains(implementCmd.Long, "agent orchestrator") {
		t.Fatalf("expected implement help to mention agent orchestrator, got:\n%s", implementCmd.Long)
	}

	if !strings.Contains(implementCmd.Long, "specture implement --spec 7") {
		t.Fatalf("expected implement help to include example usage, got:\n%s", implementCmd.Long)
	}
}
