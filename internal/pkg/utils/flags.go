package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddGlobalFlags adds global flags to a command
func AddGlobalFlags(cmd *cobra.Command) {
	// Add global flags
	cmd.PersistentFlags().StringP("config", "c", "", "Config file (default is platform-specific user config directory)")
	cmd.PersistentFlags().StringP("output", "o", "json", "Output format (json, yaml, csv, text, table)")
	cmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	cmd.PersistentFlags().Bool("dry-run", false, "Dry run (don't execute requests)")
	cmd.PersistentFlags().String("save", "", "Save response to file")
	cmd.PersistentFlags().String("extract", "", "Extract fields from response (comma-separated)")
	cmd.PersistentFlags().String("filter", "", "Filter response using JQ-like syntax")

	// Bind flags to viper
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		log.Warn("Failed to bind flag", "flag", "config", "error", err)
	}
	if err := viper.BindPFlag("output", cmd.PersistentFlags().Lookup("output")); err != nil {
		log.Warn("Failed to bind flag", "flag", "output", "error", err)
	}
	if err := viper.BindPFlag("log_level", cmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Warn("Failed to bind flag", "flag", "log_level", "error", err)
	}
	if err := viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")); err != nil {
		log.Warn("Failed to bind flag", "flag", "verbose", "error", err)
	}
	if err := viper.BindPFlag("dry_run", cmd.PersistentFlags().Lookup("dry-run")); err != nil {
		log.Warn("Failed to bind flag", "flag", "dry_run", "error", err)
	}
	if err := viper.BindPFlag("save", cmd.PersistentFlags().Lookup("save")); err != nil {
		log.Warn("Failed to bind flag", "flag", "save", "error", err)
	}
	if err := viper.BindPFlag("extract", cmd.PersistentFlags().Lookup("extract")); err != nil {
		log.Warn("Failed to bind flag", "flag", "extract", "error", err)
	}
	if err := viper.BindPFlag("filter", cmd.PersistentFlags().Lookup("filter")); err != nil {
		log.Warn("Failed to bind flag", "flag", "filter", "error", err)
	}

	// Bind environment variables
	viper.SetEnvPrefix("OTAP")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

// AddRequestFlags adds request flags to a command
func AddRequestFlags(cmd *cobra.Command) {
	// Add request flags
	cmd.Flags().StringP("data", "d", "", "Request body data (JSON string or @file)")
	cmd.Flags().StringArrayP("header", "H", nil, "Request header (key:value)")
	cmd.Flags().StringArrayP("query", "q", nil, "Query parameter (key=value)")
	cmd.Flags().StringArrayP("form", "F", nil, "Form data (key=value or key=@file)")
	cmd.Flags().StringP("auth", "a", "", "Authentication (username:password, Bearer token, or API key)")
	cmd.Flags().StringP("content-type", "t", "", "Content type")
}

// AddParameterFlags adds parameter flags to a command based on OpenAPI parameters
func AddParameterFlags(cmd *cobra.Command, parameters []Parameter) error {
	for _, param := range parameters {
		// Skip parameters that are already added
		if cmd.Flags().Lookup(param.Name) != nil {
			continue
		}

		// Add the flag based on the parameter type
		switch param.In {
		case "path":
			// Path parameters are handled by the command arguments
			continue
		case "query":
			addQueryParameterFlag(cmd, param)
		case "header":
			addHeaderParameterFlag(cmd, param)
		case "cookie":
			addCookieParameterFlag(cmd, param)
		default:
			return fmt.Errorf("unsupported parameter location: %s", param.In)
		}

		// Mark the flag as required if necessary
		if param.Required {
			if err := cmd.MarkFlagRequired(param.Name); err != nil {
				log.Warn("Failed to mark flag as required", "flag", param.Name, "error", err)
			}
		}
	}

	return nil
}

// Parameter represents an OpenAPI parameter
type Parameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Schema      *ParameterSchema
}

// ParameterSchema represents an OpenAPI parameter schema
type ParameterSchema struct {
	Type    string
	Format  string
	Default interface{}
	Enum    []interface{}
}

