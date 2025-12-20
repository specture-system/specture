package setup

import (
	"embed"
)

//go:embed templates/*.md
var templateFiles embed.FS

// GetAgentPromptTemplate returns the agent prompt template.
func GetAgentPromptTemplate() (string, error) {
	content, err := templateFiles.ReadFile("templates/agent-prompt.md")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetSpecsReadmeTemplate returns the specs README template.
func GetSpecsReadmeTemplate() (string, error) {
	content, err := templateFiles.ReadFile("templates/specs-readme.md")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
