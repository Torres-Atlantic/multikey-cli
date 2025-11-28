package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Manage your MultiKey CLI license key",
	Long: `Manage your MultiKey CLI license key.

License keys are optional. All features work without a license key.
The license key enables supporter benefits and helps track supporters.

License Key Format:
  - Format: mk-xxxxxxxxxxxxxxxxx (20 characters total)
  - Example: mk-abc123xyz789def45
  - You'll receive your license key via email after purchase`,
}

var licenseSetCmd = &cobra.Command{
	Use:   "set [key]",
	Short: "Set your license key",
	Long:  "Set your MultiKey CLI license key. If no key is provided, you'll be prompted to enter it.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLicenseSet,
}

var licenseShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show license status",
	Long:  "Display your current license key status and supporter information.",
	RunE:  runLicenseShow,
}

func init() {
	licenseCmd.AddCommand(licenseSetCmd)
	licenseCmd.AddCommand(licenseShowCmd)
	rootCmd.AddCommand(licenseCmd)
}

func runLicenseSet(cmd *cobra.Command, args []string) error {
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	var licenseKey string

	if len(args) > 0 {
		licenseKey = args[0]
	} else {
		// Prompt for license key
		if err := survey.AskOne(&survey.Input{
			Message: "Enter your license key:",
			Help:    "License key format: mk-xxxxxxxxxxxxxxxxx (20 characters total)",
		}, &licenseKey, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	// Validate format
	if err := validateLicenseKey(licenseKey); err != nil {
		return err
	}

	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	// Normalize to lowercase for storage
	cfg.Meta.LicenseKey = strings.ToLower(strings.TrimSpace(licenseKey))

	// Save config
	if err := configMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save license key: %w", err)
	}

	fmt.Println("✓ License key saved")
	fmt.Println("✓ Supporter benefits activated")
	fmt.Println()
	fmt.Println("You now have access to:")
	fmt.Println("  • Signed & notarized macOS binaries")
	fmt.Println("  • Homebrew tap with prebuilt binaries")
	fmt.Println("  • Early access to new features")

	return nil
}

func runLicenseShow(cmd *cobra.Command, args []string) error {
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	if cfg.Meta.LicenseKey == "" {
		fmt.Println("License Status: No license key set")
		fmt.Println()
		fmt.Println("MultiKey CLI is free to use. All features work without a license key.")
		fmt.Println()
		fmt.Println("To support development and get supporter benefits:")
		fmt.Println("  • Visit: https://www.multikeycli.com")
		fmt.Println("  • Purchase a license")
		fmt.Println("  • Enter your license key with: multikey license set <key>")
		return nil
	}

	// Mask the key (show first 6 and last 4 characters)
	maskedKey := maskLicenseKey(cfg.Meta.LicenseKey)

	fmt.Println("License Status: Active Supporter")
	fmt.Printf("License Key: %s\n", maskedKey)
	fmt.Println()
	fmt.Println("Supporter Benefits:")
	fmt.Println("  ✓ Signed & notarized macOS binaries")
	fmt.Println("  ✓ Homebrew tap with prebuilt binaries")
	fmt.Println("  ✓ Early access to new features")
	fmt.Println("  ✓ Name in sponsor list (optional)")

	return nil
}

// validateLicenseKey validates the license key format
func validateLicenseKey(key string) error {
	key = strings.TrimSpace(key)

	// Check format: mk- followed by 17 alphanumeric characters
	pattern := `^mk-[A-Za-z0-9]{17}$`
	matched, err := regexp.MatchString(pattern, key)
	if err != nil {
		return fmt.Errorf("failed to validate license key: %w", err)
	}

	if !matched {
		return fmt.Errorf(`invalid license key format
Expected format: mk-xxxxxxxxxxxxxxxxx (20 characters total)
Example: mk-abc123xyz789def45

Please check your license key and try again.`)
	}

	return nil
}

// maskLicenseKey masks the license key for display
func maskLicenseKey(key string) string {
	if len(key) < 10 {
		return "***"
	}
	// Show first 6 and last 4 characters
	return key[:6] + "..." + key[len(key)-4:]
}

