package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/Torres-Atlantic/multikey-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import configuration",
	Long:  "Import configuration from a JSON file.",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var importedConfig config.Config
	if err := json.Unmarshal(data, &importedConfig); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate
	if len(importedConfig.Profiles) == 0 {
		return fmt.Errorf("imported config has no profiles")
	}

	// Load current config
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	currentCfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	// Check for conflicts
	if len(currentCfg.Profiles) > 0 {
		var merge bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Merge with existing configuration? (No = replace)",
			Default: true,
		}, &merge); err != nil {
			return err
		}

		if merge {
			// Merge: add new profiles, update existing ones
			profileMap := make(map[string]config.Profile)
			for _, p := range currentCfg.Profiles {
				profileMap[p.ID] = p
			}

			for _, p := range importedConfig.Profiles {
				profileMap[p.ID] = p
			}

			mergedProfiles := make([]config.Profile, 0, len(profileMap))
			for _, p := range profileMap {
				mergedProfiles = append(mergedProfiles, p)
			}

			currentCfg.Profiles = mergedProfiles
			currentCfg.FolderMappings = append(currentCfg.FolderMappings, importedConfig.FolderMappings...)
			currentCfg.RepoMappings = append(currentCfg.RepoMappings, importedConfig.RepoMappings...)
		} else {
			// Replace
			currentCfg = &importedConfig
		}
	} else {
		// No existing config, just use imported
		currentCfg = &importedConfig
	}

	// Save config
	if err := configMgr.Save(currentCfg); err != nil {
		return err
	}

	// Update SSH config
	// Trigger SSH config update by loading and saving
	cfg, _ := configMgr.Load()
	sshProfiles := make([]ssh.Profile, len(cfg.Profiles))
	for i, p := range cfg.Profiles {
		sshProfiles[i] = ssh.Profile{
			SSHHost:     p.SSHHost,
			IdentityFile: p.IdentityFile,
		}
	}

	sshMgr, err := ssh.NewManager()
	if err != nil {
		return err
	}

	if err := sshMgr.EnsureInclude(); err != nil {
		return fmt.Errorf("failed to ensure SSH include: %w", err)
	}

	if err := sshMgr.GenerateConfig(sshProfiles); err != nil {
		return fmt.Errorf("failed to generate SSH config: %w", err)
	}

	fmt.Printf("✓ Configuration imported from: %s\n", filePath)
	return nil
}

