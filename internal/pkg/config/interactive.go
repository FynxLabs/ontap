package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

// CreateInteractiveConfig creates a configuration file interactively
func CreateInteractiveConfig(path string) error {
	var config *Config
	var existingAPIs []string
	var configExists bool

	// Create the directory structure if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if the config file already exists
	if _, err := os.Stat(path); err == nil {
		configExists = true
		// Load the existing config
		loader := NewConfigLoader()
		config, err = loader.LoadConfig(path)
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}

		// Get existing API names
		existingAPIs = make([]string, 0, len(config.APIs))
		for name := range config.APIs {
			existingAPIs = append(existingAPIs, name)
		}
	} else {
		// Create a new config
		config = &Config{
			APIs: make(map[string]APIConfig),
		}
	}

	// If config exists, ask if the user wants to edit it
	if configExists {
		var editExisting bool
		var createNew bool

		// Ask if the user wants to edit the existing config
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Config already exists at %s with %d APIs: %s", path, len(existingAPIs), strings.Join(existingAPIs, ", "))).
			Description("Would you like to edit it?").
			Affirmative("Yes, edit existing config").
			Negative("No").
			Value(&editExisting).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for editing existing config: %w", err)
		}

		if editExisting {
			// Edit existing config
			if err := editExistingConfig(config, existingAPIs); err != nil {
				return err
			}
		} else {
			// Ask if they want to create a new one
			err := huh.NewConfirm().
				Title("Create a new config instead?").
				Description("This will overwrite the existing one.").
				Affirmative("Yes, create new config").
				Negative("No, exit without changes").
				Value(&createNew).
				Run()

			if err != nil {
				return fmt.Errorf("failed to prompt for creating new config: %w", err)
			}

			if !createNew {
				// Exit without changes
				return nil
			}

			// Create a new config
			config = &Config{
				APIs: make(map[string]APIConfig),
			}
		}
	}

	// If we're creating a new config or the user chose to create a new one
	if !configExists || (configExists && len(config.APIs) == 0) {
		// Ask how many APIs to configure
		var numAPIsStr string
		err := huh.NewInput().
			Title("How many APIs would you like to configure?").
			Description("Enter a number").
			Placeholder("1").
			Validate(func(s string) error {
				if s == "" {
					return nil // Default to 1
				}
				n, err := strconv.Atoi(s)
				if err != nil || n < 1 {
					return fmt.Errorf("please enter a positive number")
				}
				return nil
			}).
			Value(&numAPIsStr).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for number of APIs: %w", err)
		}

		// Parse the number of APIs
		numAPIs := 1 // Default to 1
		if numAPIsStr != "" {
			var err error
			numAPIs, err = strconv.Atoi(numAPIsStr)
			if err != nil || numAPIs < 1 {
				numAPIs = 1
			}
		}

		// Configure each API
		for i := 0; i < numAPIs; i++ {
			if err := addNewAPI(config, i+1, numAPIs); err != nil {
				return err
			}
		}
	}

	// Save the config
	loader := NewConfigLoader()
	if err := loader.SaveConfig(config, path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Info("Configuration saved successfully", "path", path)
	return nil
}

