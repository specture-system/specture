package implement

import (
	"fmt"
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
