package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Torres-Atlantic/multikey-cli/internal/config"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export configuration",
	Long:  "Print the current configuration as JSON to stdout.",
	RunE:  runExport,
}

var exportOutput string

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
}

func runExport(cmd *cobra.Command, args []string) error {
	configMgr, err := config.NewManager()
	if err != nil {
		return err
	}

	cfg, err := configMgr.Load()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Configuration exported to: %s\n", exportOutput)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

