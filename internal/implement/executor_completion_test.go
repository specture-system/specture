package implement

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestExecutePlan_LeavesApprovedSpecUnchangedAndCommitsAcceptedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: approved
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] Add failing tests for in-progress transition
- [ ] Implement deterministic task commits

### Completion

- [ ] Finish remaining work
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusApproved}
	plan := Plan{
		Sections: []RemainingSection{{
			Name: "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{
				{Text: "Add failing tests for in-progress transition", Section: "Spec Updates and Section Delivery"},
				{Text: "Implement deterministic task commits", Section: "Spec Updates and Section Delivery"},
			},
		}},
	}

	var commits []string
	var pushes []string
	currentBranch := "main"

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return currentBranch, nil },
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
		pushBranch: func(dir, branchName string) error {
			pushes = append(pushes, branchName)
			return nil
		},
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			return AgentResult{}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec: %v", err)
	}

	updatedText := string(updated)
	if !strings.Contains(updatedText, "status: approved") {
		t.Fatalf("expected spec status to remain approved, got:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [ ] Add failing tests for in-progress transition") {
		t.Fatalf("expected first task to remain unchecked, got:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [ ] Implement deterministic task commits") {
		t.Fatalf("expected second task to remain unchecked, got:\n%s", updatedText)
	}

	expectedCommits := []string{
		"feat: complete spec 007 task: Add failing tests for in-progress transition",
		"feat: complete spec 007 task: Implement deterministic task commits",
	}
	if len(commits) != len(expectedCommits) {
		t.Fatalf("expected %d commits, got %d (%v)", len(expectedCommits), len(commits), commits)
	}
	for i, want := range expectedCommits {
		if commits[i] != want {
			t.Fatalf("unexpected commit %d: got %q want %q", i, commits[i], want)
		}
	}

	if len(pushes) != 2 ||
		pushes[0] != "implement/007-01-spec-updates-and-section-delivery" ||
		pushes[1] != "implement/007-01-spec-updates-and-section-delivery" {
		t.Fatalf("unexpected pushes: %v", pushes)
	}
}

func TestExecutePlan_LeavesInProgressSpecUnchangedAfterAllRemainingTasksFinish(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### Completion

- [ ] Implement the final completion update when all remaining tasks are done
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusInProgress}
	plan := Plan{
		Sections: []RemainingSection{{
			Name: "Completion",
			Tasks: []specpkg.Task{{
				Text:    "Implement the final completion update when all remaining tasks are done",
				Section: "Completion",
			}},
		}},
	}

	currentBranch := "main"

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			currentBranch = branchName
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:     func(dir string) error { return nil },
		commit:       func(dir, message string) error { return nil },
		pushBranch:   func(dir, branchName string) error { return nil },
		invokeAgent:  func(invocation AgentInvocation) (AgentResult, error) { return AgentResult{}, nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec: %v", err)
	}

	updatedText := string(updated)
	if !strings.Contains(updatedText, "status: in-progress") {
		t.Fatalf("expected spec status to remain in-progress, got:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [ ] Implement the final completion update when all remaining tasks are done") {
		t.Fatalf("expected task to remain unchecked, got:\n%s", updatedText)
	}
}

func TestExecutePlan_RetriesTaskCommitFailuresWithWorkerFixes(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: approved
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] First task
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusApproved}
	plan := Plan{
		Sections: []RemainingSection{{
			Name:  "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{{Text: "First task", Section: "Spec Updates and Section Delivery"}},
		}},
	}

	currentBranch := "main"
	commitCalls := 0
	commitFixWorkerCalls := 0

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			currentBranch = branchName
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:     func(dir string) error { return nil },
		commit: func(dir, message string) error {
			commitCalls++
			if message == "feat: complete spec 007 task: First task" && commitCalls == 1 {
				return os.ErrPermission
			}
			return nil
		},
		pushBranch: func(dir, branchName string) error { return nil },
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleWorker && invocation.TaskText == "First task" && invocation.Attempt > maxWorkerPassesPerTask {
				commitFixWorkerCalls++
				if !strings.Contains(invocation.Prompt, "Commit failure output:") {
					t.Fatalf("commit-fix worker prompt missing commit failure context: %s", invocation.Prompt)
				}
				if !strings.Contains(invocation.Prompt, os.ErrPermission.Error()) {
					t.Fatalf("commit-fix worker prompt missing commit error detail: %s", invocation.Prompt)
				}
			}
			return AgentResult{CriticalIssues: false}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if commitFixWorkerCalls != 1 {
		t.Fatalf("expected one commit-fix worker call, got %d", commitFixWorkerCalls)
	}
	if commitCalls < 2 {
		t.Fatalf("expected at least two commit attempts, got %d", commitCalls)
	}
}

func TestExecutePlan_FailsTaskCommitAfterFiveWorkerFixPasses(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: approved
---

# Test Spec

## Task List

### Spec Updates and Section Delivery

- [ ] First task
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	info := &specpkg.SpecInfo{Number: 7, Path: specPath, Status: StatusApproved}
	plan := Plan{
		Sections: []RemainingSection{{
			Name:  "Spec Updates and Section Delivery",
			Tasks: []specpkg.Task{{Text: "First task", Section: "Spec Updates and Section Delivery"}},
		}},
	}

	currentBranch := "main"
	commitFixWorkerCalls := 0

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return currentBranch, nil },
		createBranch: func(dir, branchName string) error {
			currentBranch = branchName
			return nil
		},
		branchExists: func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:     func(dir string) error { return nil },
		commit: func(dir, message string) error {
			if message == "feat: complete spec 007 task: First task" {
				return os.ErrPermission
			}
			return nil
		},
		pushBranch: func(dir, branchName string) error { return nil },
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleWorker && invocation.TaskText == "First task" && invocation.Attempt > maxWorkerPassesPerTask {
				commitFixWorkerCalls++
			}
			return AgentResult{CriticalIssues: false}, nil
		},
	})
	if err == nil {
		t.Fatal("expected error after exhausting commit-failure fix passes")
	}
	if !strings.Contains(err.Error(), "after 5 commit-failure fix passes") {
		t.Fatalf("unexpected error: %v", err)
	}
	if commitFixWorkerCalls != maxCommitFixPassesPerTask {
		t.Fatalf("expected %d commit-fix worker calls, got %d", maxCommitFixPassesPerTask, commitFixWorkerCalls)
	}
}
