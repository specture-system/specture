package template

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name    string
		tpl     string
		data    any
		want    string
		wantErr string // empty string means no error expected
	}{
		{
			name: "simple substitution",
			tpl:  "Hello {{.Name}}",
			data: map[string]string{"Name": "World"},
			want: "Hello World",
		},
		{
			name: "multiple variables",
			tpl:  "{{.First}} {{.Last}} ({{.Email}})",
			data: map[string]string{"First": "Alice", "Last": "Smith", "Email": "alice@example.com"},
			want: "Alice Smith (alice@example.com)",
		},
		{
			name: "nested struct access",
			tpl:  "{{.User.Name}} - {{.User.Email}}",
			data: map[string]any{"User": map[string]string{"Name": "Bob", "Email": "bob@example.com"}},
			want: "Bob - bob@example.com",
		},
		{
			name: "conditional true",
			tpl:  "{{if .IsGitLab}}Merge Request{{else}}Pull Request{{end}}",
			data: map[string]bool{"IsGitLab": true},
			want: "Merge Request",
		},
		{
			name: "conditional false",
			tpl:  "{{if .IsGitLab}}Merge Request{{else}}Pull Request{{end}}",
			data: map[string]bool{"IsGitLab": false},
			want: "Pull Request",
		},
		{
			name: "forge terminology conditional",
			tpl:  "When you're ready, open a {{if .IsGitLab}}merge request{{else}}pull request{{end}}.",
			data: map[string]bool{"IsGitLab": false},
			want: "When you're ready, open a pull request.",
		},
		{
			name: "markdown with variables",
			tpl:  "# {{.Title}}\n\n{{.Description}}\n\nAuthor: {{.Author}}",
			data: map[string]string{"Title": "My Feature", "Description": "Does something cool", "Author": "Alice"},
			want: "# My Feature\n\nDoes something cool\n\nAuthor: Alice",
		},
		{
			name: "whitespace preservation",
			tpl:  "Line 1\n  Indented: {{.Value}}\nLine 3",
			data: map[string]string{"Value": "test"},
			want: "Line 1\n  Indented: test\nLine 3",
		},
		{
			name: "missing variable renders empty",
			tpl:  "Hello {{.Missing}}",
			data: map[string]string{},
			want: "Hello <no value>",
		},
		{
			name:    "invalid template syntax",
			tpl:     "{{if .Unclosed",
			data:    nil,
			wantErr: "failed to parse template",
		},
		{
			name:    "failed to execute template",
			tpl:     "{{.Name.Invalid}}",
			data:    map[string]string{"Name": "test"},
			wantErr: "failed to execute template",
		},
		{
			name: "empty template",
			tpl:  "",
			data: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.tpl, tt.data)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("RenderTemplate() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("RenderTemplate() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("RenderTemplate() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
