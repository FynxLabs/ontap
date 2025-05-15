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

// configFlag is the path to the config file
var configFlag string

var (
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
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Parse the flags first
	rootCmd.ParseFlags(os.Args)

	// Get the config flag value
	configFlag = rootCmd.Flag("config").Value.String()

	// Initialize logging
	utils.InitLogging(utils.GetLogLevel("info"))

	// Load the config early
	if err := initConfig(); err != nil {
		// If the config file doesn't exist, suggest running init
		if strings.Contains(err.Error(), "config file not found") || strings.Contains(err.Error(), "no such file or directory") {
			log.Info("No configuration found. Run 'ontap init' to create one.")
		} else {
			fmt.Println("Error initializing config:", err)
		}
	} else {
		// Generate dynamic commands after config is loaded
		if err := generateDynamicCommands(); err != nil {
			log.Error("Failed to generate dynamic commands", "error", err)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "Config file (default is platform-specific user config directory)")
	rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format (json, yaml, csv, text, table)")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Dry run (don't execute requests)")
	rootCmd.PersistentFlags().String("save", "", "Save response to file")
	rootCmd.PersistentFlags().String("extract", "", "Extract fields from response (comma-separated)")
	rootCmd.PersistentFlags().String("filter", "", "Filter response using JQ-like syntax")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("save", rootCmd.PersistentFlags().Lookup("save"))
	viper.BindPFlag("extract", rootCmd.PersistentFlags().Lookup("extract"))
	viper.BindPFlag("filter", rootCmd.PersistentFlags().Lookup("filter"))

	// Set up logging in PersistentPreRun
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Initialize logging with the log level from the flag
		logLevel, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return fmt.Errorf("failed to get log level: %w", err)
		}
		utils.InitLogging(utils.GetLogLevel(logLevel))
		return nil
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// Parse the config flag directly
	if configFlag != "" {
		// Use config file from the flag
		viper.SetConfigFile(configFlag)
		log.Debug("Using config file from flag", "path", configFlag)
	} else {
		// Use the default config path
		loader := config.NewConfigLoader()
		defaultPath := loader.GetDefaultConfigPath()

		// Create the directory if it doesn't exist
		dir := filepath.Dir(defaultPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		viper.SetConfigFile(defaultPath)
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
		return fmt.Errorf("config file not found")
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
