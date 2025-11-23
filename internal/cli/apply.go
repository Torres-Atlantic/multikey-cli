package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/Torres-Atlantic/multikey-cli/internal/scanner"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <path>",
	Short: "Apply fixes to repositories",
	Long:  "Scan repositories and automatically fix remote URLs and Git config to match their assigned profiles.",
	Args:  cobra.ExactArgs(1),
	RunE:  runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
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

	// Create resolver and scanner
	resolver := mapping.NewResolver(cfg)
	scnr := scanner.NewScanner(resolver)

	// Scan for repos
	fmt.Printf("Scanning %s for Git repositories...\n\n", absPath)
	repos, err := scnr.Scan(absPath)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("No Git repositories found.")
		return nil
	}

	// Filter repos that need fixing
	var toFix []scanner.RepoInfo
	for _, repo := range repos {
		if repo.Status == scanner.StatusMisaligned || repo.Status == scanner.StatusUnassigned {
			if repo.ExpectedProfile != nil {
				toFix = append(toFix, repo)
			}
		}
	}

	if len(toFix) == 0 {
		fmt.Println("All repositories are already aligned.")
		return nil
	}

	fmt.Printf("Found %d repository(ies) that need fixing.\n\n", len(toFix))

	// Apply fixes
	var fixed, failed int
	for _, repo := range toFix {
		fmt.Printf("Fixing: %s\n", repo.Path)

		profile := repo.ExpectedProfile

		// Fix remote URL if needed
		if repo.CurrentRemote != "" {
			host, org, repoName, err := git.ParseRemoteURL(repo.CurrentRemote)
			if err == nil && host != profile.SSHHost {
				newURL := git.BuildRemoteURL(profile.SSHHost, org, repoName)
				if err := git.SetRemoteURL(repo.Path, newURL); err != nil {
					fmt.Printf("  ⚠ Failed to update remote URL: %v\n", err)
					failed++
					continue
				}
				fmt.Printf("  ✓ Updated remote URL to %s\n", newURL)
			}
		}

		// Fix Git config
		if repo.CurrentEmail != profile.Email {
			if err := git.SetUserEmail(repo.Path, profile.Email); err != nil {
				fmt.Printf("  ⚠ Failed to update user.email: %v\n", err)
			} else {
				fmt.Printf("  ✓ Updated user.email to %s\n", profile.Email)
			}
		}

		if repo.CurrentName != profile.Username {
			if err := git.SetUserName(repo.Path, profile.Username); err != nil {
				fmt.Printf("  ⚠ Failed to update user.name: %v\n", err)
			} else {
				fmt.Printf("  ✓ Updated user.name to %s\n", profile.Username)
			}
		}

		// Add repo mapping if it doesn't exist
		hasMapping := false
		for _, m := range cfg.RepoMappings {
			mappingAbs, _ := filepath.Abs(m.Path)
			if mappingAbs == repo.Path {
				hasMapping = true
				break
			}
		}

		if !hasMapping {
			if err := resolver.AddRepoMapping(repo.Path, profile.ID); err == nil {
				fmt.Printf("  ✓ Added repository mapping\n")
			}
		}

		fixed++
		fmt.Println()
	}

	// Save config
	if err := configMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Summary: %d fixed, %d failed\n", fixed, failed)
	return nil
}

