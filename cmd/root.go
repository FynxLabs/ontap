package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/fynxlabs/ontap/internal/pkg/config"
	"github.com/fynxlabs/ontap/internal/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is the version of the CLI
var Version = "dev"

// BuildTime is the time the CLI was built
var BuildTime = "unknown"

var (
	// cfgFile is the path to the config file
	cfgFile string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "ontap",
		Short: "OnTap - CLI generator from OpenAPI specs",
		Long: `OnTap is a CLI generator that creates command-line interfaces from OpenAPI specifications.
It enables instant CLI interaction with any API that has an OpenAPI spec without requiring custom code or SDK generation.

Examples:
  # Initialize a new configuration
  ontap init

  # Use the CLI with your API
  ontap your-api list-users
  ontap your-api create-user --data=@user.json`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logging
			logLevel, err := cmd.Flags().GetString("log-level")
			if err != nil {
				return fmt.Errorf("failed to get log level: %w", err)
			}
			utils.InitLogging(utils.GetLogLevel(logLevel))

			// Load config if not init command
			if cmd.Name() != "init" && cmd.Name() != "help" && cmd.Name() != "version" {
				if err := initConfig(); err != nil {
					return err
				}
			}

			return nil
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags
	utils.AddGlobalFlags(rootCmd)

	// Set up Viper for config file
	cobra.OnInitialize(initConfigWrapper)

	// Get the config flag value
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag != nil {
		cfgFile = configFlag.Value.String()
	}

	// Add dynamic commands
	if err := generateDynamicCommands(); err != nil {
		log.Error("Failed to generate dynamic commands", "error", err)
	}
}

// initConfigWrapper is a wrapper for initConfig that doesn't return an error
func initConfigWrapper() {
	if err := initConfig(); err != nil {
		fmt.Println("Error initializing config:", err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Search config in home directory with name ".ontap" (without extension)
		viper.AddConfigPath(filepath.Join(home, ".ontap"))
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("ONTAP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		log.Debug("No config file found")
	} else {
		log.Debug("Using config file", "path", viper.ConfigFileUsed())
	}

	return nil
}

// loadConfig loads the configuration
func loadConfig() (*config.Config, error) {
	// Create a config loader
	loader := config.NewConfigLoader()

	// Load the config
	cfg, err := loader.LoadConfig(viper.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