// addQueryParameterFlag adds a query parameter flag to a command
func addQueryParameterFlag(cmd *cobra.Command, param Parameter) {
	// Add the flag based on the parameter type
	if param.Schema == nil {
		// Default to string
		cmd.Flags().String(param.Name, "", param.Description)
		return
	}

	switch param.Schema.Type {
	case "string":
		defaultValue := ""
		if param.Schema.Default != nil {
			defaultValue = fmt.Sprintf("%v", param.Schema.Default)
		}
		cmd.Flags().String(param.Name, defaultValue, param.Description)
	case "integer", "number":
		defaultValue := 0
		if param.Schema.Default != nil {
			if v, ok := param.Schema.Default.(float64); ok {
				defaultValue = int(v)
			}
		}
		cmd.Flags().Int(param.Name, defaultValue, param.Description)
	case "boolean":
		defaultValue := false
		if param.Schema.Default != nil {
			if v, ok := param.Schema.Default.(bool); ok {
				defaultValue = v
			}
		}
		cmd.Flags().Bool(param.Name, defaultValue, param.Description)
	case "array":
		cmd.Flags().StringArray(param.Name, nil, param.Description)
	default:
		cmd.Flags().String(param.Name, "", param.Description)
	}
}

// addHeaderParameterFlag adds a header parameter flag to a command
func addHeaderParameterFlag(cmd *cobra.Command, param Parameter) {
	// Add the flag
	cmd.Flags().String(param.Name, "", param.Description)
}

// addCookieParameterFlag adds a cookie parameter flag to a command
func addCookieParameterFlag(cmd *cobra.Command, param Parameter) {
	// Add the flag
	cmd.Flags().String(param.Name, "", param.Description)
}

// GetFlagValue gets a flag value from a command
func GetFlagValue(flags *pflag.FlagSet, name string) (interface{}, error) {
	// Get the flag
	flag := flags.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag not found: %s", name)
	}

	// Get the flag value
	return flag.Value.String(), nil
}

// GetStringFlagValue gets a string flag value from a command
func GetStringFlagValue(flags *pflag.FlagSet, name string) (string, error) {
	// Get the flag value
	value, err := flags.GetString(name)
	if err != nil {
		return "", fmt.Errorf("failed to get string flag %s: %w", name, err)
	}

	return value, nil
}

// GetBoolFlagValue gets a boolean flag value from a command
func GetBoolFlagValue(flags *pflag.FlagSet, name string) (bool, error) {
	// Get the flag value
	value, err := flags.GetBool(name)
	if err != nil {
		return false, fmt.Errorf("failed to get bool flag %s: %w", name, err)
	}

	return value, nil
}

// GetIntFlagValue gets an integer flag value from a command
func GetIntFlagValue(flags *pflag.FlagSet, name string) (int, error) {
	// Get the flag value
	value, err := flags.GetInt(name)
	if err != nil {
		return 0, fmt.Errorf("failed to get int flag %s: %w", name, err)
	}

	return value, nil
}

// GetStringArrayFlagValue gets a string array flag value from a command
func GetStringArrayFlagValue(flags *pflag.FlagSet, name string) ([]string, error) {
	// Get the flag value
	value, err := flags.GetStringArray(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get string array flag %s: %w", name, err)
	}

	return value, nil
}

// ParseDataFlag parses the data flag value
func ParseDataFlag(value string) (interface{}, error) {
	// Check if the value is a file path
	if strings.HasPrefix(value, "@") {
		// Read the file
		filePath := value[1:]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read data file: %w", err)
		}

		// Parse the data as JSON
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("failed to parse data file as JSON: %w", err)
		}

		return jsonData, nil
	}

	// Parse the value as JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(value), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse data as JSON: %w", err)
	}

	return jsonData, nil
}

// ParseHeaderFlags parses the header flag values
func ParseHeaderFlags(values []string) (map[string]string, error) {
	// Create a map to store the headers
	headers := make(map[string]string)

	// Parse the header values
	for _, value := range values {
		// Split the value into key and value
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", value)
		}

		// Add the header
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return headers, nil
}

// ParseQueryFlags parses the query flag values
func ParseQueryFlags(values []string) (map[string]string, error) {
	// Create a map to store the query parameters
	params := make(map[string]string)

	// Parse the query values
	for _, value := range values {
		// Split the value into key and value
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid query parameter format: %s", value)
		}

		// Add the query parameter
		params[parts[0]] = parts[1]
	}

	return params, nil
}

// ParseFormFlags parses the form flag values
func ParseFormFlags(values []string) (map[string]string, map[string]string, error) {
	// Create maps to store the form data and files
	formData := make(map[string]string)
	formFiles := make(map[string]string)

	// Parse the form values
	for _, value := range values {
		// Split the value into key and value
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid form data format: %s", value)
		}

		// Check if the value is a file path
		if strings.HasPrefix(parts[1], "@") {
			// Add the form file
			formFiles[parts[0]] = parts[1]
		} else {
			// Add the form data
			formData[parts[0]] = parts[1]
		}
	}

	return formData, formFiles, nil
}
