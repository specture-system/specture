package templates

import (
	"embed"
	"fmt"
)

//go:embed files/spec-template.md
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
