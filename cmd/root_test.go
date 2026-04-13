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
	t.Cleanup(func() {
		SetVersion("dev", "")
		rootCmd.SetArgs(nil)
	})

	SetVersion("v0.3.0", "abc1234")

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

func TestRootCommand_VersionFlagWithoutCommit(t *testing.T) {
	t.Cleanup(func() {
		SetVersion("dev", "")
		rootCmd.SetArgs(nil)
	})

	SetVersion("v0.3.0", "")

	out := &bytes.Buffer{}
	cmd := rootCmd
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected version flag to succeed, got: %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "v0.3.0" {
		t.Fatalf("expected version output %q, got %q", "v0.3.0", got)
	}
}

func TestRootCommand_VersionFlagNormalizesVersionPrefix(t *testing.T) {
	t.Cleanup(func() {
		SetVersion("dev", "")
		rootCmd.SetArgs(nil)
	})

	SetVersion("0.3.0", "abc1234")

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

func TestRootCommand_VersionFlagTruncatesCommitHash(t *testing.T) {
	t.Cleanup(func() {
		SetVersion("dev", "")
		rootCmd.SetArgs(nil)
	})

	SetVersion("v0.3.0", "c872008af8a9efb0a424d076df1798e5ff68637f")

	out := &bytes.Buffer{}
	cmd := rootCmd
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected version flag to succeed, got: %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "v0.3.0 (c872008)" {
		t.Fatalf("expected version output %q, got %q", "v0.3.0 (c872008)", got)
	}
}
