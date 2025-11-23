package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	configPath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "multikey",
	Short: "MultiKey CLI - Manage multiple GitHub SSH identities",
	Long: `MultiKey CLI is a developer tool that manages multiple GitHub SSH identities
and applies the correct identity based on folder or repo location.

It simplifies working with multiple GitHub accounts (personal, work, clients)
by providing profile-based SSH routing tied to directory paths and repositories.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Check for first run
	configMgr, err := config.NewManager()
	if err == nil && !configMgr.Exists() {
		fmt.Println("Welcome to MultiKey CLI!")
		fmt.Println()
		fmt.Println("It looks like this is your first time running MultiKey CLI.")
		fmt.Println()
		var runSetup bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Run setup now?",
			Default: true,
		}, &runSetup); err == nil && runSetup {
			// Run setup
			setupCmd.RunE(nil, nil)
			return nil
		}
		fmt.Println()
	}

	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (default: ~/.config/multikey/config.json)")

	// Add subcommands
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(mapCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(sponsorCmd)
	rootCmd.AddCommand(versionCmd)
}

// debugLog prints a debug message if debug mode is enabled
func debugLog(format string, args ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

