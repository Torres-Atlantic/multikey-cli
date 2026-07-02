package cli

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/Torres-Atlantic/multikey-cli/internal/profile"
	"github.com/Torres-Atlantic/multikey-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run guided setup",
	Long:  "Interactive setup wizard to configure MultiKey CLI for the first time.",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("MultiKey CLI Setup")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("This wizard will help you set up MultiKey CLI.")
	fmt.Println("You can run this setup again anytime with 'multikey setup'.")
	fmt.Println()

	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	profileMgr, err := profile.NewManager()
	if err != nil {
		return err
	}

	// Step 1: Create first profile
	fmt.Println("Step 1: Create your first profile")
	fmt.Println("----------------------------------")

	var createProfile bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Create a profile now?",
		Default: true,
	}, &createProfile); err != nil {
		return err
	}

	for createProfile {
		// Reuse profile add logic
		if err := createProfileInteractive(profileMgr); err != nil {
			return err
		}

		var another bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Create another profile?",
			Default: false,
		}, &another); err != nil {
			return err
		}

		if !another {
			break
		}
	}

	// Step 2: Folder mapping
	fmt.Println()
	fmt.Println("Step 2: Map folders to profiles")
	fmt.Println("------------------------------")

	var mapFolder bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Map a folder to a profile?",
		Default: true,
	}, &mapFolder); err != nil {
		return err
	}

	for mapFolder {
		cfg, err := configMgr.Load()
		if err != nil {
			return err
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles available. Create a profile first.")
			break
		}

		// List profiles
		profileOptions := make([]string, len(cfg.Profiles))
		for i, p := range cfg.Profiles {
			profileOptions[i] = fmt.Sprintf("%s (%s)", p.ID, p.Email)
		}

		var selectedProfile string
		if err := survey.AskOne(&survey.Select{
			Message: "Select a profile:",
			Options: profileOptions,
		}, &selectedProfile); err != nil {
			return err
		}

		// Extract profile ID
		profileID := cfg.Profiles[0].ID
		for _, p := range cfg.Profiles {
			if fmt.Sprintf("%s (%s)", p.ID, p.Email) == selectedProfile {
				profileID = p.ID
				break
			}
		}

		var folderPath string
		if err := survey.AskOne(&survey.Input{
			Message: "Folder path to map:",
		}, &folderPath, survey.WithValidator(survey.Required)); err != nil {
			return err
		}

		absPath, err := filepath.Abs(folderPath)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Use mapping logic directly
		resolver := mapping.NewResolver(cfg)
		if err := resolver.AddFolderMapping(absPath, profileID); err != nil {
			fmt.Printf("⚠ Failed to add mapping: %v\n", err)
		} else {
			fmt.Printf("✓ Mapped %s to profile %s\n", absPath, profileID)
			if err := configMgr.Save(cfg); err != nil {
				return err
			}
		}

		var another bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Map another folder?",
			Default: false,
		}, &another); err != nil {
			return err
		}

		if !another {
			break
		}
	}

	// Step 3: Scan and apply
	fmt.Println()
	fmt.Println("Step 3: Scan and fix repositories")
	fmt.Println("---------------------------------")

	var scanAndFix bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Scan mapped folders for repositories and apply fixes?",
		Default: true,
	}, &scanAndFix); err != nil {
		return err
	}

	if scanAndFix {
		cfg, err := configMgr.Load()
		if err != nil {
			return err
		}

		if len(cfg.FolderMappings) > 0 {
			for _, mapping := range cfg.FolderMappings {
				fmt.Printf("\nScanning: %s\n", mapping.Path)
				// Note: In a real setup, we'd call the apply logic here
				// For now, just suggest the user run it manually
				fmt.Printf("  Run 'multikey apply %s' to fix repositories\n", mapping.Path)
			}
		} else {
			fmt.Println("No folder mappings to scan.")
		}
	}

	// Final summary
	fmt.Println()
	fmt.Println("Setup Complete!")
	fmt.Println("===============")
	fmt.Println()

	cfg, _ := configMgr.Load()
	fmt.Printf("Profiles created: %d\n", len(cfg.Profiles))
	fmt.Printf("Folder mappings: %d\n", len(cfg.FolderMappings))
	fmt.Printf("Repo mappings: %d\n", len(cfg.RepoMappings))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  • Run 'multikey profile list' to see your profiles")
	fmt.Println("  • Run 'multikey scan <path>' to check repository alignment")
	fmt.Println("  • Run 'multikey apply <path>' to fix repositories")
	fmt.Println()

	return nil
}

func createProfileInteractive(profileMgr *profile.Manager) error {
	var answers struct {
		ID          string
		Email       string
		Username    string
		KeyChoice   string
		KeyPath     string
	}

	// Prompt for profile ID
	if err := survey.AskOne(&survey.Input{
		Message: "Profile ID (e.g., 'work', 'personal'):",
	}, &answers.ID, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Check if profile already exists
	existing, _ := profileMgr.GetProfile(answers.ID)
	if existing != nil {
		fmt.Printf("Profile '%s' already exists. Skipping.\n", answers.ID)
		return nil
	}

	// Prompt for email
	if err := survey.AskOne(&survey.Input{
		Message: "Email address:",
	}, &answers.Email, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Prompt for username
	if err := survey.AskOne(&survey.Input{
		Message: "GitHub username:",
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

	var keyPath string
	var err error

	if answers.KeyChoice == "Generate new key" {
		keyPath, err = profileMgr.GenerateKeyPath(answers.ID)
		if err != nil {
			return err
		}

		passphrase, perr := promptPassphrase()
		if perr != nil {
			return perr
		}

		fmt.Printf("Generating SSH key at: %s\n", keyPath)
		if err := ssh.GenerateKey(keyPath, answers.Email, passphrase); err != nil {
			return fmt.Errorf("failed to generate SSH key: %w", err)
		}
		fmt.Println("✓ SSH key generated")
	} else {
		if err := survey.AskOne(&survey.Input{
			Message: "Path to existing SSH private key:",
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
	sshHost := profileMgr.GenerateSSHHost(answers.ID)

	// Create profile
	newProfile := config.Profile{
		ID:          answers.ID,
		Email:       answers.Email,
		Username:    answers.Username,
		IdentityFile: keyPath,
		SSHHost:     sshHost,
	}

	// Add profile
	if err := profileMgr.AddProfile(newProfile); err != nil {
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
	} else {
		fmt.Println("\n✓ Public key copied to clipboard")
	}

	// Prompt to add to GitHub
	fmt.Println("\nPlease add this public key to your GitHub account:")
	fmt.Println("  https://github.com/settings/keys")
	fmt.Println()
	var continueSetup bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Press Enter after adding the key to GitHub, then we'll test the connection",
		Default: true,
	}, &continueSetup); err != nil {
		return err
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

