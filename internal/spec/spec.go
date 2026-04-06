// Package spec provides shared spec parsing, discovery, and querying.
package spec

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	gmfrontmatter "go.abhg.dev/goldmark/frontmatter"
)

// SpecInfo represents a parsed spec file with all extracted metadata.
type SpecInfo struct {
	Path               string
	Name               string
	Number             int
	FullRef            string
	Status             string
	CurrentTask        string
	CurrentTaskSection string
	CompleteTasks      []Task
	IncompleteTasks    []Task
}

// Task represents a single task item from a spec's task list.
type Task struct {
	Text     string
	Complete bool
	Section  string
	Subtree  string
}

// frontmatter represents the YAML frontmatter of a spec.
type frontmatter struct {
	Status string `yaml:"status"`
	Number *int   `yaml:"number"`
}

// Parse reads and parses a spec file, returning a fully populated SpecInfo.
func Parse(path string) (*SpecInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseContent(path, content)
}

// ParseContent parses spec content and returns a fully populated SpecInfo.
func ParseContent(path string, content []byte) (*SpecInfo, error) {
	info := &SpecInfo{
		Path: path,
	}

	// Parse with goldmark for frontmatter and title
	md := goldmark.New(
		goldmark.WithExtensions(
			&gmfrontmatter.Extender{},
			extension.TaskList,
		),
	)
	ctx := parser.NewContext()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader, parser.WithContext(ctx))

	// Extract frontmatter
	var fmStatus string
	var fmNumber *int
	fmData := gmfrontmatter.Get(ctx)
	if fmData != nil {
		var fm frontmatter
		if err := fmData.Decode(&fm); err == nil {
			fmStatus = fm.Status
			fmNumber = fm.Number
		}
	}

	// Resolve spec number exclusively from frontmatter
	number, err := resolveNumber(fmNumber)
	if err != nil {
		return nil, err
	}
	info.Number = number
	info.FullRef, err = resolveFullRef(path, number)
	if err != nil {
		return nil, err
	}

	// Extract title (first H1 heading)
	info.Name = extractTitle(doc, content)

	// Parse tasks from raw markdown lines
	hasTaskList, completeTasks, incompleteTasks, currentTask, currentTaskSection := parseTasks(content)
	info.CompleteTasks = completeTasks
	info.IncompleteTasks = incompleteTasks
	info.CurrentTask = currentTask
	info.CurrentTaskSection = currentTaskSection

	// Infer status
	info.Status = inferStatus(fmStatus, hasTaskList, len(completeTasks), len(incompleteTasks))

	return info, nil
}

// TaskListSectionOrders returns 1-based section order by section name from the
// spec's ## Task List headings.
func TaskListSectionOrders(path string) map[string]int {
	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]int{}
	}

	lines := strings.Split(string(content), "\n")
	return extractTaskListSectionOrders(lines)
}

// ParseAll finds and parses all specs in the given directory, sorted by ascending number.
func ParseAll(specsDir string) ([]*SpecInfo, error) {
	paths, err := FindAll(specsDir)
	if err != nil {
		return nil, err
	}

	var specs []*SpecInfo
	for _, p := range paths {
		info, err := Parse(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", p, err)
		}
		specs = append(specs, info)
	}

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Number < specs[j].Number
	})

	return specs, nil
}

// FindCurrent returns the first spec with status "in-progress", sorted by ascending number.
// Returns nil if no in-progress spec is found.
func FindCurrent(specs []*SpecInfo) *SpecInfo {
	// Make a sorted copy so we don't mutate the input
	sorted := make([]*SpecInfo, len(specs))
	copy(sorted, specs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Number < sorted[j].Number
	})

	for _, s := range sorted {
		if s.Status == "in-progress" {
			return s
		}
	}
	return nil
}

// resolveNumber determines the spec number exclusively from frontmatter.
// Returns -1 if number is not present in frontmatter.
// Returns an error if the frontmatter number is negative.
func resolveNumber(fmNumber *int) (int, error) {
	if fmNumber == nil {
		return -1, nil
	}
	if *fmNumber < 0 {
		return 0, fmt.Errorf("invalid spec number %d: must be a non-negative integer", *fmNumber)
	}
	return *fmNumber, nil
}

