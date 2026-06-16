package templates

import (
	"embed"
	"fmt"
)

//go:embed files/*.md
var templateFiles embed.FS

// readTemplate reads and returns a template file from the embedded filesystem.
func readTemplate(filename, description string) (string, error) {
	content, err := templateFiles.ReadFile("files/" + filename)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", description, err)
	}
	return string(content), nil
}

// GetSpecTemplate returns the spec file template.
func GetSpecTemplate() (string, error) {
	return readTemplate("spec-template.md", "spec template")
}

// GetImplementWorkerPromptTemplate returns the implement worker prompt template.
func GetImplementWorkerPromptTemplate() (string, error) {
	return readTemplate("task-worker-prompt.md", "implement worker prompt template")
}

// GetImplementReviewPromptTemplate returns the implement review prompt template.
func GetImplementReviewPromptTemplate() (string, error) {
	return readTemplate("task-review-prompt.md", "implement review prompt template")
}

// GetImplementSectionReviewPromptTemplate returns the section-level implement review prompt template.
func GetImplementSectionReviewPromptTemplate() (string, error) {
	return readTemplate("section-review-prompt.md", "implement section review prompt template")
}

// GetImplementSectionWorkerPromptTemplate returns the section-level implement worker prompt template.
func GetImplementSectionWorkerPromptTemplate() (string, error) {
	return readTemplate("section-worker-prompt.md", "implement section worker prompt template")
}

// GetImplementCleanupReviewPromptTemplate returns the final cleanup review prompt template.
func GetImplementCleanupReviewPromptTemplate() (string, error) {
	return readTemplate("cleanup-review-prompt.md", "implement cleanup review prompt template")
}

// GetImplementCleanupWorkerPromptTemplate returns the final cleanup worker prompt template.
func GetImplementCleanupWorkerPromptTemplate() (string, error) {
	return readTemplate("cleanup-worker-prompt.md", "implement cleanup worker prompt template")
}
