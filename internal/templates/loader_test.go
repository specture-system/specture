package templates

import "testing"

func TestImplementPromptTemplatesUseTaskAndSectionPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "task worker", filename: "task-worker-prompt.md"},
		{name: "task review", filename: "task-review-prompt.md"},
		{name: "section worker", filename: "section-worker-prompt.md"},
		{name: "section review", filename: "section-review-prompt.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readTemplate(tt.filename, tt.name+" prompt template")
			if err != nil {
				t.Fatalf("readTemplate(%q) returned error: %v", tt.filename, err)
			}
		})
	}
}
