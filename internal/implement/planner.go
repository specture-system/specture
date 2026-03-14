package implement

import (
	"fmt"

	specpkg "github.com/specture-system/specture/internal/spec"
)

const (
	StatusApproved   = "approved"
	StatusInProgress = "in-progress"
)

// RemainingSection groups all remaining tasks for a task-list section.
type RemainingSection struct {
	Name  string
	Tasks []specpkg.Task
}

// Plan captures remaining tasks grouped by section in encounter order.
type Plan struct {
	Sections  []RemainingSection
	TaskCount int
}

// ValidateSpecStatus ensures a spec is in a status implement can execute.
func ValidateSpecStatus(status string) error {
	if status == StatusApproved || status == StatusInProgress {
		return nil
	}

	return fmt.Errorf("spec status %q is not implementable (must be 'approved' or 'in-progress')", status)
}

// PlanRemaining groups incomplete tasks by section while preserving order.
func PlanRemaining(info *specpkg.SpecInfo) Plan {
	sections := make([]RemainingSection, 0)
	sectionIndexByName := make(map[string]int)

	for _, task := range info.IncompleteTasks {
		idx, exists := sectionIndexByName[task.Section]
		if !exists {
			sections = append(sections, RemainingSection{Name: task.Section})
			idx = len(sections) - 1
			sectionIndexByName[task.Section] = idx
		}

		sections[idx].Tasks = append(sections[idx].Tasks, task)
	}

	return Plan{
		Sections:  sections,
		TaskCount: len(info.IncompleteTasks),
	}
}
