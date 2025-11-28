package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDir  = ".config/multikey"
	ConfigFile = "config.json"
	ConfigVersion = 1
)

// Profile represents an SSH profile configuration
type Profile struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	IdentityFile string `json:"identityFile"`
	SSHHost     string `json:"sshHost"`
}

// FolderMapping maps a folder path to a profile ID
type FolderMapping struct {
	Path      string `json:"path"`
	ProfileID string `json:"profileId"`
}

// RepoMapping maps a repository path to a profile ID
type RepoMapping struct {
	Path      string `json:"path"`
	ProfileID string `json:"profileId"`
}

// Config represents the MultiKey configuration
type Config struct {
	Profiles       []Profile       `json:"profiles"`
	FolderMappings []FolderMapping `json:"folderMappings"`
	RepoMappings   []RepoMapping   `json:"repoMappings"`
	Meta           Meta            `json:"meta"`
}

// Meta contains metadata about the configuration
type Meta struct {
	Version    int    `json:"version"`
	LicenseKey string `json:"licenseKey,omitempty"`
}

// Manager handles configuration operations
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir)
	configPath := filepath.Join(configDir, ConfigFile)

	return &Manager{
		configPath: configPath,
	}, nil
}

// GetConfigPath returns the full path to the config file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// Load reads the configuration from disk
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return m.Default(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Ensure meta version is set
	if config.Meta.Version == 0 {
		config.Meta.Version = ConfigVersion
	}

	return &config, nil
}

// Save writes the configuration to disk
func (m *Manager) Save(config *Config) error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set version if not set
	if config.Meta.Version == 0 {
		config.Meta.Version = ConfigVersion
	}

	// Write atomically
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first, then rename (atomic write)
	tmpPath := m.configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if err := os.Rename(tmpPath, m.configPath); err != nil {
		return fmt.Errorf("failed to rename config: %w", err)
	}

	return nil
}

// Default returns a default configuration
func (m *Manager) Default() *Config {
	return &Config{
		Profiles:       []Profile{},
		FolderMappings: []FolderMapping{},
		RepoMappings:   []RepoMapping{},
		Meta: Meta{
			Version: ConfigVersion,
		},
	}
}

// Exists checks if the config file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}

