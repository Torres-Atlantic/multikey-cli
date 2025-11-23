package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateKey generates a new ED25519 SSH key pair
func GenerateKey(keyPath string) error {
	// Ensure directory exists
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("key file already exists: %s", keyPath)
	}

	// Generate private key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Marshal private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Write private key
	if err := os.WriteFile(keyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// GetPublicKey extracts the public key from a private key file
func GetPublicKey(privateKeyPath string) (string, error) {
	// Use ssh-keygen to extract public key
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

