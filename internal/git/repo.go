package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsRepo checks if a path is a Git repository
func IsRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return true
	}
	return false
}

// GetRemoteURL gets the origin remote URL for a repository
func GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SetRemoteURL sets the origin remote URL for a repository
func SetRemoteURL(repoPath, url string) error {
	cmd := exec.Command("git", "remote", "set-url", "origin", url)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}
	return nil
}

// ParseRemoteURL parses a Git remote URL and extracts the host and path
func ParseRemoteURL(url string) (host, org, repo string, err error) {
	// Handle git@host:org/repo.git format
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(strings.TrimPrefix(url, "git@"), ":")
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("invalid git@ URL format: %s", url)
		}
		host = parts[0]
		path := strings.TrimSuffix(parts[1], ".git")
		pathParts := strings.Split(path, "/")
		if len(pathParts) != 2 {
			return "", "", "", fmt.Errorf("invalid path format: %s", path)
		}
		org = pathParts[0]
		repo = pathParts[1]
		return host, org, repo, nil
	}

	// Handle https://host/org/repo.git format
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
		parts := strings.Split(url, "/")
		if len(parts) < 3 {
			return "", "", "", fmt.Errorf("invalid https URL format: %s", url)
		}
		host = parts[0]
		org = parts[1]
		repo = strings.TrimSuffix(parts[2], ".git")
		return host, org, repo, nil
	}

	return "", "", "", fmt.Errorf("unsupported URL format: %s", url)
}

// BuildRemoteURL builds a Git remote URL from components
func BuildRemoteURL(host, org, repo string) string {
	return fmt.Sprintf("git@%s:%s/%s.git", host, org, repo)
}

// GetUserEmail gets the user.email config for a repository
func GetUserEmail(repoPath string) (string, error) {
	cmd := exec.Command("git", "config", "user.email")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get user.email: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetUserName gets the user.name config for a repository
func GetUserName(repoPath string) (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get user.name: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// SetUserEmail sets the user.email config for a repository
func SetUserEmail(repoPath, email string) error {
	cmd := exec.Command("git", "config", "user.email", email)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}
	return nil
}

// SetUserName sets the user.name config for a repository
func SetUserName(repoPath, name string) error {
	cmd := exec.Command("git", "config", "user.name", name)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}
	return nil
}

