package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/fynxlabs/ontap/internal/pkg/config"
	"github.com/spf13/cobra"
)

var (
	// initCmd represents the init command
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration",
		Long: `Initialize a new configuration file with default settings.
This will create a config.yaml file in the platform-specific user configuration directory:
		- macOS: ~/Library/Application Support/ontap/config.yaml
		- Linux: ~/.config/ontap/config.yaml (XDG style)
		- Windows: %APPDATA%\ontap\config.yaml

Examples:
		# Initialize a new configuration
		ontap init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the config path
			configPath, err := getConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Create the config file interactively
			if err := config.CreateInteractiveConfig(configPath); err != nil {
				return fmt.Errorf("failed to create config file interactively: %w", err)
			}

			log.Info("Created config file", "path", configPath)
			fmt.Println("Edit the config file to add your API specs.")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	// Always use the default config path
	loader := config.NewConfigLoader()
	defaultPath := loader.GetDefaultConfigPath()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(defaultPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return defaultPath, nil
}
