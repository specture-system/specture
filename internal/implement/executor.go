package implement

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	gitpkg "github.com/specture-system/specture/internal/git"
	specpkg "github.com/specture-system/specture/internal/spec"
)

const maxWorkerPassesPerTask = 3

type PrintfFunc func(format string, args ...any)

type AgentRole string

const (
	AgentRoleWorker   AgentRole = "worker"
	AgentRoleReviewer AgentRole = "reviewer"
)

type AgentInvocation struct {
	Backend     string
	Role        AgentRole
	SpecPath    string
	SectionName string
	TaskText    string
	Attempt     int
	Prompt      string
}

type AgentResult struct {
	Output         string
	CriticalIssues bool
}

type executeDeps struct {
	hasUncommittedChanges func(dir string) (bool, error)
	getCurrentBranch      func(dir string) (string, error)
	createBranch          func(dir, branchName string) error
	branchExists          func(dir, branchName string) (bool, error)
	invokeAgent           func(invocation AgentInvocation) (AgentResult, error)
}

func ExecutePlan(workDir string, info *specpkg.SpecInfo, plan Plan, backend string, printf PrintfFunc) error {
	return executePlanWithDeps(workDir, info, plan, backend, printf, defaultExecuteDeps())
}

func SectionBranchName(specNumber int, sectionName string, sectionNumber int) string {
	slug := sectionSlug(sectionName)
	if slug == "" {
		slug = "unsectioned"
	}

	return fmt.Sprintf("implement/%03d-%02d-%s", specNumber, sectionNumber, slug)
}

func ExecuteTaskWithReview(specPath, sectionName, backend string, task specpkg.Task, printf PrintfFunc, invokeAgent func(invocation AgentInvocation) (AgentResult, error)) error {
	for pass := 1; pass <= maxWorkerPassesPerTask; pass++ {
		workerPrompt := buildWorkerPrompt(specPath, sectionName, task)
		_, err := invokeAgent(AgentInvocation{
			Backend:     backend,
			Role:        AgentRoleWorker,
			SpecPath:    specPath,
			SectionName: sectionName,
			TaskText:    task.Text,
			Attempt:     pass,
			Prompt:      workerPrompt,
		})
		if err != nil {
			return fmt.Errorf("worker pass %d failed for task %q: %w", pass, task.Text, err)
		}

		reviewPrompt := buildReviewPrompt(specPath, sectionName, task)
		reviewResult, err := invokeAgent(AgentInvocation{
			Backend:     backend,
			Role:        AgentRoleReviewer,
			SpecPath:    specPath,
			SectionName: sectionName,
			TaskText:    task.Text,
			Attempt:     pass,
			Prompt:      reviewPrompt,
		})
		if err != nil {
			return fmt.Errorf("review pass %d failed for task %q: %w", pass, task.Text, err)
		}

		if !reviewResult.CriticalIssues {
			if printf != nil {
				printf("  task accepted after %d pass(es): %s\n", pass, task.Text)
			}
			return nil
		}

		if printf != nil {
			printf("  critical issues found for task %q on pass %d; retrying\n", task.Text, pass)
		}
	}

	return fmt.Errorf("task %q failed review after %d worker passes due to critical issues", task.Text, maxWorkerPassesPerTask)
}

func executePlanWithDeps(workDir string, info *specpkg.SpecInfo, plan Plan, backend string, printf PrintfFunc, deps executeDeps) error {
	sectionOrderByName := specpkg.TaskListSectionOrders(info.Path)

	for idx, section := range plan.Sections {
		sectionNumber := idx + 1
		if order, ok := sectionOrderByName[section.Name]; ok {
			sectionNumber = order
		}

		branchName := SectionBranchName(info.Number, section.Name, sectionNumber)

		if err := ensureSectionBranch(workDir, branchName, deps); err != nil {
			return err
		}

		if printf != nil {
			printf("Section %d/%d: %s\n", idx+1, len(plan.Sections), displaySectionName(section.Name))
			printf("  branch: %s\n", branchName)
		}

		for _, task := range section.Tasks {
			if printf != nil {
				printf("  running task: %s\n", task.Text)
			}

			if err := ExecuteTaskWithReview(info.Path, section.Name, backend, task, printf, deps.invokeAgent); err != nil {
				return err
			}
		}
	}

	return nil
}

