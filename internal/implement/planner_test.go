package implement

import (
	"strings"
	"testing"

	specpkg "github.com/specture-system/specture/internal/spec"
)

func TestValidateSpecStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "approved is valid", status: StatusApproved, wantErr: false},
		{name: "in-progress is valid", status: StatusInProgress, wantErr: false},
		{name: "draft is invalid", status: "draft", wantErr: true},
		{name: "completed is invalid", status: "completed", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpecStatus(tt.status)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSpecStatus_ErrorMessage(t *testing.T) {
	err := ValidateSpecStatus("draft")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "must be 'approved' or 'in-progress'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlanRemaining_GroupsIncompleteTasksBySection(t *testing.T) {
	info := &specpkg.SpecInfo{
		IncompleteTasks: []specpkg.Task{
			{Text: "task 1", Section: "CLI and Planning"},
			{Text: "task 2", Section: "CLI and Planning"},
			{Text: "task 3", Section: "Branch and Task Execution"},
			{Text: "task 4", Section: "Spec Updates and Section Delivery"},
		},
	}

	plan := PlanRemaining(info)

	if plan.TaskCount != 4 {
		t.Fatalf("expected 4 tasks, got %d", plan.TaskCount)
	}

	if len(plan.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(plan.Sections))
	}

	if plan.Sections[0].Name != "CLI and Planning" {
		t.Fatalf("expected first section CLI and Planning, got %q", plan.Sections[0].Name)
	}
	if len(plan.Sections[0].Tasks) != 2 {
		t.Fatalf("expected 2 tasks in first section, got %d", len(plan.Sections[0].Tasks))
	}

	if plan.Sections[1].Name != "Branch and Task Execution" {
		t.Fatalf("expected second section Branch and Task Execution, got %q", plan.Sections[1].Name)
	}
	if len(plan.Sections[1].Tasks) != 1 {
		t.Fatalf("expected 1 task in second section, got %d", len(plan.Sections[1].Tasks))
	}

	if plan.Sections[2].Name != "Spec Updates and Section Delivery" {
		t.Fatalf("expected third section Spec Updates and Section Delivery, got %q", plan.Sections[2].Name)
	}
	if len(plan.Sections[2].Tasks) != 1 {
		t.Fatalf("expected 1 task in third section, got %d", len(plan.Sections[2].Tasks))
	}
}

func TestPlanRemaining_HandlesUnsectionedTasks(t *testing.T) {
	info := &specpkg.SpecInfo{
		IncompleteTasks: []specpkg.Task{
			{Text: "task 1", Section: ""},
			{Text: "task 2", Section: ""},
		},
	}

	plan := PlanRemaining(info)

	if len(plan.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(plan.Sections))
	}

	if plan.Sections[0].Name != "" {
		t.Fatalf("expected empty section name, got %q", plan.Sections[0].Name)
	}
}

func TestPlanRemaining_NoRemainingTasks(t *testing.T) {
	info := &specpkg.SpecInfo{IncompleteTasks: nil}
	plan := PlanRemaining(info)

	if plan.TaskCount != 0 {
		t.Fatalf("expected 0 tasks, got %d", plan.TaskCount)
	}

	if len(plan.Sections) != 0 {
		t.Fatalf("expected 0 sections, got %d", len(plan.Sections))
	}
}
