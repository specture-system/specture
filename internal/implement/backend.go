package implement

import (
	"fmt"
)

const (
	BackendOpencode = "opencode"
	BackendCodex    = "codex"
)

var supportedBackends = []string{BackendOpencode, BackendCodex}

// SelectBackend determines which supported backend to use.
// If override is set, it must be supported and available in PATH.
func SelectBackend(override string, lookPath func(file string) (string, error)) (string, error) {
	if override != "" {
		if !isSupportedBackend(override) {
			return "", fmt.Errorf("invalid agent backend %q (must be 'opencode' or 'codex')", override)
		}

		if _, err := lookPath(override); err != nil {
			return "", fmt.Errorf("agent backend %q is not available in PATH", override)
		}

		return override, nil
	}

	for _, backend := range supportedBackends {
		if _, err := lookPath(backend); err == nil {
			return backend, nil
		}
	}

	return "", fmt.Errorf("no supported agent backend found in PATH (tried: opencode, codex)")
}

func isSupportedBackend(name string) bool {
	for _, supported := range supportedBackends {
		if supported == name {
			return true
		}
	}

	return false
}

func backendInvocationArgs(invocation AgentInvocation) ([]string, error) {
	switch invocation.Backend {
	case BackendOpencode:
		return []string{"run", invocation.Prompt}, nil
	case BackendCodex:
		return []string{"exec", invocation.Prompt}, nil
	default:
		return nil, fmt.Errorf("unsupported agent backend %q for invocation", invocation.Backend)
	}
}