func defaultExecuteDeps() executeDeps {
	return executeDeps{
		hasUncommittedChanges: gitpkg.HasUncommittedChanges,
		getCurrentBranch:      gitpkg.GetCurrentBranch,
		createBranch:          gitpkg.CreateBranch,
		branchExists:          gitpkg.BranchExists,
		invokeAgent:           invokeAgentCLI,
	}
}

func ensureSectionBranch(workDir, branchName string, deps executeDeps) error {
	hasChanges, err := deps.hasUncommittedChanges(workDir)
	if err != nil {
		return fmt.Errorf("failed to check worktree status before section branch %q: %w", branchName, err)
	}
	if hasChanges {
		return fmt.Errorf("repository has uncommitted changes; commit or stash before implementing section branch %q", branchName)
	}

	currentBranch, err := deps.getCurrentBranch(workDir)
	if err != nil {
		return fmt.Errorf("failed to determine current branch: %w", err)
	}

	exists, err := deps.branchExists(workDir, branchName)
	if err != nil {
		return fmt.Errorf("failed to inspect section branch %q: %w", branchName, err)
	}

	if exists {
		if currentBranch != branchName {
			return fmt.Errorf("rerun is fail-closed: section branch %q already exists, but current branch is %q; checkout %q to resume", branchName, currentBranch, branchName)
		}
		return nil
	}

	if err := deps.createBranch(workDir, branchName); err != nil {
		return fmt.Errorf("failed to create section branch %q: %w", branchName, err)
	}

	return nil
}

func invokeAgentCLI(invocation AgentInvocation) (AgentResult, error) {
	args, err := backendInvocationArgs(invocation)
	if err != nil {
		return AgentResult{}, err
	}

	cmd := exec.Command(invocation.Backend, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return AgentResult{}, fmt.Errorf("%s agent invocation failed: %w", invocation.Role, err)
	}

	outputText := string(output)
	return AgentResult{
		Output:         outputText,
		CriticalIssues: strings.Contains(outputText, "REVIEW_CRITICAL"),
	}, nil
}

func buildWorkerPrompt(specPath, sectionName string, task specpkg.Task) string {
	return fmt.Sprintf(
		"Implement the following Specture task.\n\nSpec Path: %s\nSection: %s\nTask: %s\n\nConstraints:\n- You must not edit the spec file.\n- You must not create commits.\n- When the task is amenable to automated testing, follow test-driven development: write failing tests first, then implement until the tests pass, then refactor.\n- Focus only on changes needed for this task.",
		specPath,
		displaySectionName(sectionName),
		task.Text,
	)
}

func buildReviewPrompt(specPath, sectionName string, task specpkg.Task) string {
	return fmt.Sprintf(
		"Review the latest implementation changes for this Specture task.\n\nSpec Path: %s\nSection: %s\nTask: %s\n\nVerify that the changes correctly and completely address the task described above. Only block on critical issues (task not fulfilled, correctness, security, data loss, build/test breakage). Ignore nits.\nRespond with REVIEW_CRITICAL if critical issues remain, otherwise REVIEW_OK.",
		specPath,
		displaySectionName(sectionName),
		task.Text,
	)
}

func sectionSlug(sectionName string) string {
	trimmed := strings.TrimSpace(strings.ToLower(sectionName))
	if trimmed == "" {
		return ""
	}

	nonAlphaNum := regexp.MustCompile(`[^a-z0-9]+`)
	slug := nonAlphaNum.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")

	multiDash := regexp.MustCompile(`-+`)
	return multiDash.ReplaceAllString(slug, "-")
}

func displaySectionName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "(unsectioned)"
	}

	return name
}
