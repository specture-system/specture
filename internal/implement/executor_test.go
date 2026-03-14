package implement

import (
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestSectionBranchName_Deterministic(t *testing.T) {
	branch := SectionBranchName(7, "Branch and Task Execution", 1)
	if branch != "implement/007-02-branch-and-task-execution" {
		t.Fatalf("unexpected branch name: %s", branch)
	}

	branchAgain := SectionBranchName(7, "Branch and Task Execution", 1)
	if branchAgain != branch {
		t.Fatalf("expected deterministic branch name, got %q and %q", branch, branchAgain)
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
	task := specpkg.Task{Text: "Implement section branch creation"}
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
	if !strings.Contains(invocations[0].Prompt, "must not edit the spec file") {
		t.Fatalf("worker prompt missing spec-edit restriction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "must not create commits") {
		t.Fatalf("worker prompt missing commit restriction: %s", invocations[0].Prompt)
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

func TestBackendInvocationArgs_UsesBackendSpecificNonInteractiveSubcommands(t *testing.T) {
	tests := []struct {
		name      string
		backend   string
		wantFirst string
	}{
		{name: "opencode uses run", backend: BackendOpencode, wantFirst: "run"},
		{name: "codex uses exec", backend: BackendCodex, wantFirst: "exec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := backendInvocationArgs(AgentInvocation{Backend: tt.backend, Prompt: "hello"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d", len(args))
			}

			if args[0] != tt.wantFirst {
				t.Fatalf("expected first arg %q, got %q", tt.wantFirst, args[0])
			}

			if args[1] != "hello" {
				t.Fatalf("expected prompt as second arg, got %q", args[1])
			}
		})
	}
}

func TestBackendInvocationArgs_RejectsUnsupportedBackend(t *testing.T) {
	_, err := backendInvocationArgs(AgentInvocation{Backend: "other", Prompt: "hello"})
	if err == nil {
		t.Fatal("expected error for unsupported backend")
	}

	if !strings.Contains(err.Error(), "unsupported agent backend") {
		t.Fatalf("unexpected error: %v", err)
	}
}
