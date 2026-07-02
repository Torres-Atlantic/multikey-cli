package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	SSHConfigFile     = ".ssh/config"
	MultiKeyConfigFile = ".ssh/multikey.conf"
	IncludeDirective  = "Include ~/.ssh/multikey.conf"
)

// Manager handles SSH configuration operations
type Manager struct {
	sshConfigPath     string
	multikeyConfigPath string
}

// NewManager creates a new SSH config manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &Manager{
		sshConfigPath:     filepath.Join(homeDir, SSHConfigFile),
		multikeyConfigPath: filepath.Join(homeDir, MultiKeyConfigFile),
	}, nil
}

// GetMultiKeyConfigPath returns the path to the multikey SSH config file
func (m *Manager) GetMultiKeyConfigPath() string {
	return m.multikeyConfigPath
}

// EnsureInclude ensures the SSH config includes the multikey config file
func (m *Manager) EnsureInclude() error {
	// Read existing SSH config
	data, err := os.ReadFile(m.sshConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	configContent := string(data)

	// Check if include line already exists
	if strings.Contains(configContent, IncludeDirective) {
		return nil // Already included
	}

	// Back up the existing config before mutating it, so a bad edit is recoverable.
	if len(data) > 0 {
		if err := os.WriteFile(m.sshConfigPath+".bak", data, 0600); err != nil {
			return fmt.Errorf("failed to back up SSH config: %w", err)
		}
	}

	// Prepend the include directive rather than appending it. SSH is
	// first-match-wins, so a pre-existing `Host github.com` or `Host *` block
	// higher in the file would otherwise shadow the multikey aliases.
	var newContent string
	if len(configContent) > 0 {
		newContent = IncludeDirective + "\n\n" + configContent
	} else {
		newContent = IncludeDirective + "\n"
	}

	// Ensure the .ssh directory exists.
	if err := os.MkdirAll(filepath.Dir(m.sshConfigPath), 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Write atomically: temp file then rename, mirroring the config Save pattern.
	tmpPath := m.sshConfigPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}
	if err := os.Rename(tmpPath, m.sshConfigPath); err != nil {
		return fmt.Errorf("failed to rename SSH config: %w", err)
	}

	return nil
}

// GenerateConfig generates the multikey SSH config from profiles
func (m *Manager) GenerateConfig(profiles []Profile) error {
	var sb strings.Builder

	isDarwin := runtime.GOOS == "darwin"

	for _, profile := range profiles {
		sb.WriteString(fmt.Sprintf("Host %s\n", profile.SSHHost))
		sb.WriteString("  HostName github.com\n")
		sb.WriteString("  User git\n")
		sb.WriteString(fmt.Sprintf("  IdentityFile %s\n", profile.IdentityFile))
		// IdentitiesOnly forces SSH to present ONLY this key and ignore any other
		// identities the agent offers. Without it GitHub authenticates as whichever
		// agent key it sees first — the exact wrong-account push this tool prevents.
		sb.WriteString("  IdentitiesOnly yes\n")
		if isDarwin {
			// On macOS, cache an encrypted key's passphrase in the keychain and load
			// it into the agent so the passphrase is only entered once.
			sb.WriteString("  UseKeychain yes\n")
			sb.WriteString("  AddKeysToAgent yes\n")
		}
		sb.WriteString("\n")
	}

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(m.multikeyConfigPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Write config file with owner-only permissions (SSH convention; avoids
	// StrictModes complaints).
	content := sb.String()
	if err := os.WriteFile(m.multikeyConfigPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write multikey SSH config: %w", err)
	}

	return nil
}

// Profile represents a profile for SSH config generation
type Profile struct {
	SSHHost     string
	IdentityFile string
}

// TestConnection tests SSH connectivity to GitHub
func TestConnection(identityFile string) error {
	// IdentitiesOnly=yes ensures the test isolates the key under test instead of
	// silently passing on some other agent key.
	cmd := exec.Command("ssh", "-T", "-o", "IdentitiesOnly=yes", "-i", identityFile, "git@github.com")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	
	// SSH returns exit code 1 on successful authentication (GitHub doesn't allow shell access)
	// Exit code 255 indicates connection/auth failure
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		if exitCode == 1 {
			// This is actually success - GitHub returns exit 1 with "Hi username!" message
			return nil
		}
		return fmt.Errorf("SSH test failed with exit code %d", exitCode)
	}
	
	if err != nil {
		return fmt.Errorf("SSH test failed: %w", err)
	}
	
	return nil
}