func resolveFullRef(path string, number int) (string, error) {
	if number < 0 {
		return "", nil
	}

	parentSpecPath := filepath.Join(filepath.Dir(filepath.Dir(path)), "SPEC.md")
	if _, err := os.Stat(parentSpecPath); err == nil {
		parentInfo, err := Parse(parentSpecPath)
		if err != nil {
			return "", err
		}
		if parentInfo.FullRef != "" {
			return parentInfo.FullRef + "." + strconv.Itoa(number), nil
		}
		return strconv.Itoa(number), nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect parent spec: %w", err)
	}

	return strconv.Itoa(number), nil
}

// extractNumberFromFilename extracts the spec number from a filename like "003-foo.md".
// Returns -1 if the filename doesn't have a numeric prefix.
// Used only by migration/setup, not by spec parsing.
func extractNumberFromFilename(filename string) int {
	re := regexp.MustCompile(`^(\d{3})-`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return -1
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}
	return n
}

// extractTitle extracts the first H1 heading text from a goldmark document.
func extractTitle(doc ast.Node, source []byte) string {
	var title string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
			var buf bytes.Buffer
			for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
				if textNode, ok := child.(*ast.Text); ok {
					buf.Write(textNode.Segment.Value(source))
				}
			}
			title = buf.String()
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return title
}

// parseTasks parses the raw markdown to extract tasks from the ## Task List section.
// Returns: hasTaskList, completeTasks, incompleteTasks, currentTask, currentTaskSection
func parseTasks(content []byte) (bool, []Task, []Task, string, string) {
	scanner := bufio.NewScanner(bytes.NewReader(content))

	// Collect all lines
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find the ## Task List section
	taskListStart := -1
	taskListEnd := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Task List" {
			taskListStart = i + 1
			continue
		}
		// If we're inside the task list and hit another ## heading, end the section
		if taskListStart >= 0 && i > taskListStart && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			taskListEnd = i
			break
		}
	}

	if taskListStart < 0 {
		return false, nil, nil, "", ""
	}

	var completeTasks []Task
	var incompleteTasks []Task
	currentTask := ""
	currentTaskSection := ""
	currentSection := ""
	// Track line index of first incomplete top-level task for section scan
	firstIncompleteIdx := -1

	for i := taskListStart; i < taskListEnd; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "### ") {
			currentSection = strings.TrimPrefix(trimmed, "### ")
			continue
		}

		if !isTopLevelTaskLine(line) {
			continue
		}

		subtreeEnd := findTaskSubtreeEnd(lines, i, taskListEnd)
		taskSubtree := strings.Join(lines[i:subtreeEnd], "\n")

		switch {
		case strings.HasPrefix(line, "- [x] "):
			taskText := strings.TrimPrefix(line, "- [x] ")
			completeTasks = append(completeTasks, Task{
				Text:     taskText,
				Complete: true,
				Section:  currentSection,
				Subtree:  taskSubtree,
			})
		case strings.HasPrefix(line, "- [ ] "):
			taskText := strings.TrimPrefix(line, "- [ ] ")
			incompleteTasks = append(incompleteTasks, Task{
				Text:     taskText,
				Complete: false,
				Section:  currentSection,
				Subtree:  taskSubtree,
			})
			if currentTask == "" {
				currentTask = taskText
				firstIncompleteIdx = i
			}
		}
	}

	// Find the current task section by scanning upward from the first incomplete task
	if firstIncompleteIdx >= 0 {
		for i := firstIncompleteIdx - 1; i >= taskListStart; i-- {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "### ") {
				currentTaskSection = strings.TrimPrefix(trimmed, "### ")
				break
			}
		}
	}

	return true, completeTasks, incompleteTasks, currentTask, currentTaskSection
}

func isTopLevelTaskLine(line string) bool {
	if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
		return false
	}

	return strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [ ] ")
}

func findTaskSubtreeEnd(lines []string, start, taskListEnd int) int {
	for i := start + 1; i < taskListEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "### ") {
			return i
		}
		if isTopLevelTaskLine(lines[i]) {
			return i
		}
	}

	return taskListEnd
}

