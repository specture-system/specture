package implement

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestSectionBranchName_Deterministic(t *testing.T) {
	branch := SectionBranchName(7, "Branch and Task Execution", 2)
	if branch != "implement/007-02-branch-and-task-execution" {
		t.Fatalf("unexpected branch name: %s", branch)
	}

	branchAgain := SectionBranchName(7, "Branch and Task Execution", 2)
	if branchAgain != branch {
		t.Fatalf("expected deterministic branch name, got %q and %q", branch, branchAgain)
	}
}

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

func TestExecuteTaskWithReview_InvokesWorkerAndReviewerWithContext(t *testing.T) {
	task := specpkg.Task{
		Text:    "Implement section branch creation",
		Subtree: "- [ ] Implement section branch creation\n  - [ ] Include nested checkbox context\n    - Nested bullet detail",
	}
	invocations := make([]AgentInvocation, 0, 2)

	err := ExecuteTaskWithReview("specs/agent.md", "Branch and Task Execution", BackendOpencode, task, nil, func(invocation AgentInvocation) (AgentResult, error) {
		invocations = append(invocations, invocation)
		if invocation.Role == AgentRoleReviewer {
			return AgentResult{CriticalIssues: false}, nil
		}
		return AgentResult{}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(invocations))
	}

	if invocations[0].Role != AgentRoleWorker {
		t.Fatalf("expected first invocation to be worker, got %s", invocations[0].Role)
	}
	if invocations[1].Role != AgentRoleReviewer {
		t.Fatalf("expected second invocation to be reviewer, got %s", invocations[1].Role)
	}

	if !strings.Contains(invocations[0].Prompt, "Spec Path: specs/agent.md") {
		t.Fatalf("worker prompt missing spec path: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Section: Branch and Task Execution") {
		t.Fatalf("worker prompt missing section: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Task: Implement section branch creation") {
		t.Fatalf("worker prompt missing task text: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Task Subtree:") {
		t.Fatalf("worker prompt missing task subtree header: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "  - [ ] Include nested checkbox context") {
		t.Fatalf("worker prompt missing nested checkbox context: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Preserve existing accepted changes already present in the branch") {
		t.Fatalf("worker prompt missing branch-baseline instruction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Do not edit the spec file") {
		t.Fatalf("worker prompt missing spec-edit restriction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Do not create commits") {
		t.Fatalf("worker prompt missing commit restriction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[1].Prompt, "Task Subtree:") {
		t.Fatalf("review prompt missing task subtree header: %s", invocations[1].Prompt)
	}
}

func TestExecuteTaskWithReview_RetriesWorkerUpToThreePassesOnCriticalReview(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}
	workerCalls := 0
	reviewCalls := 0

	err := ExecuteTaskWithReview("specs/agent.md", "Branch and Task Execution", BackendOpencode, task, nil, func(invocation AgentInvocation) (AgentResult, error) {
		if invocation.Role == AgentRoleWorker {
			workerCalls++
			return AgentResult{}, nil
		}

		reviewCalls++
		if reviewCalls < 3 {
			return AgentResult{CriticalIssues: true}, nil
		}
		return AgentResult{CriticalIssues: false}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if workerCalls != 3 {
		t.Fatalf("expected 3 worker calls, got %d", workerCalls)
	}
	if reviewCalls != 3 {
		t.Fatalf("expected 3 review calls, got %d", reviewCalls)
	}
}

func TestExecuteTaskWithReview_PrintsWorkerAndReviewPassProgress(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}
	var logs strings.Builder

	err := ExecuteTaskWithReview(
		"specs/agent.md",
		"Branch and Task Execution",
		BackendOpencode,
		task,
		func(format string, args ...any) {
			logs.WriteString(fmt.Sprintf(format, args...))
		},
		func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleReviewer {
				return AgentResult{CriticalIssues: false}, nil
			}
			return AgentResult{}, nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := logs.String()
	expected := []string{
		"worker pass 1/3 started for task: Implement retry behavior",
		"worker pass 1/3 completed for task: Implement retry behavior",
		"review pass 1/3 started for task: Implement retry behavior",
		"review pass 1/3 completed for task: Implement retry behavior",
		"task review feedback (pass 1):",
		"(no reviewer output)",
		"task accepted after 1 pass(es): Implement retry behavior",
	}
	for _, needle := range expected {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, output)
		}
	}
}

func TestExecuteTaskWithReview_PrintsReviewFeedbackAcrossMultiplePasses(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}
	var logs strings.Builder
	reviewCalls := 0

	err := ExecuteTaskWithReview(
		"specs/agent.md",
		"Branch and Task Execution",
		BackendOpencode,
		task,
		func(format string, args ...any) {
			logs.WriteString(fmt.Sprintf(format, args...))
		},
		func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleReviewer {
				reviewCalls++
				switch reviewCalls {
				case 1:
					return AgentResult{CriticalIssues: true, Output: "REVIEW_CRITICAL: missing tests"}, nil
				case 2:
					return AgentResult{CriticalIssues: true, Output: "REVIEW_CRITICAL: flaky assertion remains"}, nil
				default:
					return AgentResult{CriticalIssues: false, Output: "REVIEW_OK: all critical issues resolved"}, nil
				}
			}

			return AgentResult{}, nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := logs.String()
	expected := []string{
		"task review feedback (pass 1):",
		"REVIEW_CRITICAL: missing tests",
		"task review feedback (pass 2):",
		"REVIEW_CRITICAL: flaky assertion remains",
		"task review feedback (pass 3):",
		"REVIEW_OK: all critical issues resolved",
		"task accepted after 3 pass(es): Implement retry behavior",
	}
	for _, needle := range expected {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected output to contain %q, got:\n%s", needle, output)
		}
	}
}

func TestExecuteTaskWithReview_FailsAfterThreeCriticalReviews(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}

	err := ExecuteTaskWithReview("specs/agent.md", "Branch and Task Execution", BackendOpencode, task, nil, func(invocation AgentInvocation) (AgentResult, error) {
		if invocation.Role == AgentRoleReviewer {
			return AgentResult{CriticalIssues: true}, nil
		}
		return AgentResult{}, nil
	})
	if err == nil {
		t.Fatal("expected error after three critical reviews")
	}

	if !strings.Contains(err.Error(), "after 3 worker passes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecutePlan_UpdatesApprovedSpecAndCommitsAcceptedTasks(t *testing.T) {
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

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch:          func(dir, branchName string) error { return nil },
		branchExists:          func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:              func(dir string) error { return nil },
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
	if !strings.Contains(updatedText, "status: in-progress") {
		t.Fatalf("expected spec status to transition to in-progress, got:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [x] Add failing tests for in-progress transition") {
		t.Fatalf("expected first task to be checked off, got:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [x] Implement deterministic task commits") {
		t.Fatalf("expected second task to be checked off, got:\n%s", updatedText)
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

	if len(pushes) != 1 || pushes[0] != "implement/007-01-spec-updates-and-section-delivery" {
		t.Fatalf("unexpected pushes: %v", pushes)
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

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) {
			dirtyChecks++
			if dirtyChecks == 1 {
				return false, nil
			}
			return true, nil
		},
		getCurrentBranch: func(dir string) (string, error) { return "main", nil },
		createBranch:     func(dir, branchName string) error { return nil },
		branchExists:     func(dir, branchName string) (bool, error) { return false, nil },
		stageAll:         func(dir string) error { return nil },
		commit: func(dir, message string) error {
			commits = append(commits, message)
			return nil
		},
		pushBranch: func(dir, branchName string) error { return nil },
		invokeAgent: func(invocation AgentInvocation) (AgentResult, error) {
			if invocation.Role == AgentRoleWorker && invocation.TaskText == "" {
				sectionWorkerCalls++
				if !strings.Contains(invocation.Prompt, "REVIEW_CRITICAL: integration broke") {
					t.Fatalf("section worker prompt missing review summary: %s", invocation.Prompt)
				}
				return AgentResult{}, nil
			}

			if invocation.Role == AgentRoleReviewer && invocation.TaskText == "" {
				sectionReviewCalls++
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

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d (%v)", len(commits), commits)
	}
	if commits[1] != "fix: address spec 007 section 01 review: Spec Updates and Section Delivery" {
		t.Fatalf("unexpected section review commit message: %q", commits[1])
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

	err := executePlanWithDeps(tmpDir, info, plan, BackendOpencode, nil, executeDeps{
		hasUncommittedChanges: func(dir string) (bool, error) { return false, nil },
		getCurrentBranch:      func(dir string) (string, error) { return "main", nil },
		createBranch: func(dir, branchName string) error {
			createdBranches++
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
