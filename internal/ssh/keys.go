package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateKey generates a new ED25519 SSH key pair at keyPath using ssh-keygen.
// If passphrase is non-empty the private key is encrypted with it; pass "" to
// generate an unencrypted key. comment is embedded in the public key (typically
// the profile email). The matching public key is written to keyPath + ".pub".
func GenerateKey(keyPath, comment, passphrase string) error {
	// Ensure directory exists
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("key file already exists: %s", keyPath)
	}

	// Delegate to ssh-keygen so passphrase encryption and the .pub file are
	// handled the same way the rest of the SSH toolchain expects.
	args := []string{"-t", "ed25519", "-f", keyPath, "-N", passphrase}
	if comment != "" {
		args = append(args, "-C", comment)
	}
	cmd := exec.Command("ssh-keygen", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}

// GetPublicKey returns the public key for a private key file, preferring the
// persisted .pub file (which avoids a passphrase prompt for encrypted keys) and
// falling back to deriving it from the private key.
func GetPublicKey(privateKeyPath string) (string, error) {
	// Prefer the persisted .pub file (present for keys we generated).
	if data, err := os.ReadFile(privateKeyPath + ".pub"); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Fall back to deriving the public key from the private key.
	cmd := exec.Command("ssh-keygen", "-y", "-f", privateKeyPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to extract public key: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CopyToClipboard copies text to the system clipboard
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	// Detect OS and use appropriate command
	if _, err := exec.LookPath("pbcopy"); err == nil {
		// macOS
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		// Linux with xclip
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else if _, err := exec.LookPath("xsel"); err == nil {
		// Linux with xsel
		cmd = exec.Command("xsel", "--clipboard", "--input")
	} else {
		return fmt.Errorf("no clipboard utility found (pbcopy, xclip, or xsel)")
	}

	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// KeyExists checks if a key file exists and is readable
func KeyExists(keyPath string) bool {
	info, err := os.Stat(keyPath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

