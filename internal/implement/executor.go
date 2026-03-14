package implement

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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
	lastSectionParentBranch := initialBranch

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
		lastSectionParentBranch = parentBranch
		parentBranch = branchName
	}

	if len(plan.Sections) > 0 {
		if err := executeFinalCleanupPass(workDir, info, plan.Sections, backend, parentBranch, lastSectionParentBranch, printf, deps); err != nil {
			return err
		}
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

func executeFinalCleanupPass(workDir string, info *specpkg.SpecInfo, sections []RemainingSection, backend, currentBranch, parentBranch string, printf PrintfFunc, deps executeDeps) error {
	if printf != nil {
		printf("Final cleanup review started\n")
	}

	reviewPrompt, err := buildFinalCleanupReviewPrompt(info.Path, sections, currentBranch, parentBranch)
	if err != nil {
		return fmt.Errorf("failed to build final cleanup review prompt: %w", err)
	}

	reviewResult, err := deps.invokeAgent(AgentInvocation{
		Backend:     backend,
		Role:        AgentRoleReviewer,
		SpecPath:    info.Path,
		SectionName: "Final Cleanup",
		Attempt:     1,
		Prompt:      reviewPrompt,
	})
	if err != nil {
		return fmt.Errorf("final cleanup review failed: %w", err)
	}

	if printf != nil {
		printf("Final cleanup review completed\n")
		printReviewFeedback(printf, "final cleanup", 1, reviewResult.Output)
		printf("Final cleanup worker pass started\n")
	}

	workerPrompt, err := buildFinalCleanupWorkerPrompt(info.Path, sections, currentBranch, reviewResult.Output)
	if err != nil {
		return fmt.Errorf("failed to build final cleanup worker prompt: %w", err)
	}

	if _, err := deps.invokeAgent(AgentInvocation{
		Backend:     backend,
		Role:        AgentRoleWorker,
		SpecPath:    info.Path,
		SectionName: "Final Cleanup",
		Attempt:     1,
		Prompt:      workerPrompt,
	}); err != nil {
		return fmt.Errorf("final cleanup worker pass failed: %w", err)
	}

	if printf != nil {
		printf("Final cleanup worker pass completed\n")
	}

	hasChanges, err := deps.hasUncommittedChanges(workDir)
	if err != nil {
		return fmt.Errorf("failed to inspect final cleanup changes: %w", err)
	}
	if !hasChanges {
		if printf != nil {
			printf("Final cleanup produced no changes; skipping cleanup commit\n")
		}
		return nil
	}

	if err := deps.stageAll(workDir); err != nil {
		return fmt.Errorf("failed to stage final cleanup changes: %w", err)
	}

	if err := deps.commit(workDir, finalCleanupCommitMessage(info.Number)); err != nil {
		return fmt.Errorf("failed to commit final cleanup changes: %w", err)
	}

	if printf != nil {
		printf("Final cleanup commit created\n")
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
	codexOutputPath := ""
	if invocation.Backend == BackendCodex {
		file, err := os.CreateTemp("", "specture-codex-last-message-*.txt")
		if err != nil {
			return AgentResult{}, fmt.Errorf("failed to create codex output file: %w", err)
		}
		codexOutputPath = file.Name()
		if err := file.Close(); err != nil {
			return AgentResult{}, fmt.Errorf("failed to close codex output file: %w", err)
		}
		defer os.Remove(codexOutputPath)
	}

	args, err := backendInvocationArgs(invocation, codexOutputPath)
	if err != nil {
		return AgentResult{}, err
	}

	cmd := exec.Command(invocation.Backend, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		if detail == "" {
			return AgentResult{}, fmt.Errorf("%s agent invocation failed: %w", invocation.Role, err)
		}

		return AgentResult{}, fmt.Errorf("%s agent invocation failed: %w: %s", invocation.Role, err, detail)
	}

	var outputText string
	switch invocation.Backend {
	case BackendOpencode:
		outputText, err = parseOpencodeRunJSONOutput(stdout.String())
		if err != nil {
			return AgentResult{}, fmt.Errorf("failed to parse opencode JSON output: %w", err)
		}
	case BackendCodex:
		content, readErr := os.ReadFile(codexOutputPath)
		if readErr != nil {
			return AgentResult{}, fmt.Errorf("failed to read codex final message output: %w", readErr)
		}
		outputText = string(content)
	default:
		return AgentResult{}, fmt.Errorf("unsupported agent backend %q for invocation", invocation.Backend)
	}

	return AgentResult{
		Output:         outputText,
		CriticalIssues: strings.HasPrefix(strings.TrimSpace(outputText), "REVIEW_CRITICAL:"),
	}, nil
}

func parseOpencodeRunJSONOutput(stdout string) (string, error) {
	type opencodeEvent struct {
		Type string `json:"type"`
		Part struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"part"`
	}

	var parts []string
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event opencodeEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return "", fmt.Errorf("invalid JSONL event %q: %w", line, err)
		}

		if event.Type == "text" && event.Part.Text != "" {
			parts = append(parts, event.Part.Text)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read opencode JSON output: %w", err)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("opencode returned no text events")
	}

	return strings.Join(parts, ""), nil
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

func buildFinalCleanupReviewPrompt(specPath string, sections []RemainingSection, currentBranch, parentBranch string) (string, error) {
	promptTemplate, err := templatespkg.GetImplementCleanupReviewPromptTemplate()
	if err != nil {
		return "", err
	}

	return templatepkg.RenderTemplate(promptTemplate, struct {
		SpecPath      string
		CurrentBranch string
		ParentBranch  string
		Sections      []string
	}{
		SpecPath:      specPath,
		CurrentBranch: currentBranch,
		ParentBranch:  parentBranch,
		Sections:      sectionNames(sections),
	})
}

func buildFinalCleanupWorkerPrompt(specPath string, sections []RemainingSection, currentBranch, reviewOutput string) (string, error) {
	promptTemplate, err := templatespkg.GetImplementCleanupWorkerPromptTemplate()
	if err != nil {
		return "", err
	}

	return templatepkg.RenderTemplate(promptTemplate, struct {
		SpecPath      string
		CurrentBranch string
		Sections      []string
		ReviewOutput  string
	}{
		SpecPath:      specPath,
		CurrentBranch: currentBranch,
		Sections:      sectionNames(sections),
		ReviewOutput:  strings.TrimSpace(reviewOutput),
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

func finalCleanupCommitMessage(specNumber int) string {
	return fmt.Sprintf("refactor: final cleanup for spec %03d", specNumber)
}

func sectionNames(sections []RemainingSection) []string {
	names := make([]string, 0, len(sections))
	for _, section := range sections {
		names = append(names, displaySectionName(section.Name))
	}

	return names
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
