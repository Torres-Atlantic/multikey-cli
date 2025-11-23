package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/ssh"
)

// Manager handles profile operations
type Manager struct {
	configManager *config.Manager
	sshManager    *ssh.Manager
}

// NewManager creates a new profile manager
func NewManager() (*Manager, error) {
	configMgr, err := config.NewManager()
	if err != nil {
		return nil, err
	}

	sshMgr, err := ssh.NewManager()
	if err != nil {
		return nil, err
	}

	return &Manager{
		configManager: configMgr,
		sshManager:    sshMgr,
	}, nil
}

// AddProfile adds a new profile
func (m *Manager) AddProfile(profile config.Profile) error {
	cfg, err := m.configManager.Load()
	if err != nil {
		return err
	}

	// Validate uniqueness
	for _, p := range cfg.Profiles {
		if p.ID == profile.ID {
			return fmt.Errorf("profile with ID '%s' already exists", profile.ID)
		}
		if p.SSHHost == profile.SSHHost {
			return fmt.Errorf("SSH host '%s' already in use", profile.SSHHost)
		}
	}

	// Add profile
	cfg.Profiles = append(cfg.Profiles, profile)

	// Save config
	if err := m.configManager.Save(cfg); err != nil {
		return err
	}

	// Update SSH config
	return m.updateSSHConfig(cfg)
}

// ListProfiles returns all profiles
func (m *Manager) ListProfiles() ([]config.Profile, error) {
	cfg, err := m.configManager.Load()
	if err != nil {
		return nil, err
	}
	return cfg.Profiles, nil
}

// GetProfile gets a profile by ID
func (m *Manager) GetProfile(id string) (*config.Profile, error) {
	cfg, err := m.configManager.Load()
	if err != nil {
		return nil, err
	}

	for _, p := range cfg.Profiles {
		if p.ID == id {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("profile not found: %s", id)
}

// UpdateProfile updates an existing profile
func (m *Manager) UpdateProfile(id string, updates config.Profile) error {
	cfg, err := m.configManager.Load()
	if err != nil {
		return err
	}

	// Find and update profile
	found := false
	for i, p := range cfg.Profiles {
		if p.ID == id {
			// Preserve ID
			updates.ID = id
			cfg.Profiles[i] = updates
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Validate uniqueness of SSH host
	for _, p := range cfg.Profiles {
		if p.ID != id && p.SSHHost == updates.SSHHost {
			return fmt.Errorf("SSH host '%s' already in use", updates.SSHHost)
		}
	}

	// Save config
	if err := m.configManager.Save(cfg); err != nil {
		return err
	}

	// Update SSH config
	return m.updateSSHConfig(cfg)
}

// RemoveProfile removes a profile
func (m *Manager) RemoveProfile(id string) error {
	cfg, err := m.configManager.Load()
	if err != nil {
		return err
	}

	// Find and remove profile
	found := false
	for i, p := range cfg.Profiles {
		if p.ID == id {
			cfg.Profiles = append(cfg.Profiles[:i], cfg.Profiles[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Save config
	if err := m.configManager.Save(cfg); err != nil {
		return err
	}

	// Update SSH config
	return m.updateSSHConfig(cfg)
}

// GenerateSSHHost generates a unique SSH host alias for a profile ID
func (m *Manager) GenerateSSHHost(profileID string) string {
	// Clean profile ID for use in SSH host
	host := strings.ToLower(profileID)
	host = strings.ReplaceAll(host, " ", "-")
	host = strings.ReplaceAll(host, "_", "-")
	return fmt.Sprintf("github-%s", host)
}

// GenerateKeyPath generates a default key path for a profile
func (m *Manager) GenerateKeyPath(profileID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	keyName := fmt.Sprintf("id_ed25519_%s", profileID)
	return filepath.Join(homeDir, ".ssh", keyName), nil
}

// updateSSHConfig updates the SSH config file with all profiles
func (m *Manager) updateSSHConfig(cfg *config.Config) error {
	// Ensure include line exists
	if err := m.sshManager.EnsureInclude(); err != nil {
		return fmt.Errorf("failed to ensure SSH include: %w", err)
	}

	// Convert profiles to SSH config format
	sshProfiles := make([]ssh.Profile, len(cfg.Profiles))
	for i, p := range cfg.Profiles {
		sshProfiles[i] = ssh.Profile{
			SSHHost:     p.SSHHost,
			IdentityFile: p.IdentityFile,
		}
	}

	// Generate SSH config
	return m.sshManager.GenerateConfig(sshProfiles)
}

