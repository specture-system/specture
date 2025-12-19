package template

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name    string
		tpl     string
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple substitution",
			tpl:     "Hello {{.Name}}",
			data:    map[string]string{"Name": "World"},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "multiple variables",
			tpl:     "{{.First}} {{.Last}} ({{.Email}})",
			data:    map[string]string{"First": "Alice", "Last": "Smith", "Email": "alice@example.com"},
			want:    "Alice Smith (alice@example.com)",
			wantErr: false,
		},
		{
			name:    "nested struct access",
			tpl:     "{{.User.Name}} - {{.User.Email}}",
			data:    map[string]interface{}{"User": map[string]string{"Name": "Bob", "Email": "bob@example.com"}},
			want:    "Bob - bob@example.com",
			wantErr: false,
		},
		{
			name:    "conditional true",
			tpl:     "{{if .IsGitLab}}Merge Request{{else}}Pull Request{{end}}",
			data:    map[string]bool{"IsGitLab": true},
			want:    "Merge Request",
			wantErr: false,
		},
		{
			name:    "conditional false",
			tpl:     "{{if .IsGitLab}}Merge Request{{else}}Pull Request{{end}}",
			data:    map[string]bool{"IsGitLab": false},
			want:    "Pull Request",
			wantErr: false,
		},
		{
			name:    "forge terminology conditional",
			tpl:     "When you're ready, open a {{if .IsGitLab}}merge request{{else}}pull request{{end}}.",
			data:    map[string]bool{"IsGitLab": false},
			want:    "When you're ready, open a pull request.",
			wantErr: false,
		},
		{
			name:    "markdown with variables",
			tpl:     "# {{.Title}}\n\n{{.Description}}\n\nAuthor: {{.Author}}",
			data:    map[string]string{"Title": "My Feature", "Description": "Does something cool", "Author": "Alice"},
			want:    "# My Feature\n\nDoes something cool\n\nAuthor: Alice",
			wantErr: false,
		},
		{
			name:    "whitespace preservation",
			tpl:     "Line 1\n  Indented: {{.Value}}\nLine 3",
			data:    map[string]string{"Value": "test"},
			want:    "Line 1\n  Indented: test\nLine 3",
			wantErr: false,
		},
		{
			name:    "missing variable renders empty",
			tpl:     "Hello {{.Missing}}",
			data:    map[string]string{},
			want:    "Hello <no value>",
			wantErr: false,
		},
		{
			name:    "invalid template syntax",
			tpl:     "{{if .Unclosed",
			data:    nil,
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty template",
			tpl:     "",
			data:    nil,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate(tt.tpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
