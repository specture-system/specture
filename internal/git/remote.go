package git

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// Forge represents a Git hosting platform.
type Forge int

const (
	ForgeUnknown Forge = iota
	ForgeGitHub
	ForgeGitLab
)

// GetRemoteURL retrieves the URL for a git remote.
func GetRemoteURL(dir, remoteName string) (string, error) {
	cmd := exec.Command("git", "config", "--get", fmt.Sprintf("remote.%s.url", remoteName))
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// IdentifyForge determines which Git hosting platform a remote URL belongs to.
func IdentifyForge(remoteURL string) (Forge, error) {
	// Remove .git suffix if present
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Handle SSH URLs (git@host:...)
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) == 2 {
			host := parts[0][4:] // Remove "git@"
			return identifyForgeByHost(host), nil
		}
		return ForgeUnknown, fmt.Errorf("invalid SSH URL format: %s", remoteURL)
	}

	// Handle HTTP(S) URLs
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return ForgeUnknown, fmt.Errorf("invalid URL: %w", err)
	}

	return identifyForgeByHost(parsedURL.Host), nil
}

func identifyForgeByHost(host string) Forge {
	host = strings.ToLower(host)

	switch {
	case strings.Contains(host, "github.com"):
		return ForgeGitHub
	case host == "gitlab.com" || strings.HasSuffix(host, ".gitlab.com"):
		return ForgeGitLab
	default:
		return ForgeUnknown
	}
}

// String returns a human-readable name for the forge.
func (f Forge) String() string {
	switch f {
	case ForgeGitHub:
		return "GitHub"
	case ForgeGitLab:
		return "GitLab"
	default:
		return "Unknown"
	}
}

// GetTerminology returns the appropriate contribution type for the forge.
// Returns "merge request" for GitLab, "pull request" for others.
func GetTerminology(forge Forge) string {
	if forge == ForgeGitLab {
		return "merge request"
	}
	return "pull request"
}
