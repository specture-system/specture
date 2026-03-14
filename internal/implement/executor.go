package implement

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	gitpkg "github.com/specture-system/specture/internal/git"
	specpkg "github.com/specture-system/specture/internal/spec"
	templatepkg "github.com/specture-system/specture/internal/template"
	templatespkg "github.com/specture-system/specture/internal/templates"
)

const maxWorkerPassesPerTask = 3
const maxSectionReviewPasses = 2

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

type sectionBranchContext struct {
	CurrentBranch         string
	ExpectedSectionBranch string
	ParentBranch          string
}

type executeDeps struct {
	hasUncommittedChanges func(dir string) (bool, error)
	changedFiles          func(dir string) ([]string, error)
	getCurrentBranch      func(dir string) (string, error)
	createBranch          func(dir, branchName string) error
	branchExists          func(dir, branchName string) (bool, error)
	stageAll              func(dir string) error
	commit                func(dir, message string) error
	pushBranch            func(dir, branchName string) error
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
	return executeTaskWithReviewWithContext("", specPath, sectionName, backend, task, printf, nil, invokeAgent)
}

func executeTaskWithReviewWithContext(workDir, specPath, sectionName, backend string, task specpkg.Task, printf PrintfFunc, changedFiles func(dir string) ([]string, error), invokeAgent func(invocation AgentInvocation) (AgentResult, error)) error {
	priorCriticalReviewOutput := ""

	for pass := 1; pass <= maxWorkerPassesPerTask; pass++ {
		if printf != nil {
			printf("    worker pass %d/%d started for task: %s\n", pass, maxWorkerPassesPerTask, task.Text)
		}

		taskChangedFiles := []string{}
		if changedFiles != nil {
			files, err := changedFiles(workDir)
			if err != nil {
				return fmt.Errorf("failed to collect changed files for task %q before worker pass %d: %w", task.Text, pass, err)
			}
			taskChangedFiles = files
		}

		workerPrompt, err := buildWorkerPrompt(specPath, sectionName, task, priorCriticalReviewOutput, taskChangedFiles)
		if err != nil {
			return fmt.Errorf("failed to build worker prompt for task %q: %w", task.Text, err)
		}
		_, err = invokeAgent(AgentInvocation{
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
		if printf != nil {
			printf("    worker pass %d/%d completed for task: %s\n", pass, maxWorkerPassesPerTask, task.Text)
			printf("    review pass %d/%d started for task: %s\n", pass, maxWorkerPassesPerTask, task.Text)
		}

		taskChangedFiles = []string{}
		if changedFiles != nil {
			files, err := changedFiles(workDir)
			if err != nil {
				return fmt.Errorf("failed to collect changed files for task %q on pass %d: %w", task.Text, pass, err)
			}
			taskChangedFiles = files
		}

		reviewPrompt, err := buildReviewPrompt(specPath, sectionName, task, taskChangedFiles)
		if err != nil {
			return fmt.Errorf("failed to build review prompt for task %q: %w", task.Text, err)
		}
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
		if printf != nil {
			printf("    review pass %d/%d completed for task: %s\n", pass, maxWorkerPassesPerTask, task.Text)
			printReviewFeedback(printf, "task", pass, reviewResult.Output)
		}

		if !reviewResult.CriticalIssues {
			if printf != nil {
				printf("  task accepted after %d pass(es): %s\n", pass, task.Text)
			}
			return nil
		}
		priorCriticalReviewOutput = reviewResult.Output

		if printf != nil {
			printf("  critical issues found for task %q on pass %d; retrying\n", task.Text, pass)
		}
	}

	return fmt.Errorf("task %q failed review after %d worker passes due to critical issues", task.Text, maxWorkerPassesPerTask)
}

func executePlanWithDeps(workDir string, info *specpkg.SpecInfo, plan Plan, backend string, printf PrintfFunc, deps executeDeps) error {
	deps = withExecuteDepDefaults(deps)
	sectionOrderByName := specpkg.TaskListSectionOrders(info.Path)
	currentStatus := info.Status
	initialBranch, err := deps.getCurrentBranch(workDir)
	if err != nil {
		return fmt.Errorf("failed to determine current branch before execution: %w", err)
	}
	parentBranch := initialBranch

	for idx, section := range plan.Sections {
		sectionNumber := idx + 1
		if order, ok := sectionOrderByName[section.Name]; ok {
			sectionNumber = order
		}

		branchName := SectionBranchName(info.Number, section.Name, sectionNumber)

		if err := ensureSectionBranch(workDir, branchName, deps); err != nil {
			return err
		}
		currentBranch, err := deps.getCurrentBranch(workDir)
		if err != nil {
			return fmt.Errorf("failed to determine current branch for section %q: %w", section.Name, err)
		}

		if printf != nil {
			printf("Section %d/%d: %s\n", idx+1, len(plan.Sections), displaySectionName(section.Name))
			printf("  branch: %s\n", branchName)
		}

		if len(section.Tasks) == 0 {
			continue
		}

		for taskIdx, task := range section.Tasks {
			if printf != nil {
				printf("  running task %d/%d: %s\n", taskIdx+1, len(section.Tasks), task.Text)
			}

			if err := executeTaskWithReviewWithContext(workDir, info.Path, section.Name, backend, task, printf, deps.changedFiles, deps.invokeAgent); err != nil {
				return err
			}

			nextStatus := ""
			if currentStatus == StatusApproved {
				nextStatus = StatusInProgress
			}

			if err := applyTaskProgress(info.Path, section.Name, task.Text, nextStatus); err != nil {
				return fmt.Errorf("failed to update spec progress for task %q: %w", task.Text, err)
			}

			if err := deps.stageAll(workDir); err != nil {
				return fmt.Errorf("failed to stage accepted task %q: %w", task.Text, err)
			}

			if err := deps.commit(workDir, taskCommitMessage(info.Number, task.Text)); err != nil {
				return fmt.Errorf("failed to commit accepted task %q: %w", task.Text, err)
			}

			if nextStatus == StatusInProgress {
				currentStatus = StatusInProgress
				info.Status = StatusInProgress
			}
		}

		branchContext := sectionBranchContext{
			CurrentBranch:         currentBranch,
			ExpectedSectionBranch: branchName,
			ParentBranch:          parentBranch,
		}
		if err := executeSectionReview(workDir, info, backend, section, sectionNumber, branchContext, printf, deps); err != nil {
			return err
		}

		if err := deps.pushBranch(workDir, branchName); err != nil {
			return fmt.Errorf("failed to push completed section branch %q: %w", branchName, err)
		}
		parentBranch = branchName
	}

	return nil
}

func defaultExecuteDeps() executeDeps {
	return executeDeps{
		hasUncommittedChanges: gitpkg.HasUncommittedChanges,
		changedFiles:          gitpkg.ChangedFiles,
		getCurrentBranch:      gitpkg.GetCurrentBranch,
		createBranch:          gitpkg.CreateBranch,
		branchExists:          gitpkg.BranchExists,
		stageAll:              gitpkg.StageAll,
		commit:                gitpkg.Commit,
		pushBranch:            gitpkg.PushBranch,
		invokeAgent:           invokeAgentCLI,
	}
}

func withExecuteDepDefaults(deps executeDeps) executeDeps {
	if deps.stageAll == nil {
		deps.stageAll = func(dir string) error { return nil }
	}
	if deps.changedFiles == nil {
		deps.changedFiles = func(dir string) ([]string, error) { return []string{}, nil }
	}
	if deps.commit == nil {
		deps.commit = func(dir, message string) error { return nil }
	}
	if deps.pushBranch == nil {
		deps.pushBranch = func(dir, branchName string) error { return nil }
	}

	return deps
}

func executeSectionReview(workDir string, info *specpkg.SpecInfo, backend string, section RemainingSection, sectionNumber int, branchContext sectionBranchContext, printf PrintfFunc, deps executeDeps) error {
	for pass := 1; pass <= maxSectionReviewPasses; pass++ {
		currentBranch, err := deps.getCurrentBranch(workDir)
		if err != nil {
			return fmt.Errorf("failed to determine current branch before section review for %q: %w", section.Name, err)
		}
		if currentBranch != branchContext.ExpectedSectionBranch {
			return fmt.Errorf("section review for %q requires current branch %q, but got %q", section.Name, branchContext.ExpectedSectionBranch, currentBranch)
		}
		branchContext.CurrentBranch = currentBranch

		if printf != nil {
			printf("  section review pass %d/%d started: %s\n", pass, maxSectionReviewPasses, section.Name)
		}

		reviewPrompt, err := buildSectionReviewPrompt(info.Path, section.Name, section.Tasks, branchContext)
		if err != nil {
			return fmt.Errorf("failed to build section review prompt for %q: %w", section.Name, err)
		}

		reviewResult, err := deps.invokeAgent(AgentInvocation{
			Backend:     backend,
			Role:        AgentRoleReviewer,
			SpecPath:    info.Path,
			SectionName: section.Name,
			Attempt:     pass,
			Prompt:      reviewPrompt,
		})
		if err != nil {
			return fmt.Errorf("section review pass %d failed for %q: %w", pass, section.Name, err)
		}
		if printf != nil {
			printf("  section review pass %d/%d completed: %s\n", pass, maxSectionReviewPasses, section.Name)
			printReviewFeedback(printf, "section", pass, reviewResult.Output)
		}

		if !reviewResult.CriticalIssues {
			if printf != nil {
				printf("  section accepted after %d review pass(es): %s\n", pass, section.Name)
			}
			return nil
		}

		if pass == maxSectionReviewPasses {
			return fmt.Errorf("section %q failed review after %d passes due to critical issues", section.Name, maxSectionReviewPasses)
		}

		if printf != nil {
			printf("  section review found critical issues for %q; retrying once\n", section.Name)
			printf("  section worker retry started: %s\n", section.Name)
		}

		workerPrompt, err := buildSectionWorkerPrompt(info.Path, section.Name, section.Tasks, reviewResult.Output)
		if err != nil {
			return fmt.Errorf("failed to build section worker prompt for %q: %w", section.Name, err)
		}

		if _, err := deps.invokeAgent(AgentInvocation{
			Backend:     backend,
			Role:        AgentRoleWorker,
			SpecPath:    info.Path,
			SectionName: section.Name,
			Attempt:     pass + 1,
			Prompt:      workerPrompt,
		}); err != nil {
			return fmt.Errorf("section worker retry failed for %q: %w", section.Name, err)
		}
		if printf != nil {
			printf("  section worker retry completed: %s\n", section.Name)
		}

		hasChanges, err := deps.hasUncommittedChanges(workDir)
		if err != nil {
			return fmt.Errorf("failed to inspect section retry changes for %q: %w", section.Name, err)
		}
		if !hasChanges {
			continue
		}

		if err := deps.stageAll(workDir); err != nil {
			return fmt.Errorf("failed to stage section retry changes for %q: %w", section.Name, err)
		}

		if err := deps.commit(workDir, sectionReviewCommitMessage(info.Number, section.Name, sectionNumber)); err != nil {
			return fmt.Errorf("failed to commit section retry changes for %q: %w", section.Name, err)
		}
	}

	return nil
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

func buildWorkerPrompt(specPath, sectionName string, task specpkg.Task, reviewOutput string, changedFiles []string) (string, error) {
	return renderPromptTemplate(templatespkg.GetImplementWorkerPromptTemplate, specPath, sectionName, task, reviewOutput, changedFiles)
}

func buildReviewPrompt(specPath, sectionName string, task specpkg.Task, changedFiles []string) (string, error) {
	return renderPromptTemplate(templatespkg.GetImplementReviewPromptTemplate, specPath, sectionName, task, "", changedFiles)
}

func buildSectionReviewPrompt(specPath, sectionName string, tasks []specpkg.Task, branchContext sectionBranchContext) (string, error) {
	promptTemplate, err := templatespkg.GetImplementSectionReviewPromptTemplate()
	if err != nil {
		return "", err
	}

	return templatepkg.RenderTemplate(promptTemplate, struct {
		SpecPath              string
		SectionName           string
		Tasks                 []string
		CurrentBranch         string
		ExpectedSectionBranch string
		ParentBranch          string
	}{
		SpecPath:              specPath,
		SectionName:           displaySectionName(sectionName),
		Tasks:                 taskTexts(tasks),
		CurrentBranch:         branchContext.CurrentBranch,
		ExpectedSectionBranch: branchContext.ExpectedSectionBranch,
		ParentBranch:          branchContext.ParentBranch,
	})
}

func buildSectionWorkerPrompt(specPath, sectionName string, tasks []specpkg.Task, reviewOutput string) (string, error) {
	promptTemplate, err := templatespkg.GetImplementSectionWorkerPromptTemplate()
	if err != nil {
		return "", err
	}

	return templatepkg.RenderTemplate(promptTemplate, struct {
		SpecPath     string
		SectionName  string
		Tasks        []string
		ReviewOutput string
	}{
		SpecPath:     specPath,
		SectionName:  displaySectionName(sectionName),
		Tasks:        taskTexts(tasks),
		ReviewOutput: strings.TrimSpace(reviewOutput),
	})
}

func renderPromptTemplate(loadTemplate func() (string, error), specPath, sectionName string, task specpkg.Task, reviewOutput string, changedFiles []string) (string, error) {
	promptTemplate, err := loadTemplate()
	if err != nil {
		return "", err
	}

	return templatepkg.RenderTemplate(promptTemplate, struct {
		SpecPath     string
		SectionName  string
		TaskText     string
		TaskSubtree  string
		ReviewOutput string
		ChangedFiles []string
	}{
		SpecPath:     specPath,
		SectionName:  displaySectionName(sectionName),
		TaskText:     task.Text,
		TaskSubtree:  task.Subtree,
		ReviewOutput: strings.TrimSpace(reviewOutput),
		ChangedFiles: changedFiles,
	})
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

func taskCommitMessage(specNumber int, taskText string) string {
	return fmt.Sprintf("feat: complete spec %03d task: %s", specNumber, taskText)
}

func sectionReviewCommitMessage(specNumber int, sectionName string, sectionNumber int) string {
	return fmt.Sprintf("fix: address spec %03d section %02d review: %s", specNumber, sectionNumber, displaySectionName(sectionName))
}

func taskTexts(tasks []specpkg.Task) []string {
	texts := make([]string, 0, len(tasks))
	for _, task := range tasks {
		texts = append(texts, task.Text)
	}

	return texts
}

func printReviewFeedback(printf PrintfFunc, scope string, pass int, output string) {
	trimmed := strings.TrimSpace(output)
	printf("    %s review feedback (pass %d):\n", scope, pass)
	if trimmed == "" {
		printf("      (no reviewer output)\n")
		return
	}

	for _, line := range strings.Split(trimmed, "\n") {
		printf("      %s\n", line)
	}
}
