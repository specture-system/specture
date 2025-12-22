package new

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenEditor launches the user's editor for the given file path.
// It respects the $EDITOR environment variable.
// Returns an error if no editor is configured or if the editor exits with a non-zero code.
//
// Note: The editor is given /dev/tty for stdin so it can interact with the terminal
// without consuming the application's stdin, allowing subsequent prompts to work.
func OpenEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("no EDITOR environment variable set")
	}

	cmd := exec.Command(editor, filePath)
	// Open /dev/tty for the editor so it can interact with the terminal
	// This prevents consuming the application's stdin
	if tty, err := os.Open("/dev/tty"); err == nil {
		cmd.Stdin = tty
		defer tty.Close()
	} else {
		// Fallback: if /dev/tty is not available, don't set stdin
		// (useful for testing and non-interactive environments)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	return nil
}
