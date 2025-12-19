package prompt

import (
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
		want    bool
		wantErr bool
	}{
		{
			name:    "yes response",
			input:   "yes\n",
			message: "Continue?",
			want:    true,
			wantErr: false,
		},
		{
			name:    "y response",
			input:   "y\n",
			message: "Continue?",
			want:    true,
			wantErr: false,
		},
		{
			name:    "no response",
			input:   "no\n",
			message: "Continue?",
			want:    false,
			wantErr: false,
		},
		{
			name:    "n response",
			input:   "n\n",
			message: "Continue?",
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid response with retry",
			input:   "invalid\nyes\n",
			message: "Continue?",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := confirm(tt.message, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		prompt  string
		want    string
		wantErr bool
	}{
		{
			name:    "simple input",
			input:   "hello\n",
			prompt:  "Name: ",
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "input with spaces",
			input:   "hello world\n",
			prompt:  "Message: ",
			want:    "hello world",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "\n",
			prompt:  "Optional: ",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := promptString(tt.prompt, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PromptString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfirmWithDefault(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
		defVal  bool
		want    bool
		wantErr bool
	}{
		{
			name:    "yes response overrides default",
			input:   "yes\n",
			message: "Continue?",
			defVal:  false,
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty input uses default true",
			input:   "\n",
			message: "Continue?",
			defVal:  true,
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty input uses default false",
			input:   "\n",
			message: "Continue?",
			defVal:  false,
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := confirmWithDefault(tt.message, tt.defVal, reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfirmWithDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfirmWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
