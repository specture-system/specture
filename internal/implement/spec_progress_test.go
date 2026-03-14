package implement

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyTaskProgress_MarksNestedCheckboxSubtreeComplete(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.md")
	specBody := `---
number: 7
status: in-progress
---

# Test Spec

## Task List

### Task Structure and Validation

- [ ] Update the implement command to support nested checkboxes
  - [ ] Treat each top-level checkbox as one implementation, review, and commit unit
    - [ ] Deeply nested checkbox is also part of the same unit
  - [ ] Include nested checkboxes and nested bullets at every depth
    - Nested bullet detail
- [ ] Update the validate command to require every top-level task-list checkbox to appear under a ### section
`

	if err := os.WriteFile(specPath, []byte(specBody), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	if err := applyTaskProgress(specPath, "Task Structure and Validation", "Update the implement command to support nested checkboxes", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec: %v", err)
	}

	updatedText := string(updated)
	if !strings.Contains(updatedText, "- [x] Update the implement command to support nested checkboxes") {
		t.Fatalf("expected parent task to be complete:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "  - [x] Treat each top-level checkbox as one implementation, review, and commit unit") {
		t.Fatalf("expected nested task to be complete:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "    - [x] Deeply nested checkbox is also part of the same unit") {
		t.Fatalf("expected deeply nested task to be complete:\n%s", updatedText)
	}
	if !strings.Contains(updatedText, "- [ ] Update the validate command to require every top-level task-list checkbox to appear under a ### section") {
		t.Fatalf("expected next top-level task to remain incomplete:\n%s", updatedText)
	}
}