func extractTaskListSectionOrders(lines []string) map[string]int {
	orders := make(map[string]int)
	inTaskList := false
	order := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "## Task List" {
			inTaskList = true
			continue
		}

		if inTaskList && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			break
		}

		if !inTaskList || !strings.HasPrefix(trimmed, "### ") {
			continue
		}

		sectionName := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
		if _, exists := orders[sectionName]; exists {
			continue
		}

		order++
		orders[sectionName] = order
	}

	return orders
}

// inferStatus determines the spec status based on frontmatter and task state.
// Explicit frontmatter status always overrides inference.
func inferStatus(fmStatus string, hasTaskList bool, completeCount, incompleteCount int) string {
	// Explicit frontmatter status always wins
	if fmStatus != "" {
		return fmStatus
	}

	// No task list → draft
	if !hasTaskList {
		return "draft"
	}

	// Has task list but no tasks at all → draft
	if completeCount == 0 && incompleteCount == 0 {
		return "draft"
	}

	// No complete tasks → draft
	if completeCount == 0 {
		return "draft"
	}

	// All complete → completed
	if incompleteCount == 0 {
		return "completed"
	}

	// Mixed → in-progress
	return "in-progress"
}

// FindAll finds all spec files in the given specs directory.
// It supports both the current flat .md layout and the future nested
// SPEC.md layout inside spec directories.
func FindAll(specsDir string) ([]string, error) {
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("specs directory not found: %s", specsDir)
	}

	var paths []string
	if err := collectSpecPaths(specsDir, specsDir, &paths); err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}

// ResolvePath resolves a spec reference or path to a file path.
// Accepts:
//   - Full path: specs/000-mvp.md or specs/my-feature.md
//   - Numeric references with or without leading zeros: 0, 00, 000
//   - Hierarchical references: 1.4.3
//
// Lookups are performed against the parsed full reference derived from the
// directory tree.
func ResolvePath(specsDir, arg string) (string, error) {
	// If it's already a path that exists, use it
	if _, err := os.Stat(arg); err == nil {
		return arg, nil
	}

	fullRef, err := normalizeSpecRef(arg)
	if err != nil {
		return "", err
	}

	paths, err := FindAll(specsDir)
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		info, err := Parse(p)
		if err != nil {
			continue
		}
		if info.FullRef == fullRef {
			return p, nil
		}
	}

	return "", fmt.Errorf("spec not found: %s", arg)
}

// collectSpecPaths walks the specs tree and records every discoverable spec file.
// It keeps the current flat root .md files and the future nested SPEC.md files.
func collectSpecPaths(rootDir, dir string, paths *[]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := collectSpecPaths(rootDir, path, paths); err != nil {
				return err
			}
			continue
		}

		if !shouldIncludeSpecPath(rootDir, path) {
			continue
		}

		*paths = append(*paths, path)
	}

	return nil
}

// shouldIncludeSpecPath returns true for files that count as specs.
// Today that means flat markdown files at the specs root and, in the new
// hierarchy, SPEC.md files inside spec directories. README files are skipped.
func shouldIncludeSpecPath(rootDir, path string) bool {
	base := filepath.Base(path)
	if base == "README.md" {
		return false
	}

	if base == "SPEC.md" {
		return true
	}

	parentDir := filepath.Dir(path)
	return parentDir == rootDir && strings.HasSuffix(base, ".md")
}

// normalizeSpecRef canonicalizes a user-provided reference so lookup can match
// against parsed FullRef values. It trims whitespace and removes leading zeros
// from each segment, so values like 001.002 compare as 1.2.
func normalizeSpecRef(arg string) (string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return "", fmt.Errorf("invalid spec reference: %s (expected a numeric reference like 0, 00, 000, or 1.4.3)", arg)
	}

	parts := strings.Split(trimmed, ".")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid spec reference: %s (expected a numeric reference like 0, 00, 000, or 1.4.3)", arg)
		}

		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return "", fmt.Errorf("invalid spec reference: %s (expected a numeric reference like 0, 00, 000, or 1.4.3)", arg)
		}
		normalized = append(normalized, strconv.Itoa(number))
	}

	return strings.Join(normalized, "."), nil
}
