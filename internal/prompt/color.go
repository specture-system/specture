package prompt

// ANSI color codes for terminal output
const (
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
)

// Yellow wraps a string in yellow ANSI color codes.
func Yellow(s string) string {
	return ColorYellow + s + ColorReset
}
