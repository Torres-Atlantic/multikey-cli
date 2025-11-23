package mapping

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
)

// Resolver resolves the profile for a given path
type Resolver struct {
	config *config.Config
}

// NewResolver creates a new mapping resolver
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{
		config: cfg,
	}
}

// ResolveProfile resolves which profile should be used for a given path
// Precedence:
// 1. Repo mapping (exact match)
// 2. Deepest matching folder mapping
// 3. Default profile (if set)
// 4. No match (returns nil)
func (r *Resolver) ResolveProfile(path string) (*config.Profile, error) {
	// Normalize path to absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Resolve symlinks
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		absPath = resolvedPath
	}

	// 1. Check repo mappings (exact match)
	for _, mapping := range r.config.RepoMappings {
		mappingAbs, err := filepath.Abs(mapping.Path)
		if err != nil {
			continue
		}
		if mappingAbs == absPath {
			return r.findProfile(mapping.ProfileID)
		}
	}

	// 2. Check folder mappings (deepest match wins)
	var bestMatch *config.FolderMapping
	var bestMatchDepth int = -1

	for _, mapping := range r.config.FolderMappings {
		mappingAbs, err := filepath.Abs(mapping.Path)
		if err != nil {
			continue
		}

		// Check if path is under this mapping
		if strings.HasPrefix(absPath+string(filepath.Separator), mappingAbs+string(filepath.Separator)) {
			// Calculate depth (number of path separators)
			depth := strings.Count(mappingAbs, string(filepath.Separator))
			if depth > bestMatchDepth {
				bestMatch = &mapping
				bestMatchDepth = depth
			}
		}
	}

	if bestMatch != nil {
		return r.findProfile(bestMatch.ProfileID)
	}

	// 3. Default profile (not implemented in v1, return nil)
	// TODO: Add default profile support if needed

	// 4. No match
	return nil, nil
}

// findProfile finds a profile by ID
func (r *Resolver) findProfile(profileID string) (*config.Profile, error) {
	for _, profile := range r.config.Profiles {
		if profile.ID == profileID {
			return &profile, nil
		}
	}
	return nil, fmt.Errorf("profile not found: %s", profileID)
}

// AddFolderMapping adds a folder mapping
func (r *Resolver) AddFolderMapping(path, profileID string) error {
	// Validate profile exists
	if _, err := r.findProfile(profileID); err != nil {
		return err
	}

	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
		return fmt.Errorf("path does not exist or is not a directory: %s", absPath)
	}

	// Check for duplicates
	for _, mapping := range r.config.FolderMappings {
		mappingAbs, _ := filepath.Abs(mapping.Path)
		if mappingAbs == absPath {
			// Update existing mapping
			mapping.ProfileID = profileID
			return nil
		}
	}

	// Add new mapping
	r.config.FolderMappings = append(r.config.FolderMappings, config.FolderMapping{
		Path:      absPath,
		ProfileID: profileID,
	})

	return nil
}

// AddRepoMapping adds a repo mapping
func (r *Resolver) AddRepoMapping(path, profileID string) error {
	// Validate profile exists
	if _, err := r.findProfile(profileID); err != nil {
		return err
	}

	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path is a repo
	// Note: We'll import git package later, for now just check if .git exists
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("path is not a Git repository: %s", absPath)
	}

	// Check for duplicates
	for _, mapping := range r.config.RepoMappings {
		mappingAbs, _ := filepath.Abs(mapping.Path)
		if mappingAbs == absPath {
			// Update existing mapping
			mapping.ProfileID = profileID
			return nil
		}
	}

	// Add new mapping
	r.config.RepoMappings = append(r.config.RepoMappings, config.RepoMapping{
		Path:      absPath,
		ProfileID: profileID,
	})

	return nil
}

// RemoveFolderMapping removes a folder mapping
func (r *Resolver) RemoveFolderMapping(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	for i, mapping := range r.config.FolderMappings {
		mappingAbs, _ := filepath.Abs(mapping.Path)
		if mappingAbs == absPath {
			r.config.FolderMappings = append(r.config.FolderMappings[:i], r.config.FolderMappings[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("folder mapping not found: %s", path)
}

// RemoveRepoMapping removes a repo mapping
func (r *Resolver) RemoveRepoMapping(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	for i, mapping := range r.config.RepoMappings {
		mappingAbs, _ := filepath.Abs(mapping.Path)
		if mappingAbs == absPath {
			r.config.RepoMappings = append(r.config.RepoMappings[:i], r.config.RepoMappings[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("repo mapping not found: %s", path)
}

