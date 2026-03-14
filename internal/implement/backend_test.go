package implement

import (
	"errors"
	"strings"
	"testing"
)

func TestSelectBackend_AutoDetectPriority(t *testing.T) {
	lookPath := func(file string) (string, error) {
		switch file {
		case BackendOpencode:
			return "/usr/bin/opencode", nil
		case BackendCodex:
			return "/usr/bin/codex", nil
		default:
			return "", errors.New("missing")
		}
	}

	backend, err := SelectBackend("", lookPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend != BackendOpencode {
		t.Fatalf("expected %q, got %q", BackendOpencode, backend)
	}
}

func TestSelectBackend_AutoDetectFallbackToCodex(t *testing.T) {
	lookPath := func(file string) (string, error) {
		switch file {
		case BackendCodex:
			return "/usr/bin/codex", nil
		default:
			return "", errors.New("missing")
		}
	}

	backend, err := SelectBackend("", lookPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend != BackendCodex {
		t.Fatalf("expected %q, got %q", BackendCodex, backend)
	}
}

func TestSelectBackend_AgentOverride(t *testing.T) {
	lookPath := func(file string) (string, error) {
		if file == BackendCodex {
			return "/usr/bin/codex", nil
		}

		return "", errors.New("missing")
	}

	backend, err := SelectBackend(BackendCodex, lookPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend != BackendCodex {
		t.Fatalf("expected %q, got %q", BackendCodex, backend)
	}
}

func TestSelectBackend_InvalidOverride(t *testing.T) {
	_, err := SelectBackend("claude", func(file string) (string, error) {
		return "", errors.New("missing")
	})
	if err == nil {
		t.Fatal("expected error for invalid backend override")
	}

	if !strings.Contains(err.Error(), "invalid agent backend") {
		t.Fatalf("expected invalid backend message, got: %v", err)
	}
}

func TestSelectBackend_OverrideNotInstalled(t *testing.T) {
	_, err := SelectBackend(BackendCodex, func(file string) (string, error) {
		return "", errors.New("missing")
	})
	if err == nil {
		t.Fatal("expected error for unavailable override backend")
	}

	if !strings.Contains(err.Error(), "not available in PATH") {
		t.Fatalf("expected PATH availability message, got: %v", err)
	}
}

func TestSelectBackend_NoBackendFound(t *testing.T) {
	_, err := SelectBackend("", func(file string) (string, error) {
		return "", errors.New("missing")
	})
	if err == nil {
		t.Fatal("expected error when no backend exists")
	}

	if !strings.Contains(err.Error(), "no supported agent backend") {
		t.Fatalf("expected no backend message, got: %v", err)
	}
}

func TestBackendInvocationArgs_UsesBackendSpecificNonInteractiveSubcommands(t *testing.T) {
	tests := []struct {
		name      string
		backend   string
		wantFirst string
	}{
		{name: "opencode uses run", backend: BackendOpencode, wantFirst: "run"},
		{name: "codex uses exec", backend: BackendCodex, wantFirst: "exec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := backendInvocationArgs(AgentInvocation{Backend: tt.backend, Prompt: "hello"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d", len(args))
			}

			if args[0] != tt.wantFirst {
				t.Fatalf("expected first arg %q, got %q", tt.wantFirst, args[0])
			}

			if args[1] != "hello" {
				t.Fatalf("expected prompt as second arg, got %q", args[1])
			}
		})
	}
}

func TestBackendInvocationArgs_RejectsUnsupportedBackend(t *testing.T) {
	_, err := backendInvocationArgs(AgentInvocation{Backend: "other", Prompt: "hello"})
	if err == nil {
		t.Fatal("expected error for unsupported backend")
	}

	if !strings.Contains(err.Error(), "unsupported agent backend") {
		t.Fatalf("unexpected error: %v", err)
	}
}
