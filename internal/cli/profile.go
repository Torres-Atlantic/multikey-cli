package cli

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/profile"
	"github.com/Torres-Atlantic/multikey-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage SSH profiles",
	Long:  "Create, list, edit, and remove SSH profiles for different GitHub accounts.",
}

var profileAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new SSH profile",
	Long:  "Interactively create a new SSH profile with email, username, and SSH key.",
	RunE:  runProfileAdd,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE:  runProfileList,
}

var profileEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an existing profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileEdit,
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileRemove,
}

func init() {
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileEditCmd)
	profileCmd.AddCommand(profileRemoveCmd)
}

func runProfileAdd(cmd *cobra.Command, args []string) error {
	mgr, err := profile.NewManager()
	if err != nil {
		return err
	}

	var answers struct {
		ID          string
		Email       string
		Username    string
		KeyChoice   string
		KeyPath     string
		GenerateKey bool
	}

	// Prompt for profile ID
	if err := survey.AskOne(&survey.Input{
		Message: "Profile ID (e.g., 'work', 'personal'):",
		Help:    "A unique identifier for this profile",
	}, &answers.ID, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Check if profile already exists
	existing, _ := mgr.GetProfile(answers.ID)
	if existing != nil {
		return fmt.Errorf("profile with ID '%s' already exists", answers.ID)
	}

	// Prompt for email
	if err := survey.AskOne(&survey.Input{
		Message: "Email address:",
		Help:    "Git email address for this profile",
	}, &answers.Email, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for username
	if err := survey.AskOne(&survey.Input{
		Message: "GitHub username:",
		Help:    "GitHub username for this profile",
	}, &answers.Username, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for key choice
	if err := survey.AskOne(&survey.Select{
		Message: "SSH key:",
		Options: []string{"Generate new key", "Use existing key"},
	}, &answers.KeyChoice); err != nil {
		return err
	}

	answers.GenerateKey = answers.KeyChoice == "Generate new key"

	var keyPath string
	if answers.GenerateKey {
		// Generate key path
		keyPath, err = mgr.GenerateKeyPath(answers.ID)
		if err != nil {
			return err
		}

		fmt.Printf("Generating SSH key at: %s\n", keyPath)
		if err := ssh.GenerateKey(keyPath); err != nil {
			return fmt.Errorf("failed to generate SSH key: %w", err)
		}
		fmt.Println("✓ SSH key generated")
	} else {
		// Prompt for existing key path
		if err := survey.AskOne(&survey.Input{
			Message: "Path to existing SSH private key:",
			Help:    "Full path to your existing SSH private key file",
		}, &answers.KeyPath, survey.WithValidator(survey.Required)); err != nil {
			return err
		}

		keyPath, err = filepath.Abs(answers.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to resolve key path: %w", err)
		}

		if !ssh.KeyExists(keyPath) {
			return fmt.Errorf("key file does not exist: %s", keyPath)
		}
	}

	// Generate SSH host
	sshHost := mgr.GenerateSSHHost(answers.ID)

	// Create profile
	newProfile := config.Profile{
		ID:          answers.ID,
		Email:       answers.Email,
		Username:    answers.Username,
		IdentityFile: keyPath,
		SSHHost:     sshHost,
	}

	// Add profile
	if err := mgr.AddProfile(newProfile); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' created\n", answers.ID)

	// Get and display public key
	pubKey, err := ssh.GetPublicKey(keyPath)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	fmt.Println("\nPublic key:")
	fmt.Println(pubKey)

	// Copy to clipboard
	if err := ssh.CopyToClipboard(pubKey); err != nil {
		fmt.Printf("⚠ Could not copy to clipboard: %v\n", err)
		fmt.Println("Please copy the public key above manually.")
	} else {
		fmt.Println("\n✓ Public key copied to clipboard")
	}

	// Test SSH connection
	fmt.Println("\nTesting SSH connection to GitHub...")
	if err := ssh.TestConnection(keyPath); err != nil {
		fmt.Printf("⚠ SSH test failed: %v\n", err)
		fmt.Println("Make sure you've added the public key to your GitHub account.")
	} else {
		fmt.Println("✓ SSH connection test successful")
	}

	return nil
}

func runProfileList(cmd *cobra.Command, args []string) error {
	mgr, err := profile.NewManager()
	if err != nil {
		return err
	}

	profiles, err := mgr.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("Run 'multikey profile add' to create your first profile.")
		return nil
	}

	fmt.Printf("Found %d profile(s):\n\n", len(profiles))
	for _, p := range profiles {
		fmt.Printf("ID:          %s\n", p.ID)
		fmt.Printf("Email:       %s\n", p.Email)
		fmt.Printf("Username:    %s\n", p.Username)
		fmt.Printf("SSH Host:    %s\n", p.SSHHost)
		fmt.Printf("Key File:    %s\n", p.IdentityFile)
		fmt.Println()
	}

	return nil
}

func runProfileEdit(cmd *cobra.Command, args []string) error {
	profileID := args[0]

	mgr, err := profile.NewManager()
	if err != nil {
		return err
	}

	existing, err := mgr.GetProfile(profileID)
	if err != nil {
		return err
	}

	var answers struct {
		Email    string
		Username string
		KeyPath  string
	}

	// Prompt for email (pre-filled)
	if err := survey.AskOne(&survey.Input{
		Message: "Email address:",
		Default: existing.Email,
	}, &answers.Email, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for username (pre-filled)
	if err := survey.AskOne(&survey.Input{
		Message: "GitHub username:",
		Default: existing.Username,
	}, &answers.Username, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for key path (pre-filled)
	if err := survey.AskOne(&survey.Input{
		Message: "SSH key path:",
		Default: existing.IdentityFile,
	}, &answers.KeyPath, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	keyPath, err := filepath.Abs(answers.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to resolve key path: %w", err)
	}

	if !ssh.KeyExists(keyPath) {
		return fmt.Errorf("key file does not exist: %s", keyPath)
	}

	// Update profile
	updated := config.Profile{
		ID:          existing.ID,
		Email:       answers.Email,
		Username:    answers.Username,
		IdentityFile: keyPath,
		SSHHost:     existing.SSHHost, // Don't allow changing SSH host
	}

	if err := mgr.UpdateProfile(profileID, updated); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' updated\n", profileID)
	return nil
}

func runProfileRemove(cmd *cobra.Command, args []string) error {
	profileID := args[0]

	mgr, err := profile.NewManager()
	if err != nil {
		return err
	}

	// Check if profile exists
	_, err = mgr.GetProfile(profileID)
	if err != nil {
		return err
	}

	// Confirm deletion
	var confirm bool
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to remove profile '%s'?", profileID),
		Default: false,
	}, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := mgr.RemoveProfile(profileID); err != nil {
		return err
	}

	fmt.Printf("✓ Profile '%s' removed\n", profileID)
	return nil
}

