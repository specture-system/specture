package template

import (
	"bytes"
	"fmt"
	tpl "text/template"
)

// RenderTemplate renders a Go template with the given data and returns the result as a string.
func RenderTemplate(templateStr string, data interface{}) (string, error) {
	t, err := tpl.New("").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
