package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultConfigFileName is the default name for the config file
	DefaultConfigFileName = "config.yaml"

	// DefaultConfigDir is the default directory for the config file
	DefaultConfigDir = ".ontap"
)

// ConfigLoader is the interface for loading and saving configurations
type ConfigLoader interface {
	// LoadConfig loads the configuration from the specified path
	LoadConfig(path string) (*Config, error)

	// SaveConfig saves the configuration to the specified path
	SaveConfig(config *Config, path string) error

	// GetAPIConfig returns the configuration for the specified API
	GetAPIConfig(name string) (*APIConfig, error)

	// GetDefaultConfigPath returns the default path for the config file
	GetDefaultConfigPath() string
}

// ViperConfigLoader implements ConfigLoader using Viper
type ViperConfigLoader struct {
	viper  *viper.Viper
	config *Config
}

// NewConfigLoader creates a new ViperConfigLoader
func NewConfigLoader() *ViperConfigLoader {
	return &ViperConfigLoader{
		viper: viper.New(),
	}
}

// LoadConfig loads the configuration from the specified path
func (l *ViperConfigLoader) LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = l.GetDefaultConfigPath()
	}

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	l.viper.SetConfigFile(path)

	// Enable environment variable substitution
	l.viper.AutomaticEnv()
	l.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := l.viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := l.viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	l.config = config
	return config, nil
}

// SaveConfig saves the configuration to the specified path
func (l *ViperConfigLoader) SaveConfig(config *Config, path string) error {
	if path == "" {
		path = l.GetDefaultConfigPath()
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the config to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Info("Config saved", "path", path)
	return nil
}

// GetAPIConfig returns the configuration for the specified API
func (l *ViperConfigLoader) GetAPIConfig(name string) (*APIConfig, error) {
	if l.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	apiConfig, ok := l.config.APIs[name]
	if !ok {
		return nil, fmt.Errorf("API not found: %s", name)
	}

	return &apiConfig, nil
}

// GetDefaultConfigPath returns the default path for the config file
func (l *ViperConfigLoader) GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warn("Failed to get user home directory", "error", err)
		return DefaultConfigFileName
	}

	configDir := filepath.Join(homeDir, "config")
	return filepath.Join(configDir, DefaultConfigDir, DefaultConfigFileName)
}
