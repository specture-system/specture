package implement

import (
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
	if !strings.Contains(invocations[0].Prompt, "Preserve existing accepted changes already present in the branch") {
		t.Fatalf("worker prompt missing branch-baseline instruction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Do not edit the spec file") {
		t.Fatalf("worker prompt missing spec-edit restriction: %s", invocations[0].Prompt)
	}
	if !strings.Contains(invocations[0].Prompt, "Do not create commits") {
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
