package implement

import (
	"fmt"
	"os"
	"strings"
)

func applyTaskProgress(specPath, sectionName, taskText, status string) error {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	if status != "" {
		if err := updateFrontmatterStatus(lines, status); err != nil {
			return err
		}
	}

	if err := markTaskComplete(lines, sectionName, taskText); err != nil {
		return err
	}

	updated := strings.Join(lines, "\n")
	if err := os.WriteFile(specPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}

func updateFrontmatterStatus(lines []string, status string) error {
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Errorf("spec file is missing YAML frontmatter")
	}

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return fmt.Errorf("spec file frontmatter missing status field")
		}

		if strings.HasPrefix(strings.TrimSpace(lines[i]), "status:") {
			indent := lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]
			lines[i] = indent + "status: " + status
			return nil
		}
	}

	return fmt.Errorf("spec file frontmatter is not terminated")
}

func markTaskComplete(lines []string, sectionName, taskText string) error {
	inTaskList := false
	currentSection := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "## Task List" {
			inTaskList = true
			currentSection = ""
			continue
		}

		if !inTaskList {
			continue
		}

		if strings.HasPrefix(trimmed, "## ") && trimmed != "## Task List" {
			break
		}

		if strings.HasPrefix(trimmed, "### ") {
			currentSection = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			continue
		}

		if currentSection != sectionName {
			continue
		}

		prefix := "- [ ] "
		if strings.HasPrefix(trimmed, prefix) && strings.TrimPrefix(trimmed, prefix) == taskText {
			lines[i] = strings.Replace(line, "[ ]", "[x]", 1)
			return nil
		}
	}

	return fmt.Errorf("failed to find incomplete task %q in section %q", taskText, sectionName)
}
