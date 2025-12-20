package new

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenEditor launches the user's editor for the given file path.
// It respects the $EDITOR environment variable.
// Returns an error if no editor is configured or if the editor exits with a non-zero code.
func OpenEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("no EDITOR environment variable set")
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	return nil
}