// editExistingConfig allows the user to edit an existing configuration
func editExistingConfig(config *Config, existingAPIs []string) error {
	// Ask what the user wants to do
	var action string
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Add a new API", "add"),
			huh.NewOption("Edit an existing API", "edit"),
			huh.NewOption("Remove an API", "remove"),
		).
		Value(&action).
		Run()

	if err != nil {
		return fmt.Errorf("failed to prompt for action: %w", err)
	}

	switch action {
	case "add":
		// Add a new API
		var numAPIsStr string
		err := huh.NewInput().
			Title("How many APIs would you like to add?").
			Description("Enter a number").
			Placeholder("1").
			Validate(func(s string) error {
				if s == "" {
					return nil // Default to 1
				}
				n, err := strconv.Atoi(s)
				if err != nil || n < 1 {
					return fmt.Errorf("please enter a positive number")
				}
				return nil
			}).
			Value(&numAPIsStr).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for number of APIs: %w", err)
		}

		// Parse the number of APIs
		numAPIs := 1 // Default to 1
		if numAPIsStr != "" {
			var err error
			numAPIs, err = strconv.Atoi(numAPIsStr)
			if err != nil || numAPIs < 1 {
				numAPIs = 1
			}
		}

		// Add each API
		for i := 0; i < numAPIs; i++ {
			if err := addNewAPI(config, i+1, numAPIs); err != nil {
				return err
			}
		}

	case "edit":
		// Edit an existing API
		var apiToEdit string
		err := huh.NewSelect[string]().
			Title("Which API would you like to edit?").
			Options(huh.NewOptions(existingAPIs...)...).
			Value(&apiToEdit).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for API to edit: %w", err)
		}

		// Edit the selected API
		if err := editAPI(config, apiToEdit); err != nil {
			return err
		}

	case "remove":
		// Remove an API
		var apiToRemove string
		err := huh.NewSelect[string]().
			Title("Which API would you like to remove?").
			Options(huh.NewOptions(existingAPIs...)...).
			Value(&apiToRemove).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for API to remove: %w", err)
		}

		// Confirm removal
		var confirmRemove bool
		err = huh.NewConfirm().
			Title(fmt.Sprintf("Are you sure you want to remove the API '%s'?", apiToRemove)).
			Affirmative("Yes, remove it").
			Negative("No, keep it").
			Value(&confirmRemove).
			Run()

		if err != nil {
			return fmt.Errorf("failed to confirm API removal: %w", err)
		}

		if confirmRemove {
			// Remove the API
			delete(config.APIs, apiToRemove)
			log.Info("API removed", "name", apiToRemove)
		}
	}

	return nil
}

// addNewAPI adds a new API to the configuration
func addNewAPI(config *Config, index, total int) error {
	var apiName string
	var apiSpec string
	var baseURL string
	var cacheTTL string
	var auth string
	var output string
	var addHeaders bool

	// Get API details
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("API Name (%d/%d)", index, total)).
				Description("A unique identifier for this API").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API name cannot be empty")
					}
					if _, exists := config.APIs[s]; exists {
						return fmt.Errorf("API '%s' already exists", s)
					}
					return nil
				}).
				Value(&apiName),

			huh.NewInput().
				Title("API Spec Location").
				Description("URL or file path to the OpenAPI specification").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API spec location cannot be empty")
					}
					return nil
				}).
				Value(&apiSpec),

			huh.NewInput().
				Title("Base URL").
				Description("Base URL for API requests").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("base URL cannot be empty")
					}
					return nil
				}).
				Value(&baseURL),

			huh.NewInput().
				Title("Cache TTL").
				Description("Time-to-live for caching the OpenAPI spec").
				Placeholder("1h").
				Validate(func(s string) error {
					if s == "" {
						return nil // Default to 1h
					}
					_, err := time.ParseDuration(s)
					if err != nil {
						return fmt.Errorf("invalid duration format: %w", err)
					}
					return nil
				}).
				Value(&cacheTTL),

			huh.NewInput().
				Title("Authentication").
				Description("Authentication credentials (username:password or token)").
				Value(&auth),

			huh.NewSelect[string]().
				Title("Default Output Format").
				Options(
					huh.NewOption("JSON", "json"),
					huh.NewOption("YAML", "yaml"),
					huh.NewOption("CSV", "csv"),
					huh.NewOption("Text", "text"),
					huh.NewOption("Table", "table"),
				).
				Value(&output),

			huh.NewConfirm().
				Title("Would you like to add custom headers?").
				Value(&addHeaders),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to get API details: %w", err)
	}

	// Set default values
	if cacheTTL == "" {
		cacheTTL = "1h"
	}
	if output == "" {
		output = "json"
	}

	// Parse cache TTL
	duration, err := time.ParseDuration(cacheTTL)
	if err != nil {
		return fmt.Errorf("failed to parse cache TTL: %w", err)
	}

	// Create the API config
	apiConfig := APIConfig{
		APISpec:       apiSpec,
		URL:           baseURL,
		Auth:          auth,
		CacheTTL:      Duration{Duration: duration},
		DefaultOutput: output,
		Headers:       make(map[string]string),
	}

	// Add custom headers if requested
	if addHeaders {
		if err := addCustomHeaders(&apiConfig); err != nil {
			return err
		}
	}

	// Add the API to the config
	config.APIs[apiName] = apiConfig
	log.Info("API added", "name", apiName)

	return nil
}

