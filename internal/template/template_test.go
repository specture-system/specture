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
			name:    "simple template",
			tpl:     "Hello {{.Name}}",
			data:    map[string]string{"Name": "World"},
			want:    "Hello World",
			wantErr: false,
		},
		{
			name:    "conditional template",
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
			name:    "multiline template",
			tpl:     "Line1\nLine2: {{.Value}}\nLine3",
			data:    map[string]string{"Value": "test"},
			want:    "Line1\nLine2: test\nLine3",
			wantErr: false,
		},
		{
			name:    "missing variable",
			tpl:     "Hello {{.Missing}}",
			data:    map[string]string{},
			want:    "Hello <no value>",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			tpl:     "{{if .Unclosed",
			data:    nil,
			want:    "",
			wantErr: true,
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

func TestRenderTemplateString(t *testing.T) {
	tests := []struct {
		name    string
		tpl     string
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "nested struct",
			tpl:     "{{.User.Name}} ({{.User.Email}})",
			data:    map[string]interface{}{"User": map[string]string{"Name": "Alice", "Email": "alice@example.com"}},
			want:    "Alice (alice@example.com)",
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
