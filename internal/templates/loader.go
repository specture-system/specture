package templates

import (
	"embed"
	"fmt"
)

//go:embed files/*.md
var templateFiles embed.FS

// GetSpecTemplate returns the spec file template.
func GetSpecTemplate() (string, error) {
	content, err := templateFiles.ReadFile("files/spec-template.md")
	if err != nil {
		return "", fmt.Errorf("failed to read spec template: %w", err)
	}
	return string(content), nil
}

// GetAgentPromptTemplate returns the agent prompt template.
func GetAgentPromptTemplate() (string, error) {
	content, err := templateFiles.ReadFile("files/agent-prompt.md")
	if err != nil {
		return "", fmt.Errorf("failed to read agent prompt template: %w", err)
	}
	return string(content), nil
}

// GetSpecsReadmeTemplate returns the specs README template.
func GetSpecsReadmeTemplate() (string, error) {
	content, err := templateFiles.ReadFile("files/specs-readme.md")
	if err != nil {
		return "", fmt.Errorf("failed to read specs readme template: %w", err)
	}
	return string(content), nil
}
