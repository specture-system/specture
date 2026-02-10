package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to build a minimal spec with frontmatter, title, and a task list section.
func buildSpec(frontmatter, title, taskListBody string) []byte {
	var s string
	if frontmatter != "" {
		s += "---\n" + frontmatter + "\n---\n\n"
	}
	if title != "" {
		s += "# " + title + "\n\n"
	}
	if taskListBody != "" {
		s += "## Task List\n\n" + taskListBody + "\n"
	}
	return []byte(s)
}

// ---------- Task parsing tests ----------

func TestParseTasks_OnlyComplete(t *testing.T) {
	content := buildSpec("", "Test", "- [x] Task A\n- [x] Task B\n- [x] Task C")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.CompleteTasks) != 3 {
		t.Errorf("expected 3 complete tasks, got %d", len(info.CompleteTasks))
	}
	if len(info.IncompleteTasks) != 0 {
		t.Errorf("expected 0 incomplete tasks, got %d", len(info.IncompleteTasks))
	}
}

func TestParseTasks_OnlyIncomplete(t *testing.T) {
	content := buildSpec("", "Test", "- [ ] Task A\n- [ ] Task B")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.CompleteTasks) != 0 {
		t.Errorf("expected 0 complete tasks, got %d", len(info.CompleteTasks))
	}
	if len(info.IncompleteTasks) != 2 {
		t.Errorf("expected 2 incomplete tasks, got %d", len(info.IncompleteTasks))
	}
}

func TestParseTasks_Mixed(t *testing.T) {
	content := buildSpec("", "Test", "- [x] Done\n- [ ] Not done\n- [x] Also done")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.CompleteTasks) != 2 {
		t.Errorf("expected 2 complete tasks, got %d", len(info.CompleteTasks))
	}
	if len(info.IncompleteTasks) != 1 {
		t.Errorf("expected 1 incomplete task, got %d", len(info.IncompleteTasks))
	}
}

func TestParseTasks_EmptyTaskList(t *testing.T) {
	content := buildSpec("", "Test", "Tasks will be added later.")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.CompleteTasks) != 0 {
		t.Errorf("expected 0 complete tasks, got %d", len(info.CompleteTasks))
	}
	if len(info.IncompleteTasks) != 0 {
		t.Errorf("expected 0 incomplete tasks, got %d", len(info.IncompleteTasks))
	}
}

func TestParseTasks_IndentedSubTasksSkipped(t *testing.T) {
	content := buildSpec("", "Test", "- [ ] Top-level task\n  - [ ] Sub-task 1\n  - [x] Sub-task 2\n- [x] Another top-level")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(info.IncompleteTasks) != 1 {
		t.Errorf("expected 1 incomplete top-level task, got %d", len(info.IncompleteTasks))
	}
	if len(info.CompleteTasks) != 1 {
		t.Errorf("expected 1 complete top-level task, got %d", len(info.CompleteTasks))
	}
	if len(info.IncompleteTasks) > 0 && info.IncompleteTasks[0].Text != "Top-level task" {
		t.Errorf("expected incomplete task text 'Top-level task', got %q", info.IncompleteTasks[0].Text)
	}
	if len(info.CompleteTasks) > 0 && info.CompleteTasks[0].Text != "Another top-level" {
		t.Errorf("expected complete task text 'Another top-level', got %q", info.CompleteTasks[0].Text)
	}
}

func TestParseTasks_SectionField(t *testing.T) {
	content := buildSpec("", "Test", "### Phase 1\n\n- [x] Task A\n- [ ] Task B\n\n### Phase 2\n\n- [ ] Task C")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check section assignment for complete tasks
	if len(info.CompleteTasks) != 1 {
		t.Fatalf("expected 1 complete task, got %d", len(info.CompleteTasks))
	}
	if info.CompleteTasks[0].Section != "Phase 1" {
		t.Errorf("expected section 'Phase 1', got %q", info.CompleteTasks[0].Section)
	}

	// Check section assignment for incomplete tasks
	if len(info.IncompleteTasks) != 2 {
		t.Fatalf("expected 2 incomplete tasks, got %d", len(info.IncompleteTasks))
	}
	if info.IncompleteTasks[0].Section != "Phase 1" {
		t.Errorf("expected section 'Phase 1' for first incomplete, got %q", info.IncompleteTasks[0].Section)
	}
	if info.IncompleteTasks[1].Section != "Phase 2" {
		t.Errorf("expected section 'Phase 2' for second incomplete, got %q", info.IncompleteTasks[1].Section)
	}
}

