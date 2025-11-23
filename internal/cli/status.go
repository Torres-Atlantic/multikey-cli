package cli

import (
	"fmt"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <path>",
	Short: "Show repository status",
	Long:  "Display the current state of a repository versus its expected configuration.",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if it's a repo
	if !git.IsRepo(absPath) {
		return fmt.Errorf("path is not a Git repository: %s", absPath)
	}

	// Load config
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	// Create resolver
	resolver := mapping.NewResolver(cfg)

	// Resolve expected profile
	expectedProfile, err := resolver.ResolveProfile(absPath)
	if err != nil {
		return err
	}

	// Get current state
	currentRemote, _ := git.GetRemoteURL(absPath)
	currentEmail, _ := git.GetUserEmail(absPath)
	currentName, _ := git.GetUserName(absPath)

	var currentHost string
	if currentRemote != "" {
		host, _, _, err := git.ParseRemoteURL(currentRemote)
		if err == nil {
			currentHost = host
		}
	}

	// Display status
	fmt.Printf("Repository: %s\n\n", absPath)

	if expectedProfile == nil {
		fmt.Println("Expected Profile: (none - repository is not mapped)")
		fmt.Println("\nTo assign a profile, run:")
		fmt.Printf("  multikey assign %s --profile <profile-id>\n", absPath)
	} else {
		fmt.Printf("Expected Profile: %s\n", expectedProfile.ID)
		fmt.Printf("  Email:    %s\n", expectedProfile.Email)
		fmt.Printf("  Username: %s\n", expectedProfile.Username)
		fmt.Printf("  SSH Host: %s\n", expectedProfile.SSHHost)
	}

	fmt.Println("\nCurrent Configuration:")
	if currentRemote != "" {
		fmt.Printf("  Remote URL: %s\n", currentRemote)
		if currentHost != "" {
			fmt.Printf("  Remote Host: %s\n", currentHost)
		}
	} else {
		fmt.Println("  Remote URL: (not configured)")
	}
	if currentEmail != "" {
		fmt.Printf("  user.email: %s\n", currentEmail)
	} else {
		fmt.Println("  user.email: (not configured)")
	}
	if currentName != "" {
		fmt.Printf("  user.name: %s\n", currentName)
	} else {
		fmt.Println("  user.name: (not configured)")
	}

	// Compare and show issues
	if expectedProfile != nil {
		fmt.Println("\nAlignment:")
		aligned := true

		if currentHost != "" && currentHost != expectedProfile.SSHHost {
			fmt.Printf("  ⚠ Remote host mismatch: %s (expected %s)\n", currentHost, expectedProfile.SSHHost)
			aligned = false
		}

		if currentEmail != "" && currentEmail != expectedProfile.Email {
			fmt.Printf("  ⚠ Email mismatch: %s (expected %s)\n", currentEmail, expectedProfile.Email)
			aligned = false
		}

		if currentName != "" && currentName != expectedProfile.Username {
			fmt.Printf("  ⚠ Name mismatch: %s (expected %s)\n", currentName, expectedProfile.Username)
			aligned = false
		}

		if aligned {
			fmt.Println("  ✓ Repository is aligned with expected profile")
		} else {
			fmt.Println("\nTo fix, run:")
			fmt.Printf("  multikey apply %s\n", absPath)
		}
	}

	return nil
}

