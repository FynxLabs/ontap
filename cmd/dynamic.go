package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fynxlabs/ontap/internal/pkg/cache"
	"github.com/fynxlabs/ontap/internal/pkg/config"
	"github.com/fynxlabs/ontap/internal/pkg/http"
	"github.com/fynxlabs/ontap/internal/pkg/openapi"
	"github.com/fynxlabs/ontap/internal/pkg/output"
	"github.com/fynxlabs/ontap/internal/pkg/utils"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateDynamicCommands generates dynamic commands for all APIs
func generateDynamicCommands() error {
	// Get the config file path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		log.Debug("No config file used")
		return nil
	}

	log.Debug("Generating dynamic commands from config", "path", configFile)

	// Create a config loader
	loader := config.NewConfigLoader()

	// Load the config
	cfg, err := loader.LoadConfig(configFile)
	if err != nil {
		// Check if the error is because the config file doesn't exist
		if strings.Contains(err.Error(), "config file not found") {
			log.Info("No configuration found. Run 'ontap init' to create one.")
			return nil
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no APIs are configured, return
	if len(cfg.APIs) == 0 {
		log.Info("No APIs configured. Run 'ontap init' to configure APIs.")
		return nil
	}

	// Create a cache manager with proper error handling
	cacheManager, err := cache.NewLibOpenAPICacheManager("")
	if err != nil {
		log.Error("Failed to create cache manager", "error", err)
		return fmt.Errorf("failed to create cache manager: %w", err)
	}

	// Add a command for each API
	for name, apiConfig := range cfg.APIs {
		log.Debug("Adding command for API", "name", name, "url", apiConfig.URL)

		// Create a new command
		apiCmd := &cobra.Command{
			Use:   name,
			Short: fmt.Sprintf("Commands for %s API", name),
			Long:  fmt.Sprintf("Commands for %s API at %s", name, apiConfig.URL),
		}

		// Add the command to the root command
		rootCmd.AddCommand(apiCmd)

		// Add dynamic commands for the API
		if err := generateDynamicAPICommands(apiCmd, name, apiConfig, cacheManager); err != nil {
			log.Error("Failed to generate commands for API", "api", name, "error", err)
			continue
		}
	}

	return nil
}

// generateDynamicAPICommands generates dynamic commands for an API
func generateDynamicAPICommands(cmd *cobra.Command, apiName string, apiConfig config.APIConfig, cacheManager *cache.LibOpenAPICacheManager) error {
	// Get the cache TTL
	ttl := apiConfig.CacheTTL.Duration
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	// Load the OpenAPI spec with proper error handling
	spec, err := loadOpenAPISpec(cacheManager, apiConfig.APISpec, ttl)
	if err != nil {
		return fmt.Errorf("failed to load spec for API %s: %w", apiName, err)
	}

	// Check if the spec is nil
	if spec == nil {
		return fmt.Errorf("failed to load spec for API %s: spec is nil", apiName)
	}

	// Create a parser
	parser := openapi.NewLibOpenAPISpecParser()

	// Get the endpoints
	endpoints, err := parser.GetEndpoints(spec)
	if err != nil {
		return fmt.Errorf("failed to get endpoints: %w", err)
	}

	// Group endpoints by tag
	taggedEndpoints := make(map[string][]openapi.Endpoint)
	for _, endpoint := range endpoints {
		// Skip deprecated endpoints
		if endpoint.Deprecated {
			continue
		}

		// Use the first tag as the group, or "default" if no tags
		tag := "default"
		if len(endpoint.Tags) > 0 {
			tag = endpoint.Tags[0]
		}

		taggedEndpoints[tag] = append(taggedEndpoints[tag], endpoint)
	}

	// Add a command for each tag
	for tag, endpoints := range taggedEndpoints {
		// Create a new command
		tagCmd := &cobra.Command{
			Use:   tag,
			Short: fmt.Sprintf("Commands for %s", tag),
		}

		// Add the command to the API command
		cmd.AddCommand(tagCmd)

		// Add a command for each endpoint
		for _, endpoint := range endpoints {
			// Create a new command
			endpointCmd := createEndpointCommand(endpoint, apiConfig)

			// Add the command to the tag command
			tagCmd.AddCommand(endpointCmd)
		}
	}

	return nil
}

// loadOpenAPISpec loads an OpenAPI spec with proper error handling
func loadOpenAPISpec(cacheManager *cache.LibOpenAPICacheManager, specPath string, ttl time.Duration) (*v3.Document, error) {
	if cacheManager == nil {
		return nil, fmt.Errorf("cache manager is nil")
	}

	// Check if we need to clear the cache
	if os.Getenv("ONTAP_CLEAR_CACHE") == "true" {
		clearCache(cacheManager)
	}

	// Try to get the spec from the cache manager
	spec, err := safeGetSpec(cacheManager, specPath, ttl)
	if err != nil || spec == nil {
		log.Error("Failed to get spec from cache", "path", specPath, "error", err)

		// Try to load the spec directly
		parser := openapi.NewLibOpenAPISpecParser()
		spec, err = parser.ParseSpec(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec: %w", err)
		}

		// Check if the spec is nil
		if spec == nil {
			return nil, fmt.Errorf("failed to parse spec: spec is nil")
		}

		// Try to cache the spec
		// Create a cache key
		cacheKey := generateCacheKey(specPath)
		if cacheErr := cacheManager.Store.Set(cacheKey, spec, ttl); cacheErr != nil {
			log.Warn("Failed to cache spec", "error", cacheErr)
		}
	}

	return spec, nil
}

// safeGetSpec safely gets a spec from the cache manager
func safeGetSpec(cacheManager *cache.LibOpenAPICacheManager, specPath string, ttl time.Duration) (*v3.Document, error) {
	var spec *v3.Document
	var err error

	// Use a defer to catch panics
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic in GetSpec", "error", r)
			// Clear the cache
			clearCache(cacheManager)
			spec = nil
			err = fmt.Errorf("panic in GetSpec: %v", r)
		}
	}()

	// Try to get the spec
	spec, err = cacheManager.GetSpec(specPath, ttl)
	return spec, err
}

