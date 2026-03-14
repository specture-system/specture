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
	if !strings.Contains(invocations[0].Prompt, "Current changed files in working tree:\n- (none detected)") {
		t.Fatalf("worker prompt missing changed files fallback: %s", invocations[0].Prompt)
	}
	if strings.Contains(invocations[0].Prompt, "Prior critical review findings:") {
		t.Fatalf("worker prompt should not include prior critical findings on first pass: %s", invocations[0].Prompt)
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
	if !strings.Contains(invocations[1].Prompt, "Files changed in current task pass:\n- (none detected)") {
		t.Fatalf("review prompt missing changed files fallback: %s", invocations[1].Prompt)
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

func TestExecuteTaskWithReview_PassesPriorCriticalReviewFeedbackToNextWorkerPass(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}
	var workerPrompts []string
	reviewCalls := 0

	err := ExecuteTaskWithReview("specs/agent.md", "Branch and Task Execution", BackendOpencode, task, nil, func(invocation AgentInvocation) (AgentResult, error) {
		if invocation.Role == AgentRoleWorker {
			workerPrompts = append(workerPrompts, invocation.Prompt)
			return AgentResult{}, nil
		}

		reviewCalls++
		if reviewCalls == 1 {
			return AgentResult{CriticalIssues: true, Output: "REVIEW_CRITICAL: missing tests"}, nil
		}
		return AgentResult{CriticalIssues: false, Output: "REVIEW_OK"}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(workerPrompts) != 2 {
		t.Fatalf("expected 2 worker prompts, got %d", len(workerPrompts))
	}
	if strings.Contains(workerPrompts[0], "Prior critical review findings:") {
		t.Fatalf("first worker prompt should not include prior critical findings: %s", workerPrompts[0])
	}
	if !strings.Contains(workerPrompts[1], "Prior critical review findings:\nREVIEW_CRITICAL: missing tests") {
		t.Fatalf("second worker prompt missing prior critical findings context: %s", workerPrompts[1])
	}
}

func TestParseOpencodeRunJSONOutput_ExtractsTextEventsOnly(t *testing.T) {
	raw := strings.Join([]string{
		`{"type":"step_start","part":{"type":"step-start"}}`,
		`{"type":"text","part":{"type":"text","text":"REVIEW_"}}`,
		`{"type":"text","part":{"type":"text","text":"OK"}}`,
		`{"type":"step_finish","part":{"type":"step-finish"}}`,
	}, "\n")

	got, err := parseOpencodeRunJSONOutput(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "REVIEW_OK" {
		t.Fatalf("expected REVIEW_OK, got %q", got)
	}
}

func TestParseOpencodeRunJSONOutput_RejectsInvalidJSONEvent(t *testing.T) {
	raw := strings.Join([]string{
		`{"type":"text","part":{"type":"text","text":"REVIEW_OK"}}`,
		`not-json`,
	}, "\n")

	_, err := parseOpencodeRunJSONOutput(raw)
	if err == nil {
		t.Fatal("expected invalid JSONL error")
	}

	if !strings.Contains(err.Error(), "invalid JSONL event") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseOpencodeRunJSONOutput_RejectsWhenNoTextEvents(t *testing.T) {
	raw := strings.Join([]string{
		`{"type":"step_start","part":{"type":"step-start"}}`,
		`{"type":"step_finish","part":{"type":"step-finish"}}`,
	}, "\n")

	_, err := parseOpencodeRunJSONOutput(raw)
	if err == nil {
		t.Fatal("expected missing text event error")
	}

	if !strings.Contains(err.Error(), "no text events") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteTaskWithReview_WithContextPassesChangedFilesToReviewer(t *testing.T) {
	task := specpkg.Task{Text: "Implement retry behavior"}
	invocations := make([]AgentInvocation, 0, 2)

	err := executeTaskWithReviewWithContext(
		"/tmp/repo",
		"specs/agent.md",
		"Branch and Task Execution",
		BackendOpencode,
		task,
		nil,
		func(dir string) ([]string, error) {
			if dir != "/tmp/repo" {
				t.Fatalf("unexpected work dir: %s", dir)
			}
			return []string{"cmd/implement.go", "internal/implement/executor.go"}, nil
		},
		func(invocation AgentInvocation) (AgentResult, error) {
			invocations = append(invocations, invocation)
			if invocation.Role == AgentRoleReviewer {
				return AgentResult{CriticalIssues: false}, nil
			}
			return AgentResult{}, nil
		},
	)
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
	if !strings.Contains(invocations[0].Prompt, "Current changed files in working tree:") {
		t.Fatalf("worker prompt missing changed-files header: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "- cmd/implement.go") {
		t.Fatalf("worker prompt missing first changed file: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "- internal/implement/executor.go") {
		t.Fatalf("worker prompt missing second changed file: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[1].Prompt, "Files changed in current task pass:") {
		t.Fatalf("review prompt missing changed-files header: %s", invocations[1].Prompt)
	}
	if !strings.Contains(invocations[1].Prompt, "- cmd/implement.go") {
		t.Fatalf("review prompt missing first changed file: %s", invocations[1].Prompt)
	}
	if !strings.Contains(invocations[1].Prompt, "- internal/implement/executor.go") {
		t.Fatalf("review prompt missing second changed file: %s", invocations[1].Prompt)
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

func TestExecutePlan_MarksSpecCompletedAfterAllRemainingTasksFinish(t *testing.T) {
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
	if !strings.Contains(updatedText, "status: completed") {
		t.Fatalf("expected spec status to transition to completed when all remaining tasks are done, got:\n%s", updatedText)
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
