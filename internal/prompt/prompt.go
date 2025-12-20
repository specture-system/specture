package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Confirm prompts the user for a yes/no response and returns true if yes.
func Confirm(message string) (bool, error) {
	return confirm(message, os.Stdin)
}

// ConfirmWithDefault prompts with a default answer if user just presses enter.
func ConfirmWithDefault(message string, defaultValue bool) (bool, error) {
	return confirmWithDefault(message, defaultValue, os.Stdin)
}

// PromptString asks the user for a string input.
func PromptString(prompt string) (string, error) {
	return promptString(prompt, os.Stdin)
}

// Internal implementations that accept io.Reader for testing

func confirm(message string, reader io.Reader) (bool, error) {
	scanner := bufio.NewScanner(reader)
	for {
		fmt.Printf("%s (yes/no): ", message)
		if !scanner.Scan() {
			return false, fmt.Errorf("failed to read input")
		}
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch response {
		case "yes", "y":
			return true, nil
		case "no", "n":
			return false, nil
		default:
			fmt.Println("Please answer 'yes' or 'no'")
		}
	}
}

func confirmWithDefault(message string, defaultValue bool, reader io.Reader) (bool, error) {
	scanner := bufio.NewScanner(reader)
	defaultStr := "y/n"
	if defaultValue {
		defaultStr = "Y/n"
	} else {
		defaultStr = "y/N"
	}
	for {
		fmt.Printf("%s (%s): ", message, defaultStr)
		if !scanner.Scan() {
			return false, fmt.Errorf("failed to read input")
		}
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response == "" {
			return defaultValue, nil
		}
		switch response {
		case "yes", "y":
			return true, nil
		case "no", "n":
			return false, nil
		default:
			fmt.Println("Please answer 'yes'/'y' or 'no'/'n', or press Enter for default")
		}
	}
}

func promptString(prompt string, reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	fmt.Print(prompt)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}
	return strings.TrimSpace(scanner.Text()), nil
}