// clearCache clears the cache
func clearCache(cacheManager *cache.LibOpenAPICacheManager) {
	if err := cacheManager.Store.Clear(); err != nil {
		log.Error("Failed to clear cache", "error", err)
	} else {
		log.Info("Cache cleared")
	}
}

// generateCacheKey generates a cache key for a spec path
func generateCacheKey(specPath string) string {
	// Use the spec path as the key
	return specPath
}

// createEndpointCommand creates a command for an endpoint
func createEndpointCommand(endpoint openapi.Endpoint, apiConfig config.APIConfig) *cobra.Command {
	// Create a new command
	cmd := &cobra.Command{
		Use:   getCommandUse(endpoint),
		Short: endpoint.Summary,
		Long:  endpoint.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeEndpoint(cmd, args, endpoint, apiConfig)
		},
	}

	// Add request flags
	utils.AddRequestFlags(cmd)

	// Add parameter flags
	if err := utils.AddParameterFlags(cmd, convertParameters(endpoint.Parameters)); err != nil {
		log.Error("Failed to add parameter flags", "endpoint", endpoint.OperationID, "error", err)
	}

	return cmd
}

// getCommandUse returns the use string for a command
func getCommandUse(endpoint openapi.Endpoint) string {
	// Use the operation ID as the command name
	name := endpoint.OperationID
	if name == "" {
		// Use the method and path as the command name
		name = strings.ToLower(endpoint.Method) + "-" + strings.ReplaceAll(endpoint.Path, "/", "-")
		name = strings.TrimPrefix(name, "-")
	}

	// Add path parameters to the use string
	var pathParams []string
	for _, param := range endpoint.Parameters {
		if param.In == "path" {
			pathParams = append(pathParams, param.Name)
		}
	}

	if len(pathParams) > 0 {
		return fmt.Sprintf("%s [%s]", name, strings.Join(pathParams, " "))
	}

	return name
}