// editAPI edits an existing API in the configuration
func editAPI(config *Config, apiName string) error {
	// Get the existing API config
	apiConfig, exists := config.APIs[apiName]
	if !exists {
		return fmt.Errorf("API '%s' not found", apiName)
	}

	var newAPIName string
	var apiSpec string
	var baseURL string
	var cacheTTL string
	var auth string
	var output string
	var editHeaders bool

	// Pre-fill with existing values
	newAPIName = apiName
	apiSpec = apiConfig.APISpec
	baseURL = apiConfig.URL
	cacheTTL = apiConfig.CacheTTL.String()
	auth = apiConfig.Auth
	output = apiConfig.DefaultOutput

	// Get updated API details
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("API Name").
				Description("A unique identifier for this API").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API name cannot be empty")
					}
					if s != apiName {
						if _, exists := config.APIs[s]; exists {
							return fmt.Errorf("API '%s' already exists", s)
						}
					}
					return nil
				}).
				Value(&newAPIName),

			huh.NewInput().
				Title("API Spec Location").
				Description("URL or file path to the OpenAPI specification").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("API spec location cannot be empty")
					}
					return nil
				}).
				Value(&apiSpec),

			huh.NewInput().
				Title("Base URL").
				Description("Base URL for API requests").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("base URL cannot be empty")
					}
					return nil
				}).
				Value(&baseURL),

			huh.NewInput().
				Title("Cache TTL").
				Description("Time-to-live for caching the OpenAPI spec").
				Validate(func(s string) error {
					if s == "" {
						return nil // Keep existing value
					}
					_, err := time.ParseDuration(s)
					if err != nil {
						return fmt.Errorf("invalid duration format: %w", err)
					}
					return nil
				}).
				Value(&cacheTTL),

			huh.NewInput().
				Title("Authentication").
				Description("Authentication credentials (username:password or token)").
				Value(&auth),

			huh.NewSelect[string]().
				Title("Default Output Format").
				Options(
					huh.NewOption("JSON", "json"),
					huh.NewOption("YAML", "yaml"),
					huh.NewOption("CSV", "csv"),
					huh.NewOption("Text", "text"),
					huh.NewOption("Table", "table"),
				).
				Value(&output),

			huh.NewConfirm().
				Title("Would you like to edit custom headers?").
				Value(&editHeaders),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("failed to get updated API details: %w", err)
	}

	// Parse cache TTL
	duration, err := time.ParseDuration(cacheTTL)
	if err != nil {
		return fmt.Errorf("failed to parse cache TTL: %w", err)
	}

	// Update the API config
	apiConfig.APISpec = apiSpec
	apiConfig.URL = baseURL
	apiConfig.Auth = auth
	apiConfig.CacheTTL = Duration{Duration: duration}
	apiConfig.DefaultOutput = output

	// Edit custom headers if requested
	if editHeaders {
		if err := editCustomHeaders(&apiConfig); err != nil {
			return err
		}
	}

	// Update the API in the config
	if newAPIName != apiName {
		// Remove the old API
		delete(config.APIs, apiName)
		// Add the new API
		config.APIs[newAPIName] = apiConfig
		log.Info("API renamed", "old", apiName, "new", newAPIName)
	} else {
		// Update the existing API
		config.APIs[apiName] = apiConfig
		log.Info("API updated", "name", apiName)
	}

	return nil
}