// ---------- Current task / section tests ----------

func TestCurrentTask_HappyPath(t *testing.T) {
	content := buildSpec("", "Test", "### Setup\n\n- [x] Done task\n- [ ] Current task\n- [ ] Future task")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.CurrentTask != "Current task" {
		t.Errorf("expected CurrentTask 'Current task', got %q", info.CurrentTask)
	}
	if info.CurrentTaskSection != "Setup" {
		t.Errorf("expected CurrentTaskSection 'Setup', got %q", info.CurrentTaskSection)
	}
}

func TestCurrentTask_NoIncompleteTasks(t *testing.T) {
	content := buildSpec("", "Test", "- [x] Done 1\n- [x] Done 2")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.CurrentTask != "" {
		t.Errorf("expected empty CurrentTask, got %q", info.CurrentTask)
	}
}

func TestCurrentTask_NoSectionHeader(t *testing.T) {
	content := buildSpec("", "Test", "- [x] Done\n- [ ] No section above me")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.CurrentTask != "No section above me" {
		t.Errorf("expected CurrentTask 'No section above me', got %q", info.CurrentTask)
	}
	if info.CurrentTaskSection != "" {
		t.Errorf("expected empty CurrentTaskSection, got %q", info.CurrentTaskSection)
	}
}

func TestCurrentTask_MultipleSections(t *testing.T) {
	content := buildSpec("", "Test", "### Phase 1\n\n- [x] Done 1\n- [x] Done 2\n\n### Phase 2\n\n- [ ] First in phase 2\n- [ ] Second in phase 2\n\n### Phase 3\n\n- [ ] Task in phase 3")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.CurrentTask != "First in phase 2" {
		t.Errorf("expected CurrentTask 'First in phase 2', got %q", info.CurrentTask)
	}
	if info.CurrentTaskSection != "Phase 2" {
		t.Errorf("expected CurrentTaskSection 'Phase 2', got %q", info.CurrentTaskSection)
	}
}

// ---------- Status inference tests ----------

func TestStatusInference(t *testing.T) {
	tests := []struct {
		name       string
		content    []byte
		wantStatus string
	}{
		{
			name:       "no task list, no frontmatter → draft",
			content:    []byte("# Test\n\nJust a description.\n"),
			wantStatus: "draft",
		},
		{
			name:       "no task list, frontmatter approved → approved",
			content:    buildSpec("status: approved", "Test", ""),
			wantStatus: "approved",
		},
		{
			name:       "task list with no tasks → draft",
			content:    buildSpec("", "Test", "No actual tasks here."),
			wantStatus: "draft",
		},
		{
			name:       "all incomplete → draft",
			content:    buildSpec("", "Test", "- [ ] Task 1\n- [ ] Task 2"),
			wantStatus: "draft",
		},
		{
			name:       "mixed → in-progress",
			content:    buildSpec("", "Test", "- [x] Done\n- [ ] Not done"),
			wantStatus: "in-progress",
		},
		{
			name:       "all complete → completed",
			content:    buildSpec("", "Test", "- [x] Done 1\n- [x] Done 2"),
			wantStatus: "completed",
		},
		{
			name:       "frontmatter draft overrides mixed tasks → draft",
			content:    buildSpec("status: draft", "Test", "- [x] Done\n- [ ] Not done"),
			wantStatus: "draft",
		},
		{
			name:       "frontmatter completed overrides incomplete tasks → completed",
			content:    buildSpec("status: completed", "Test", "- [ ] Not done\n- [ ] Also not done"),
			wantStatus: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseContent("001-test.md", tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Status != tt.wantStatus {
				t.Errorf("expected status %q, got %q", tt.wantStatus, info.Status)
			}
		})
	}
}

