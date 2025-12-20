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

// GetAgentPromptTemplate returns the agent prompt template.
func GetAgentPromptTemplate() (string, error) {
	return readTemplate("agent-prompt.md", "agent prompt template")
}

// GetSpecsReadmeTemplate returns the specs README template.
func GetSpecsReadmeTemplate() (string, error) {
	return readTemplate("specs-readme.md", "specs readme template")
}
