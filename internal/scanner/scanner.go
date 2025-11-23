package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
)

// RepoInfo contains information about a scanned repository
type RepoInfo struct {
	Path           string
	CurrentRemote  string
	CurrentHost    string
	CurrentEmail   string
	CurrentName    string
	ExpectedProfile *config.Profile
	Status         RepoStatus
	Errors         []string
}

// RepoStatus represents the alignment status of a repository
type RepoStatus string

const (
	StatusAligned   RepoStatus = "aligned"
	StatusMisaligned RepoStatus = "misaligned"
	StatusUnassigned RepoStatus = "unassigned"
)

// Scanner scans directories for Git repositories
type Scanner struct {
	resolver *mapping.Resolver
}

// NewScanner creates a new scanner
func NewScanner(resolver *mapping.Resolver) *Scanner {
	return &Scanner{
		resolver: resolver,
	}
}

// Scan scans a path for Git repositories
func (s *Scanner) Scan(rootPath string) ([]RepoInfo, error) {
	var repos []RepoInfo

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip .git directories themselves
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if this is a Git repository
		if !info.IsDir() {
			return nil
		}

		if !git.IsRepo(path) {
			return nil
		}

		// Analyze this repository
		repoInfo, err := s.analyzeRepo(path)
		if err != nil {
			// Add repo with error
			repos = append(repos, RepoInfo{
				Path:   path,
				Status: StatusUnassigned,
				Errors: []string{err.Error()},
			})
			return nil // Continue scanning
		}

		repos = append(repos, *repoInfo)

		// Skip subdirectories of this repo (avoid scanning nested repos unless needed)
		return filepath.SkipDir
	})

	return repos, err
}

// analyzeRepo analyzes a single repository
func (s *Scanner) analyzeRepo(repoPath string) (*RepoInfo, error) {
	info := &RepoInfo{
		Path:   repoPath,
		Errors: []string{},
	}

	// Get current remote
	remoteURL, err := git.GetRemoteURL(repoPath)
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("failed to get remote URL: %v", err))
	} else {
		info.CurrentRemote = remoteURL
		host, _, _, err := git.ParseRemoteURL(remoteURL)
		if err == nil {
			info.CurrentHost = host
		}
	}

	// Get current Git config
	email, err := git.GetUserEmail(repoPath)
	if err == nil {
		info.CurrentEmail = email
	}

	name, err := git.GetUserName(repoPath)
	if err == nil {
		info.CurrentName = name
	}

	// Resolve expected profile
	profile, err := s.resolver.ResolveProfile(repoPath)
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("failed to resolve profile: %v", err))
	}

	if profile == nil {
		info.Status = StatusUnassigned
		return info, nil
	}

	// Store profile reference
	info.ExpectedProfile = profile

	// Determine status
	isAligned := true

	// Check remote host
	if info.CurrentHost != profile.SSHHost && info.CurrentHost != "" {
		isAligned = false
	}

	// Check email
	if info.CurrentEmail != profile.Email && info.CurrentEmail != "" {
		isAligned = false
	}

	// Check name
	if info.CurrentName != profile.Username && info.CurrentName != "" {
		isAligned = false
	}

	if isAligned {
		info.Status = StatusAligned
	} else {
		info.Status = StatusMisaligned
	}

	return info, nil
}