// ---------- FindAll tests ----------

func TestFindAll_MatchesOnlySpecFiles(t *testing.T) {
	dir := t.TempDir()

	// Create spec files and non-spec files
	files := map[string]bool{
		"001-first.md":  true,
		"002-second.md": true,
		"010-tenth.md":  true,
		"README.md":     false,
		"notes.txt":     false,
		"1-bad.md":      false, // not 3 digits
		"abc-bad.md":    false,
	}

	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("# Test\n"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	paths, err := FindAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count expected matches
	expectedCount := 0
	for _, isSpec := range files {
		if isSpec {
			expectedCount++
		}
	}

	if len(paths) != expectedCount {
		t.Errorf("expected %d spec files, got %d: %v", expectedCount, len(paths), paths)
	}

	// Verify all returned paths are actual spec files
	for _, p := range paths {
		base := filepath.Base(p)
		if !files[base] {
			t.Errorf("unexpected file in results: %s", base)
		}
	}
}

func TestFindAll_NonexistentDirectory(t *testing.T) {
	_, err := FindAll("/nonexistent/directory/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// ---------- ResolvePath tests ----------

func TestResolvePath(t *testing.T) {
	dir := t.TempDir()

	// Create a spec file
	specFile := "007-feature.md"
	specPath := filepath.Join(dir, specFile)
	if err := os.WriteFile(specPath, []byte("# Feature\n"), 0644); err != nil {
		t.Fatalf("failed to create spec file: %v", err)
	}

	tests := []struct {
		name      string
		arg       string
		wantPath  string
		wantError bool
	}{
		{
			name:     "single digit",
			arg:      "7",
			wantPath: specPath,
		},
		{
			name:     "double digit",
			arg:      "07",
			wantPath: specPath,
		},
		{
			name:     "triple digit",
			arg:      "007",
			wantPath: specPath,
		},
		{
			name:     "full path",
			arg:      specPath,
			wantPath: specPath,
		},
		{
			name:      "nonexistent spec number",
			arg:       "999",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolvePath(dir, tt.arg)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.wantPath {
				t.Errorf("expected path %q, got %q", tt.wantPath, result)
			}
		})
	}
}

// ---------- ParseAll tests ----------

func TestParseAll_SortedByNumber(t *testing.T) {
	dir := t.TempDir()

	// Create specs out of order, with number in frontmatter
	specs := map[string]string{
		"003-third.md":  "---\nnumber: 3\n---\n\n# Third\n",
		"001-first.md":  "---\nnumber: 1\n---\n\n# First\n",
		"002-second.md": "---\nnumber: 2\n---\n\n# Second\n",
	}
	for name, content := range specs {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	result, err := ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(result))
	}

	expectedOrder := []int{1, 2, 3}
	for i, s := range result {
		if s.Number != expectedOrder[i] {
			t.Errorf("position %d: expected number %d, got %d", i, expectedOrder[i], s.Number)
		}
	}
}

func TestParseAll_NonexistentDir(t *testing.T) {
	_, err := ParseAll("/nonexistent/directory/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// ---------- FindCurrent tests ----------

func TestFindCurrent_ReturnsFirstInProgress(t *testing.T) {
	dir := t.TempDir()

	// Spec 1: completed (all tasks done)
	spec1 := buildSpec("number: 1", "First", "- [x] Done")
	// Spec 2: in-progress (mixed tasks)
	spec2 := buildSpec("number: 2", "Second", "- [x] Done\n- [ ] Not done")
	// Spec 3: also in-progress
	spec3 := buildSpec("number: 3", "Third", "- [x] A\n- [ ] B")

	files := map[string][]byte{
		"001-first.md":  spec1,
		"002-second.md": spec2,
		"003-third.md":  spec3,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	specs, err := ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	current := FindCurrent(specs)
	if current == nil {
		t.Fatal("expected a current spec, got nil")
	}
	if current.Number != 2 {
		t.Errorf("expected current spec number 2, got %d", current.Number)
	}
}

func TestFindCurrent_NoInProgress(t *testing.T) {
	dir := t.TempDir()

	// Spec 1: completed
	spec1 := buildSpec("number: 1", "First", "- [x] Done")
	// Spec 2: draft (all incomplete)
	spec2 := buildSpec("number: 2", "Second", "- [ ] Not started")

	files := map[string][]byte{
		"001-first.md":  spec1,
		"002-second.md": spec2,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	specs, err := ParseAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	current := FindCurrent(specs)
	if current != nil {
		t.Errorf("expected nil, got spec number %d", current.Number)
	}
}

func TestFindCurrent_MultipleInProgress_ReturnsLowest(t *testing.T) {
	// Use ParseContent directly to build specs with known statuses
	specs := []*SpecInfo{
		{Number: 5, Status: "in-progress"},
		{Number: 3, Status: "in-progress"},
		{Number: 7, Status: "in-progress"},
		{Number: 1, Status: "completed"},
	}

	current := FindCurrent(specs)
	if current == nil {
		t.Fatal("expected a current spec, got nil")
	}
	if current.Number != 3 {
		t.Errorf("expected lowest in-progress spec number 3, got %d", current.Number)
	}
}

// ---------- Additional edge case tests ----------

func TestParseContent_ExtractsNumber(t *testing.T) {
	// Number is read exclusively from frontmatter, not filename
	tests := []struct {
		name       string
		path       string
		content    []byte
		wantNumber int
	}{
		{"frontmatter number 0", "000-mvp.md", buildSpec("number: 0\nstatus: draft", "Test", ""), 0},
		{"frontmatter number 1", "001-feature.md", buildSpec("number: 1\nstatus: draft", "Test", ""), 1},
		{"frontmatter number 42", "042-answer.md", buildSpec("number: 42\nstatus: draft", "Test", ""), 42},
		{"no frontmatter number", "100-big.md", []byte("# Test\n"), -1},
		{"slug-only filename with number", "big-feature.md", buildSpec("number: 100\nstatus: draft", "Test", ""), 100},
		{"no number anywhere", "bad-name.md", []byte("# Test\n"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseContent(tt.path, tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Number != tt.wantNumber {
				t.Errorf("expected number %d, got %d", tt.wantNumber, info.Number)
			}
		})
	}
}

// ---------- Number from frontmatter tests ----------

func TestParseContent_NumberFromFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		content    []byte
		path       string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "number present in frontmatter",
			content:    buildSpec("number: 5\nstatus: draft", "Test", ""),
			path:       "test.md",
			wantNumber: 5,
		},
		{
			name:       "number zero in frontmatter",
			content:    buildSpec("number: 0\nstatus: draft", "Test", ""),
			path:       "test.md",
			wantNumber: 0,
		},
		{
			name:       "large number in frontmatter",
			content:    buildSpec("number: 999\nstatus: draft", "Test", ""),
			path:       "test.md",
			wantNumber: 999,
		},
		{
			name:       "frontmatter number overrides filename number",
			content:    buildSpec("number: 42\nstatus: draft", "Test", ""),
			path:       "003-test.md",
			wantNumber: 42,
		},
		{
			name:       "missing number defaults to -1 even with numeric prefix filename",
			content:    buildSpec("status: draft", "Test", ""),
			path:       "007-test.md",
			wantNumber: -1,
		},
		{
			name:       "missing number without numeric prefix defaults to -1",
			content:    buildSpec("status: draft", "Test", ""),
			path:       "test.md",
			wantNumber: -1,
		},
		{
			name:       "negative number is invalid",
			content:    buildSpec("number: -1\nstatus: draft", "Test", ""),
			path:       "test.md",
			wantErr:    true,
		},
		{
			name:       "no frontmatter defaults to -1",
			content:    []byte("# Test\n"),
			path:       "005-test.md",
			wantNumber: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseContent(tt.path, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Number != tt.wantNumber {
				t.Errorf("expected number %d, got %d", tt.wantNumber, info.Number)
			}
		})
	}
}

func TestParseContent_ExtractsName(t *testing.T) {
	content := []byte("# My Great Feature\n\nDescription.\n")
	info, err := ParseContent("001-test.md", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "My Great Feature" {
		t.Errorf("expected name 'My Great Feature', got %q", info.Name)
	}
}
