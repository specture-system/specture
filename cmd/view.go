package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	specpkg "github.com/specture-system/specture/internal/spec"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <ref>",
	Args:  cobra.ExactArgs(1),
	Short: "Open a spec in your configured editor",
	Long: `Resolve a spec reference and open its spec file in your preferred editor.

VISUAL is used when set, with EDITOR as a fallback. Editor commands may include
arguments, such as "code --wait" or "nvim -f".

Examples:
  specture view 4
  specture view 4.2`,
	RunE: runView,
}

func runView(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	path, err := specpkg.ResolveRef(filepath.Join(cwd, "specs"), args[0])
	if err != nil {
		return err
	}

	editor, err := configuredEditor()
	if err != nil {
		return err
	}

	// Run through the shell so conventional editor settings with quoted
	// arguments work, then replace the shell so the command lifetime matches
	// the editor's lifetime.
	editorCmd := exec.Command("sh", "-c", `exec `+editor+` "$1"`, "specture-editor", path)
	editorCmd.Stdin = cmd.InOrStdin()
	editorCmd.Stdout = cmd.OutOrStdout()
	editorCmd.Stderr = cmd.ErrOrStderr()
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor command failed: %w", err)
	}

	return nil
}

func configuredEditor() (string, error) {
	if editor := os.Getenv("VISUAL"); strings.TrimSpace(editor) != "" {
		return editor, nil
	}
	if editor := os.Getenv("EDITOR"); strings.TrimSpace(editor) != "" {
		return editor, nil
	}
	return "", fmt.Errorf("no editor configured: set VISUAL or EDITOR")
}
