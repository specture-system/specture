package main

import (
	"strings"
	"testing"
)

func TestVersionDefaultsFromVersionFile(t *testing.T) {
	if got, want := version, strings.TrimSpace(versionFile); got != want {
		t.Fatalf("expected version default %q, got %q", want, got)
	}
}
