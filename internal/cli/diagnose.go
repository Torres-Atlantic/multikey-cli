package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/git"
	"github.com/Torres-Atlantic/multikey-cli/internal/mapping"
	"github.com/Torres-Atlantic/multikey-cli/internal/scanner"
	"github.com/Torres-Atlantic/multikey-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose <path>",
	Short: "Run full health check",
	Long:  "Perform a comprehensive health check on configuration, SSH setup, and repositories.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDiagnose,
}

func runDiagnose(cmd *cobra.Command, args []string) error {
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

	fmt.Println("MultiKey Health Check")
	fmt.Println("====================")
	fmt.Println()

	// Load config
	configMgr, err := config.NewManager()
	if err != nil {
		fmt.Printf("✗ Failed to load config: %v\n", err)
		return err
	}

	cfg, err := configMgr.Load()
	if err != nil {
		fmt.Printf("✗ Failed to load config: %v\n", err)
		return err
	}

	fmt.Println("1. Configuration")
	fmt.Printf("   Config file: %s\n", configMgr.GetConfigPath())
	if configMgr.Exists() {
		fmt.Printf("   ✓ Config file exists\n")
	} else {
		fmt.Printf("   ⚠ Config file does not exist\n")
	}
	fmt.Printf("   Profiles: %d\n", len(cfg.Profiles))
	fmt.Printf("   Folder mappings: %d\n", len(cfg.FolderMappings))
	fmt.Printf("   Repo mappings: %d\n", len(cfg.RepoMappings))
	fmt.Println()

	// Check SSH config
	fmt.Println("2. SSH Configuration")
	sshMgr, err := ssh.NewManager()
	if err != nil {
		fmt.Printf("   ✗ Failed to initialize SSH manager: %v\n", err)
	} else {
		fmt.Printf("   SSH config: %s\n", sshMgr.GetMultiKeyConfigPath())
		
		// Check if include line exists
		if err := sshMgr.EnsureInclude(); err != nil {
			fmt.Printf("   ⚠ Failed to ensure SSH include: %v\n", err)
		} else {
			fmt.Printf("   ✓ SSH include line configured\n")
		}
	}
	fmt.Println()

	// Check profiles
	fmt.Println("3. Profile Health")
	if len(cfg.Profiles) == 0 {
		fmt.Println("   ⚠ No profiles configured")
	} else {
		for _, profile := range cfg.Profiles {
			fmt.Printf("   Profile: %s\n", profile.ID)
			
			// Check key file
			if ssh.KeyExists(profile.IdentityFile) {
				fmt.Printf("     ✓ Key file exists: %s\n", profile.IdentityFile)
				
				// Test SSH connection
				if err := ssh.TestConnection(profile.IdentityFile); err != nil {
					fmt.Printf("     ⚠ SSH test failed: %v\n", err)
				} else {
					fmt.Printf("     ✓ SSH connection test passed\n")
				}
			} else {
				fmt.Printf("     ✗ Key file missing: %s\n", profile.IdentityFile)
			}
		}
	}
	fmt.Println()

	// Check if path is a repo or folder
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if git.IsRepo(absPath) {
			// Single repo diagnosis
			fmt.Println("4. Repository Status")
			runStatusForPath(absPath, cfg)
		} else {
			// Folder scan
			fmt.Println("4. Repository Scan")
			resolver := mapping.NewResolver(cfg)
			scnr := scanner.NewScanner(resolver)
			
			repos, err := scnr.Scan(absPath)
			if err != nil {
				fmt.Printf("   ✗ Scan failed: %v\n", err)
			} else {
				fmt.Printf("   Found %d repository(ies)\n", len(repos))
				
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
				
				fmt.Printf("     ✓ Aligned: %d\n", aligned)
				fmt.Printf("     ⚠ Misaligned: %d\n", misaligned)
				fmt.Printf("     ? Unassigned: %d\n", unassigned)
				
				if misaligned > 0 || unassigned > 0 {
					fmt.Println("\n   To fix, run:")
					fmt.Printf("     multikey apply %s\n", absPath)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("Diagnosis complete.")
	return nil
}

func runStatusForPath(path string, cfg *config.Config) {
	resolver := mapping.NewResolver(cfg)
	expectedProfile, _ := resolver.ResolveProfile(path)

	currentRemote, _ := git.GetRemoteURL(path)
	currentEmail, _ := git.GetUserEmail(path)
	currentName, _ := git.GetUserName(path)

	if expectedProfile == nil {
		fmt.Println("   ⚠ Repository is not mapped to any profile")
	} else {
		fmt.Printf("   Expected Profile: %s\n", expectedProfile.ID)
		
		aligned := true
		if currentRemote != "" {
			host, _, _, err := git.ParseRemoteURL(currentRemote)
			if err == nil && host != expectedProfile.SSHHost {
				fmt.Printf("   ⚠ Remote host mismatch\n")
				aligned = false
			}
		}
		if currentEmail != expectedProfile.Email {
			fmt.Printf("   ⚠ Email mismatch\n")
			aligned = false
		}
		if currentName != expectedProfile.Username {
			fmt.Printf("   ⚠ Name mismatch\n")
			aligned = false
		}
		
		if aligned {
			fmt.Println("   ✓ Repository is aligned")
		}
	}
}

