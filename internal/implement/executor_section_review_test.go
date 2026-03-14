package implement

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestExecutePlan_UsesOverallSectionOrderFromSpecFile(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### CLI and Planning

- [x] done

### Branch and Task Execution

- [ ] remaining
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath}
	plan := Plan{Sections: []RemainingSection{{Name: "Branch and Task Execution"}}}

	created := ""
	err := executePlanWithDeps("/tmp/repo", info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch:          func(dir, branchName string) error { created = branchName; return nil },
		branchExists:          func(dir, branchName string) (bool, error) { return false, nil },
		invokeAgent:           func(invocation AgentInvocation) (AgentResult, error) { return AgentResult{}, nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created != "implement/007-02-branch-and-task-execution" {
		t.Fatalf("expected branch to use overall section number 2, got %q", created)
	}
}

func TestExecutePlan_RejectsDirtyWorktree(t *testing.T) {
	info := &specpkg.SpecInfo{Number: 7, Path: "specs/agent.md"}
	plan := Plan{Sections: []RemainingSection{{Name: "Branch and Task Execution"}}}

	err := executePlanWithDeps("/tmp/repo", info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return true, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch:          func(dir, branchName string) error { return nil },
		branchExists:          func(dir, branchName string) (bool, error) { return false, nil },
		invokeAgent:           func(invocation AgentInvocation) (AgentResult, error) { return AgentResult{}, nil },
	})
	if err == nil {
		t.Fatal("expected dirty worktree error")
	}

	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecutePlan_RerunFailClosedWhenBranchExistsButNotCheckedOut(t *testing.T) {
	info := &specpkg.SpecInfo{Number: 7, Path: "specs/agent.md"}
	plan := Plan{Sections: []RemainingSection{{Name: "Branch and Task Execution"}}}

	err := executePlanWithDeps("/tmp/repo", info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch: func(dir, branchName string) error {
			t.Fatal("createBranch should not be called when branch already exists")
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return true, nil },
		invokeAgent:  func(invocation AgentInvocation) (AgentResult, error) { return AgentResult{}, nil },
	})
	if err == nil {
		t.Fatal("expected fail-closed rerun error")
	}

	if !strings.Contains(err.Error(), "fail-closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecutePlan_CreatesSectionBranchWhenMissing(t *testing.T) {
	info := &specpkg.SpecInfo{Number: 7, Path: "specs/agent.md"}
	plan := Plan{Sections: []RemainingSection{{Name: "Branch and Task Execution"}}}

	created := ""
	err := executePlanWithDeps("/tmp/repo", info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch:          func(dir, branchName string) error { created = branchName; return nil },
		branchExists:          func(dir, branchName string) (bool, error) { return false, nil },
		invokeAgent:           func(invocation AgentInvocation) (AgentResult, error) { return AgentResult{}, nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created != "implement/007-01-branch-and-task-execution" {
		t.Fatalf("unexpected created branch: %s", created)
	}
}

func TestExecutePlan_SectionReviewGetsSingleRetryAndCommitsFixes(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] Add failing tests for section review
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusInProgress}
	plan := Plan{
		Sections: []RemainingSection{{
			Name: "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{
				{Text: "Add failing tests for section review", Section: "Spec Updates and Section Delivery"},
			},
		}},
	}

	var commits []string
	var sectionReviewCalls int
	var sectionWorkerCalls int
	dirtyChecks := 0
	currentBranch := "main"

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
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:     func(dir string) error { return nil },
		commit: func(dir, message string) error {
			commits = append(commits, message)
			return nil
		},
		pushBranch: func(dir, branchName string) error { return nil },
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleWorker && invocation.TaskText == "" {
				if strings.Contains(invocation.Prompt, "final cleanup") {
					return AgentResult{}, nil
				}
				sectionWorkerCalls++
				if !strings.Contains(invocation.Prompt, "REVIEW_CRITICAL: integration broke") {
					t.Fatalf("section worker prompt missing review summary: %s", invocation.Prompt)
				}
				return AgentResult{}, nil
			}

			if invocation.Role == AgentRoleReviewer && invocation.TaskText == "" {
				if strings.Contains(invocation.Prompt, "final cleanup") {
					return AgentResult{Output: "cleanup notes"}, nil
				}
				sectionReviewCalls++
				if !strings.Contains(invocation.Prompt, "Branch context:") {
					t.Fatalf("section review prompt missing branch context header: %s", invocation.Prompt)
				}
				if !strings.Contains(invocation.Prompt, "Current branch: implement/007-01-spec-updates-and-section-delivery") {
					t.Fatalf("section review prompt missing current branch: %s", invocation.Prompt)
				}
				if !strings.Contains(invocation.Prompt, "Parent branch: main") {
					t.Fatalf("section review prompt missing parent branch: %s", invocation.Prompt)
				}
				if sectionReviewCalls == 1 {
					return AgentResult{CriticalIssues: true, Output: "REVIEW_CRITICAL: integration broke"}, nil
				}
				return AgentResult{CriticalIssues: false, Output: "REVIEW_OK"}, nil
			}

			return AgentResult{CriticalIssues: false}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sectionReviewCalls != 2 {
		t.Fatalf("expected 2 section review calls, got %d", sectionReviewCalls)
	}
	if sectionWorkerCalls != 1 {
		t.Fatalf("expected 1 section worker retry, got %d", sectionWorkerCalls)
	}

	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d (%v)", len(commits), commits)
	}
	if commits[1] != "fix: address spec 007 section 01 review: Spec Updates and Section Delivery" {
		t.Fatalf("unexpected section review commit message: %q", commits[1])
	}
	if commits[2] != "refactor: final cleanup for spec 007" {
		t.Fatalf("unexpected cleanup commit message: %q", commits[2])
	}
}

func TestExecutePlan_StopsImmediatelyWhenSectionPushFails(t *testing.T) {
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

### Completion

- [ ] Second task
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusInProgress}
	plan := Plan{
		Sections: []RemainingSection{
			{
				Name:  "Spec Updates and Section Delivery",
				Tasks: []specpkg.Task{{Text: "First task", Section: "Spec Updates and Section Delivery"}},
			},
			{
				Name:  "Completion",
				Tasks: []specpkg.Task{{Text: "Second task", Section: "Completion"}},
			},
		},
	}

	createdBranches := 0
	pushes := 0
	currentBranch := "main"

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			createdBranches++
			currentBranch = branchName
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:     func(dir string) error { return nil },
		commit:       func(dir, message string) error { return nil },
		pushBranch: func(dir, branchName string) error {
			pushes++
			return os.ErrPermission
		},
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			return AgentResult{CriticalIssues: false}, nil
		},
	})
	if err == nil {
		t.Fatal("expected push failure")
	}

	if !strings.Contains(err.Error(), "failed to push completed section branch") {
		t.Fatalf("unexpected error: %v", err)
	}
	if pushes != 1 {
		t.Fatalf("expected exactly one push attempt, got %d", pushes)
	}
	if createdBranches != 1 {
		t.Fatalf("expected execution to stop before creating second section branch, got %d branches", createdBranches)
	}
}
