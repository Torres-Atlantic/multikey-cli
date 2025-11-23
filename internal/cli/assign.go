package cli

import (
	"fmt"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign <path>",
	Short: "Assign a profile to a repository",
	Long:  "Assign a profile to a repository and immediately apply the configuration.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssign,
}

var assignProfileFlag string

func init() {
	assignCmd.Flags().StringVarP(&assignProfileFlag, "profile", "p", "", "Profile ID to assign (required)")
	assignCmd.MarkFlagRequired("profile")
}

func runAssign(cmd *cobra.Command, args []string) error {
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

	// Validate profile exists
	var profile *config.Profile
	for _, p := range cfg.Profiles {
		if p.ID == assignProfileFlag {
			profile = &p
			break
		}
	}
	if profile == nil {
		return fmt.Errorf("profile not found: %s", assignProfileFlag)
	}

	// Create resolver
	resolver := mapping.NewResolver(cfg)

	// Add repo mapping
	if err := resolver.AddRepoMapping(absPath, assignProfileFlag); err != nil {
		return err
	}

	// Apply fixes
	fmt.Printf("Assigning profile '%s' to repository: %s\n\n", assignProfileFlag, absPath)

	// Fix remote URL
	remoteURL, err := git.GetRemoteURL(absPath)
	if err == nil {
		host, org, repoName, err := git.ParseRemoteURL(remoteURL)
		if err == nil && host != profile.SSHHost {
			newURL := git.BuildRemoteURL(profile.SSHHost, org, repoName)
			if err := git.SetRemoteURL(absPath, newURL); err != nil {
				return fmt.Errorf("failed to update remote URL: %w", err)
			}
			fmt.Printf("✓ Updated remote URL to %s\n", newURL)
		}
	}

	// Fix Git config
	if err := git.SetUserEmail(absPath, profile.Email); err != nil {
		return fmt.Errorf("failed to update user.email: %w", err)
	}
	fmt.Printf("✓ Updated user.email to %s\n", profile.Email)

	if err := git.SetUserName(absPath, profile.Username); err != nil {
		return fmt.Errorf("failed to update user.name: %w", err)
	}
	fmt.Printf("✓ Updated user.name to %s\n", profile.Username)

	// Save config
	if err := configMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✓ Repository assigned to profile '%s'\n", assignProfileFlag)
	return nil
}