// executeEndpoint executes an endpoint
func executeEndpoint(cmd *cobra.Command, args []string, endpoint openapi.Endpoint, apiConfig config.APIConfig) error {
	// Get the output format
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output format: %w", err)
	}

	// Get the save path
	savePath, err := cmd.Flags().GetString("save")
	if err != nil {
		return fmt.Errorf("failed to get save path: %w", err)
	}

	// Get the extract fields
	extractStr, err := cmd.Flags().GetString("extract")
	if err != nil {
		return fmt.Errorf("failed to get extract fields: %w", err)
	}
	var extractFields []string
	if extractStr != "" {
		extractFields = strings.Split(extractStr, ",")
	}

	// Get the filter
	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return fmt.Errorf("failed to get filter: %w", err)
	}

	// Get the verbose flag
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Get the dry run flag
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry run flag: %w", err)
	}

	// Get the data flag
	dataStr, err := cmd.Flags().GetString("data")
	if err != nil {
		return fmt.Errorf("failed to get data flag: %w", err)
	}
	var data interface{}
	if dataStr != "" {
		data, err = utils.ParseDataFlag(dataStr)
		if err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}
	}

	// Get the header flags
	headerStrs, err := cmd.Flags().GetStringArray("header")
	if err != nil {
		return fmt.Errorf("failed to get header flags: %w", err)
	}
	headers, err := utils.ParseHeaderFlags(headerStrs)
	if err != nil {
		return fmt.Errorf("failed to parse headers: %w", err)
	}

	// Get the query flags
	queryStrs, err := cmd.Flags().GetStringArray("query")
	if err != nil {
		return fmt.Errorf("failed to get query flags: %w", err)
	}
	queryParams, err := utils.ParseQueryFlags(queryStrs)
	if err != nil {
		return fmt.Errorf("failed to parse query parameters: %w", err)
	}

	// Get the form flags
	formStrs, err := cmd.Flags().GetStringArray("form")
	if err != nil {
		return fmt.Errorf("failed to get form flags: %w", err)
	}
	formData, formFiles, err := utils.ParseFormFlags(formStrs)
	if err != nil {
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Get the auth flag
	auth, err := cmd.Flags().GetString("auth")
	if err != nil {
		return fmt.Errorf("failed to get auth flag: %w", err)
	}
	if auth == "" {
		auth = apiConfig.Auth
	}

	// Get the content type flag
	contentType, err := cmd.Flags().GetString("content-type")
	if err != nil {
		return fmt.Errorf("failed to get content type flag: %w", err)
	}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}

	// Create an HTTP client
	client := http.NewClient(apiConfig.URL, auth)
	client.Verbose = verbose

	// Create a request
	req := &http.Request{
		Method:      endpoint.Method,
		Path:        endpoint.Path,
		QueryParams: url.Values{},
		Headers:     headers,
		Body:        data,
		FormData:    formData,
		FormFiles:   formFiles,
		DryRun:      dryRun,
	}

	// Add path parameters
	pathParams := make(map[string]string)
	for i, param := range endpoint.Parameters {
		if param.In == "path" && i < len(args) {
			pathParams[param.Name] = args[i]
			req.Path = strings.ReplaceAll(req.Path, fmt.Sprintf("{%s}", param.Name), args[i])
		}
	}

	// Add query parameters
	for _, param := range endpoint.Parameters {
		if param.In == "query" {
			// Check if the parameter was provided as a flag
			value, err := utils.GetStringFlagValue(cmd.Flags(), param.Name)
			if err == nil && value != "" {
				req.QueryParams.Add(param.Name, value)
			}
		}
	}

	// Add query parameters from the query flags
	for k, v := range queryParams {
		req.QueryParams.Add(k, v)
	}

	// Execute the request
	resp, err := client.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// Check if this is a dry run
	if dryRun {
		fmt.Println("Dry run completed. No request was sent.")
		return nil
	}

	// Process the response
	var responseData interface{}
	if len(resp.Body) > 0 {
		// Try to parse the response as JSON
		if err := json.Unmarshal(resp.Body, &responseData); err != nil {
			// If parsing fails, use the raw response
			responseData = string(resp.Body)
		}
	}

	// Extract fields if requested
	if len(extractFields) > 0 {
		responseData, err = output.ExtractFields(responseData, extractFields)
		if err != nil {
			log.Warn("Failed to extract fields", "error", err)
		}
	}

	// Filter the response if requested
	if filter != "" {
		responseData, err = output.FilterData(responseData, filter)
		if err != nil {
			log.Warn("Failed to filter response", "error", err)
		}
	}

	// Create a formatter
	formatter, err := output.NewFormatter(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Write the output
	if err := output.WriteOutput(responseData, formatter, savePath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// convertParameters converts openapi.Parameter to utils.Parameter
func convertParameters(params []openapi.Parameter) []utils.Parameter {
	var result []utils.Parameter
	for _, param := range params {
		// Skip path parameters
		if param.In == "path" {
			continue
		}

		// Convert the parameter
		result = append(result, utils.Parameter{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
			Required:    param.Required,
			Schema: &utils.ParameterSchema{
				Type:    param.Schema.Type,
				Format:  param.Schema.Format,
				Default: param.Schema.Default,
				Enum:    param.Schema.Enum,
			},
		})
	}
	return result
}
