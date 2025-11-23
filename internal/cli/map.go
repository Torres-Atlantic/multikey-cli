package cli

import (
	"fmt"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/spf13/cobra"
)

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Manage folder and repository mappings",
	Long:  "Map folders or repositories to profiles to automatically apply the correct SSH identity.",
}

var mapAddCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a mapping",
	Long:  "Map a folder or repository to a profile. If the path is a Git repository, it creates a repo mapping. Otherwise, it creates a folder mapping.",
	Args:  cobra.ExactArgs(1),
	RunE:  runMapAdd,
}

var mapListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all mappings",
	RunE:  runMapList,
}

var mapRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove a mapping",
	Args:  cobra.ExactArgs(1),
	RunE:  runMapRemove,
}

var profileFlag string

func init() {
	mapCmd.AddCommand(mapAddCmd)
	mapCmd.AddCommand(mapListCmd)
	mapCmd.AddCommand(mapRemoveCmd)

	mapAddCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "Profile ID to map to (required)")
	mapAddCmd.MarkFlagRequired("profile")
}

func runMapAdd(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
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
	profileFound := false
	for _, p := range cfg.Profiles {
		if p.ID == profileFlag {
			profileFound = true
			break
		}
	}
	if !profileFound {
		return fmt.Errorf("profile not found: %s", profileFlag)
	}

	// Create resolver
	resolver := mapping.NewResolver(cfg)

	// Check if path is a repo
	if git.IsRepo(absPath) {
		// Add repo mapping
		if err := resolver.AddRepoMapping(absPath, profileFlag); err != nil {
			return err
		}
		fmt.Printf("✓ Repository mapping added: %s -> %s\n", absPath, profileFlag)
	} else {
		// Add folder mapping
		if err := resolver.AddFolderMapping(absPath, profileFlag); err != nil {
			return err
		}
		fmt.Printf("✓ Folder mapping added: %s -> %s\n", absPath, profileFlag)
	}

	// Save config
	return configMgr.Save(cfg)
}

func runMapList(cmd *cobra.Command, args []string) error {
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	if len(cfg.FolderMappings) == 0 && len(cfg.RepoMappings) == 0 {
		fmt.Println("No mappings configured.")
		return nil
	}

	if len(cfg.FolderMappings) > 0 {
		fmt.Println("Folder Mappings:")
		for _, m := range cfg.FolderMappings {
			fmt.Printf("  %s -> %s\n", m.Path, m.ProfileID)
		}
		fmt.Println()
	}

	if len(cfg.RepoMappings) > 0 {
		fmt.Println("Repository Mappings:")
		for _, m := range cfg.RepoMappings {
			fmt.Printf("  %s -> %s\n", m.Path, m.ProfileID)
		}
	}

	return nil
}

func runMapRemove(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
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

	// Try to remove as repo mapping first
	err = resolver.RemoveRepoMapping(absPath)
	if err == nil {
		fmt.Printf("✓ Repository mapping removed: %s\n", absPath)
		return configMgr.Save(cfg)
	}

	// Try folder mapping
	err = resolver.RemoveFolderMapping(absPath)
	if err == nil {
		fmt.Printf("✓ Folder mapping removed: %s\n", absPath)
		return configMgr.Save(cfg)
	}

	return fmt.Errorf("mapping not found: %s", absPath)
}