// addCustomHeaders adds custom headers to an API configuration
func addCustomHeaders(apiConfig *APIConfig) error {
	var numHeadersStr string
	err := huh.NewInput().
		Title("How many headers would you like to add?").
		Description("Enter a number").
		Placeholder("1").
		Validate(func(s string) error {
			if s == "" {
				return nil // Default to 1
			}
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return fmt.Errorf("please enter a positive number")
			}
			return nil
		}).
		Value(&numHeadersStr).
		Run()

	if err != nil {
		return fmt.Errorf("failed to prompt for number of headers: %w", err)
	}

	// Parse the number of headers
	numHeaders := 1 // Default to 1
	if numHeadersStr != "" {
		var err error
		numHeaders, err = strconv.Atoi(numHeadersStr)
		if err != nil || numHeaders < 1 {
			numHeaders = 1
		}
	}

	// Add each header
	for i := 0; i < numHeaders; i++ {
		var headerName string
		var headerValue string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Header Name (%d/%d)", i+1, numHeaders)).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("header name cannot be empty")
						}
						return nil
					}).
					Value(&headerName),

				huh.NewInput().
					Title(fmt.Sprintf("Header Value (%d/%d)", i+1, numHeaders)).
					Value(&headerValue),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to get header details: %w", err)
		}

		// Add the header
		apiConfig.Headers[headerName] = headerValue
		log.Info("Header added", "name", headerName)
	}

	return nil
}

// editCustomHeaders edits custom headers in an API configuration
func editCustomHeaders(apiConfig *APIConfig) error {
	// Get existing header names
	headerNames := make([]string, 0, len(apiConfig.Headers))
	for name := range apiConfig.Headers {
		headerNames = append(headerNames, name)
	}

	// Ask what the user wants to do
	var action string
	var options []huh.Option[string]

	options = append(options, huh.NewOption("Add a new header", "add"))
	if len(headerNames) > 0 {
		options = append(options, huh.NewOption("Edit an existing header", "edit"))
		options = append(options, huh.NewOption("Remove a header", "remove"))
	}

	err := huh.NewSelect[string]().
		Title("What would you like to do with headers?").
		Options(options...).
		Value(&action).
		Run()

	if err != nil {
		return fmt.Errorf("failed to prompt for header action: %w", err)
	}

	switch action {
	case "add":
		// Add a new header
		var headerName string
		var headerValue string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Header Name").
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("header name cannot be empty")
						}
						return nil
					}).
					Value(&headerName),

				huh.NewInput().
					Title("Header Value").
					Value(&headerValue),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to get header details: %w", err)
		}

		// Add the header
		apiConfig.Headers[headerName] = headerValue
		log.Info("Header added", "name", headerName)

	case "edit":
		// Edit an existing header
		var headerToEdit string
		err := huh.NewSelect[string]().
			Title("Which header would you like to edit?").
			Options(huh.NewOptions(headerNames...)...).
			Value(&headerToEdit).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for header to edit: %w", err)
		}

		// Get the existing header value
		existingValue := apiConfig.Headers[headerToEdit]

		// Get the new header value
		var newHeaderValue string
		err = huh.NewInput().
			Title(fmt.Sprintf("Header Value for '%s'", headerToEdit)).
			Placeholder(existingValue).
			Value(&newHeaderValue).
			Run()

		if err != nil {
			return fmt.Errorf("failed to get header value: %w", err)
		}

		// Update the header
		apiConfig.Headers[headerToEdit] = newHeaderValue
		log.Info("Header updated", "name", headerToEdit)

	case "remove":
		// Remove a header
		var headerToRemove string
		err := huh.NewSelect[string]().
			Title("Which header would you like to remove?").
			Options(huh.NewOptions(headerNames...)...).
			Value(&headerToRemove).
			Run()

		if err != nil {
			return fmt.Errorf("failed to prompt for header to remove: %w", err)
		}

		// Confirm removal
		var confirmRemove bool
		err = huh.NewConfirm().
			Title(fmt.Sprintf("Are you sure you want to remove the header '%s'?", headerToRemove)).
			Affirmative("Yes, remove it").
			Negative("No, keep it").
			Value(&confirmRemove).
			Run()

		if err != nil {
			return fmt.Errorf("failed to confirm header removal: %w", err)
		}

		if confirmRemove {
			// Remove the header
			delete(apiConfig.Headers, headerToRemove)
			log.Info("Header removed", "name", headerToRemove)
		}
	}

	return nil
}
