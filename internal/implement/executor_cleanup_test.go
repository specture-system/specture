package implement

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestExecutePlan_RunsSingleFinalCleanupPassAndCreatesRefactorCommit(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] First task
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusInProgress}
	plan := Plan{
		Sections: []RemainingSection{{
			Name:  "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{{Text: "First task", Section: "Spec Updates and Section Delivery"}},
		}},
	}

	var commits []string
	var events []string
	var cleanupReviewCalls int
	var cleanupWorkerCalls int
	currentBranch := "main"
	dirtyChecks := 0

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) {
			dirtyChecks++
			if dirtyChecks == 1 {
				return false, nil
			}

			return true, nil
		},
		getCurrentBranch: func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			currentBranch = branchName
			events = append(events, "create-branch")
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll: func(dir string) error {
			events = append(events, "stage")
			return nil
		},
		commit: func(dir, message string) error {
			commits = append(commits, message)
			events = append(events, "commit")
			return nil
		},
		pushBranch: func(dir, branchName string) error {
			events = append(events, "push")
			return nil
		},
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if strings.Contains(invocation.Prompt, "final cleanup") {
				if invocation.Role == AgentRoleReviewer {
					cleanupReviewCalls++
					events = append(events, "cleanup-review")
					if !strings.Contains(invocation.Prompt, "unnecessary abstraction") {
						t.Fatalf("cleanup review prompt missing unnecessary abstraction guidance: %s", invocation.Prompt)
					}
					if !strings.Contains(invocation.Prompt, "clear AGENTS.md guideline violations") {
						t.Fatalf("cleanup review prompt missing AGENTS.md guidance: %s", invocation.Prompt)
					}
					if !strings.Contains(invocation.Prompt, "low-risk maintainability improvements") {
						t.Fatalf("cleanup review prompt missing maintainability guidance: %s", invocation.Prompt)
					}
					if !strings.Contains(invocation.Prompt, "Parent branch: main") {
						t.Fatalf("cleanup review prompt missing parent branch: %s", invocation.Prompt)
					}
					return AgentResult{Output: "- simplify helper layering"}, nil
				}

				cleanupWorkerCalls++
				events = append(events, "cleanup-worker")
				if !strings.Contains(invocation.Prompt, "simplify helper layering") {
					t.Fatalf("cleanup worker prompt missing cleanup recommendations: %s", invocation.Prompt)
				}
				return AgentResult{}, nil
			}

			return AgentResult{CriticalIssues: false}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cleanupReviewCalls != 1 {
		t.Fatalf("expected exactly one cleanup review call, got %d", cleanupReviewCalls)
	}
	if cleanupWorkerCalls != 1 {
		t.Fatalf("expected exactly one cleanup worker call, got %d", cleanupWorkerCalls)
	}

	if len(commits) != 2 {
		t.Fatalf("expected task + cleanup commits, got %d (%v)", len(commits), commits)
	}
	if commits[1] != "refactor: final cleanup for spec 007" {
		t.Fatalf("unexpected cleanup commit message: %q", commits[1])
	}

	pushIdx := -1
	cleanupReviewIdx := -1
	cleanupWorkerIdx := -1
	for i, event := range events {
		switch event {
		case "push":
			pushIdx = i
		case "cleanup-review":
			cleanupReviewIdx = i
		case "cleanup-worker":
			cleanupWorkerIdx = i
		}
	}

	if pushIdx == -1 || cleanupReviewIdx == -1 || cleanupWorkerIdx == -1 {
		t.Fatalf("expected push and cleanup events, got %v", events)
	}
	if !(pushIdx < cleanupReviewIdx && cleanupReviewIdx < cleanupWorkerIdx) {
		t.Fatalf("expected cleanup to run after push in review->worker order, got %v", events)
	}
}

func TestExecutePlan_SkipsFinalCleanupCommitWhenWorkerMakesNoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] First task
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusInProgress}
	plan := Plan{
		Sections: []RemainingSection{{
			Name:  "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{{Text: "First task", Section: "Spec Updates and Section Delivery"}},
		}},
	}

	var commits []string
	stageCalls := 0
	currentBranch := "main"
	dirtyChecks := 0

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) {
			dirtyChecks++
			if dirtyChecks == 1 {
				return false, nil
			}

			// The cleanup worker made no edits.
			return false, nil
		},
		getCurrentBranch: func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			currentBranch = branchName
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll: func(dir string) error {
			stageCalls++
			return nil
		},
		commit: func(dir, message string) error {
			commits = append(commits, message)
			return nil
		},
		pushBranch: func(dir, branchName string) error { return nil },
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if strings.Contains(invocation.Prompt, "final cleanup") && invocation.Role == AgentRoleReviewer {
				return AgentResult{Output: "cleanup notes"}, nil
			}
			return AgentResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected only task commit when cleanup has no changes, got %d (%v)", len(commits), commits)
	}
	if commits[0] != "feat: complete spec 007 task: First task" {
		t.Fatalf("unexpected task commit message: %q", commits[0])
	}
	if stageCalls != 1 {
		t.Fatalf("expected one stage call for task changes only, got %d", stageCalls)
	}
}
