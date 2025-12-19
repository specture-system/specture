package prompt

import (
	"io"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name    string
		input   io.Reader
		message string
		want    bool
		wantErr string // empty string means no error expected
	}{
		{
			name:    "yes response",
			input:   strings.NewReader("yes\n"),
			message: "Continue?",
			want:    true,
			wantErr: "",
		},
		{
			name:    "y response",
			input:   strings.NewReader("y\n"),
			message: "Continue?",
			want:    true,
			wantErr: "",
		},
		{
			name:    "no response",
			input:   strings.NewReader("no\n"),
			message: "Continue?",
			want:    false,
			wantErr: "",
		},
		{
			name:    "n response",
			input:   strings.NewReader("n\n"),
			message: "Continue?",
			want:    false,
			wantErr: "",
		},
		{
			name:    "invalid response with retry",
			input:   strings.NewReader("invalid\nyes\n"),
			message: "Continue?",
			want:    true,
			wantErr: "",
		},
		{
			name:    "EOF input",
			input:   strings.NewReader(""),
			message: "Continue?",
			want:    false,
			wantErr: "failed to read input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := confirm(tt.message, tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Confirm() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Confirm() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Confirm() unexpected error = %v", err)
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
		input   io.Reader
		prompt  string
		want    string
		wantErr string // empty string means no error expected
	}{
		{
			name:    "simple input",
			input:   strings.NewReader("hello\n"),
			prompt:  "Name: ",
			want:    "hello",
			wantErr: "",
		},
		{
			name:    "input with spaces",
			input:   strings.NewReader("hello world\n"),
			prompt:  "Message: ",
			want:    "hello world",
			wantErr: "",
		},
		{
			name:    "empty input",
			input:   strings.NewReader("\n"),
			prompt:  "Optional: ",
			want:    "",
			wantErr: "",
		},
		{
			name:    "EOF input",
			input:   strings.NewReader(""),
			prompt:  "Name: ",
			want:    "",
			wantErr: "failed to read input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := promptString(tt.prompt, tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("PromptString() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("PromptString() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("PromptString() unexpected error = %v", err)
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
		input   io.Reader
		message string
		defVal  bool
		want    bool
		wantErr string // empty string means no error expected
	}{
		{
			name:    "yes response overrides default",
			input:   strings.NewReader("yes\n"),
			message: "Continue?",
			defVal:  false,
			want:    true,
			wantErr: "",
		},
		{
			name:    "empty input uses default true",
			input:   strings.NewReader("\n"),
			message: "Continue?",
			defVal:  true,
			want:    true,
			wantErr: "",
		},
		{
			name:    "empty input uses default false",
			input:   strings.NewReader("\n"),
			message: "Continue?",
			defVal:  false,
			want:    false,
			wantErr: "",
		},
		{
			name:    "EOF input",
			input:   strings.NewReader(""),
			message: "Continue?",
			defVal:  true,
			want:    false,
			wantErr: "failed to read input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := confirmWithDefault(tt.message, tt.defVal, tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("ConfirmWithDefault() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("ConfirmWithDefault() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ConfirmWithDefault() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ConfirmWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
