package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

const sponsorURL = "https://multikeycli.com/support"

var sponsorCmd = &cobra.Command{
	Use:   "sponsor",
	Short: "Show sponsor information",
	Long:  "Display information about supporting MultiKey CLI development.",
	Run:   runSponsor,
}

func runSponsor(cmd *cobra.Command, args []string) {
	fmt.Println("MultiKey CLI - Support the Developer")
	fmt.Println("====================================")
	fmt.Println()
	fmt.Println("MultiKey CLI is free and open source.")
	fmt.Println("All features are available without any restrictions.")
	fmt.Println()
	fmt.Println("If you find MultiKey CLI useful, consider supporting development:")
	fmt.Printf("  %s\n", sponsorURL)
	fmt.Println()
	fmt.Println("Supporter benefits:")
	fmt.Println("  • Signed & notarized macOS binaries")
	fmt.Println("  • Homebrew tap with prebuilt binaries")
	fmt.Println("  • Early access to new features")
	fmt.Println("  • Name in sponsor list (optional)")
	fmt.Println()

	// Ask if user wants to open URL
	var open bool
	fmt.Print("Open support page in browser? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		open = true
	}

	if open {
		var err error
		switch runtime.GOOS {
		case "darwin":
			err = exec.Command("open", sponsorURL).Start()
		case "linux":
			err = exec.Command("xdg-open", sponsorURL).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", sponsorURL).Start()
		default:
			fmt.Printf("Please visit: %s\n", sponsorURL)
			return
		}

		if err != nil {
			fmt.Printf("Could not open browser. Please visit: %s\n", sponsorURL)
		}
	}
}

