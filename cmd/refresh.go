package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fynxlabs/ontap/internal/pkg/cache"
	"github.com/fynxlabs/ontap/internal/pkg/config"
	"github.com/spf13/cobra"
)

var (
	// refreshCmd represents the refresh command
	refreshCmd = &cobra.Command{
		Use:   "refresh [api-name]",
		Short: "Refresh cached OpenAPI specs",
		Long: `Refresh the cached OpenAPI specs for one or all APIs.
This will re-download and re-parse the OpenAPI specs, updating the cache.

Examples:
  # Refresh all API specs
  ontap refresh

  # Refresh a specific API spec
  ontap refresh my-api`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load the config
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create a cache manager
			cacheManager, err := cache.NewLibOpenAPICacheManager("")
			if err != nil {
				return fmt.Errorf("failed to create cache manager: %w", err)
			}

			// Check if an API name was provided
			if len(args) > 0 {
				// Refresh a specific API
				apiName := args[0]
				apiConfig, ok := cfg.APIs[apiName]
				if !ok {
					return fmt.Errorf("API not found: %s", apiName)
				}

				// Refresh the API spec
				if err := refreshAPISpec(cacheManager, apiName, apiConfig); err != nil {
					return fmt.Errorf("failed to refresh API spec: %w", err)
				}
			} else {
				// Refresh all APIs
				for apiName, apiConfig := range cfg.APIs {
					if err := refreshAPISpec(cacheManager, apiName, apiConfig); err != nil {
						log.Error("Failed to refresh API spec", "api", apiName, "error", err)
						continue
					}
				}
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(refreshCmd)
}

// refreshAPISpec refreshes an API spec
func refreshAPISpec(cacheManager *cache.LibOpenAPICacheManager, apiName string, apiConfig config.APIConfig) error {
	log.Info("Refreshing API spec", "api", apiName)

	// Get the cache TTL
	ttl := apiConfig.CacheTTL.Duration
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	// Refresh the spec
	_, err := cacheManager.RefreshSpec(apiConfig.APISpec, ttl)
	if err != nil {
		return fmt.Errorf("failed to refresh spec: %w", err)
	}

	log.Info("Refreshed API spec", "api", apiName)
	return nil
}
