package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// Note: This test intentionally does not use t.Parallel() because rootCmd is a
// global Cobra command that maintains mutable state (SetOut/SetErr/SetArgs and
// the version fields). Parallel execution would cause tests to interfere with
// each other.
func TestRootCommand_VersionFlag(t *testing.T) {
	oldVersion := Version
	oldCommit := Commit
	t.Cleanup(func() {
		Version = oldVersion
		Commit = oldCommit
		refreshVersion()
		rootCmd.SetArgs(nil)
	})

	Version = "v0.3.0"
	Commit = "abc1234"
	refreshVersion()

	out := &bytes.Buffer{}
	cmd := rootCmd
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected version flag to succeed, got: %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "v0.3.0 (abc1234)" {
		t.Fatalf("expected version output %q, got %q", "v0.3.0 (abc1234)", got)
	}
}
