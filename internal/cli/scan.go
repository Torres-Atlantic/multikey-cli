package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/Torres-Atlantic/multikey-cli/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <path>",
	Short: "Scan for Git repositories",
	Long:  "Scan a directory for Git repositories and show their alignment status.",
	Args:  cobra.ExactArgs(1),
	RunE:  runScan,
}

func runScan(cmd *cobra.Command, args []string) error {
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

	// Count by status
	var aligned, misaligned, unassigned int
	for _, repo := range repos {
		switch repo.Status {
		case scanner.StatusAligned:
			aligned++
		case scanner.StatusMisaligned:
			misaligned++
		case scanner.StatusUnassigned:
			unassigned++
		}
	}

	// Print summary
	fmt.Printf("Found %d repository(ies):\n", len(repos))
	fmt.Printf("  ✓ Aligned:   %d\n", aligned)
	fmt.Printf("  ⚠ Misaligned: %d\n", misaligned)
	fmt.Printf("  ? Unassigned: %d\n\n", unassigned)

	// Print details
	for _, repo := range repos {
		fmt.Printf("Repository: %s\n", repo.Path)
		fmt.Printf("  Status: %s\n", repo.Status)

		if repo.ExpectedProfile != nil {
			fmt.Printf("  Expected Profile: %s (%s)\n", repo.ExpectedProfile.ID, repo.ExpectedProfile.Email)
		} else {
			fmt.Printf("  Expected Profile: (none)\n")
		}

		if repo.CurrentRemote != "" {
			fmt.Printf("  Current Remote: %s\n", repo.CurrentRemote)
		}
		if repo.CurrentEmail != "" {
			fmt.Printf("  Current Email: %s\n", repo.CurrentEmail)
		}
		if repo.CurrentName != "" {
			fmt.Printf("  Current Name: %s\n", repo.CurrentName)
		}

		if len(repo.Errors) > 0 {
			fmt.Printf("  Errors:\n")
			for _, err := range repo.Errors {
				fmt.Printf("    - %s\n", err)
			}
		}

		// Show what needs to be fixed
		if repo.Status == scanner.StatusMisaligned && repo.ExpectedProfile != nil {
			fmt.Printf("  Fixes needed:\n")
			if repo.CurrentHost != repo.ExpectedProfile.SSHHost {
				fmt.Printf("    - Update remote URL to use %s\n", repo.ExpectedProfile.SSHHost)
			}
			if repo.CurrentEmail != repo.ExpectedProfile.Email {
				fmt.Printf("    - Update git config user.email to %s\n", repo.ExpectedProfile.Email)
			}
			if repo.CurrentName != repo.ExpectedProfile.Username {
				fmt.Printf("    - Update git config user.name to %s\n", repo.ExpectedProfile.Username)
			}
		}

		fmt.Println()
	}

	return nil
}

