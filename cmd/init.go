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
This will create a config.yaml file in the current directory or in the
$HOME/.ontap directory if the --global flag is set.

Examples:
  # Initialize a new configuration in the current directory
  ontap init

  # Initialize a new configuration in the home directory
  ontap init --global`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the global flag
			global, err := cmd.Flags().GetBool("global")
			if err != nil {
				return fmt.Errorf("failed to get global flag: %w", err)
			}

			// Get the force flag
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("failed to get force flag: %w", err)
			}

			// Get the config path
			configPath, err := getConfigPath(global)
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Check if the config file already exists
			if _, err := os.Stat(configPath); err == nil {
				if !force {
					return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
				}
				log.Warn("Overwriting existing config file", "path", configPath)
			}

			// Create the config file
			if err := config.CreateDefaultConfig(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}

			log.Info("Created config file", "path", configPath)
			fmt.Println("Edit the config file to add your API specs.")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)

	// Add flags
	initCmd.Flags().BoolP("global", "g", false, "Create the config file in the home directory")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing config file")
}

// getConfigPath returns the path to the config file
func getConfigPath(global bool) (string, error) {
	if global {
		// Get the home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Create the .ontap directory if it doesn't exist
		ontapDir := filepath.Join(home, ".ontap")
		if err := os.MkdirAll(ontapDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		return filepath.Join(ontapDir, "config.yaml"), nil
	}

	// Use the current directory
	return "config.yaml", nil
}
