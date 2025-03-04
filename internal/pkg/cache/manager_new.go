package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fynxlabs/ontap/internal/pkg/openapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// LibOpenAPICacheManager manages the caching of OpenAPI specs using libopenapi
type LibOpenAPICacheManager struct {
	// Store is the cache store
	Store LibOpenAPICacheStore
}

// NewLibOpenAPICacheManager creates a new LibOpenAPICacheManager
func NewLibOpenAPICacheManager(cacheDir string) (*LibOpenAPICacheManager, error) {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}

	store, err := NewLibOpenAPIFileSystemCacheStore(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache store: %w", err)
	}

	return &LibOpenAPICacheManager{
		Store: store,
	}, nil
}

// GetSpec retrieves a cached spec or loads it from the source
func (m *LibOpenAPICacheManager) GetSpec(specPath string, ttl time.Duration) (*v3.Document, error) {
	// Generate a cache key for the spec
	key := m.generateCacheKey(specPath)

	// Try to get from cache
	entry, err := m.Store.Get(key)
	if err == nil && entry != nil && !entry.IsExpired() {
		log.Info("Using cached OpenAPI spec", "path", specPath)
		return entry.Spec, nil
	}

	// Load the spec from the source
	log.Info("Loading OpenAPI spec", "path", specPath)
	spec, err := LoadLibOpenAPISpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	// Cache the spec
	if err := m.Store.Set(key, spec, ttl); err != nil {
		log.Warn("Failed to cache spec", "error", err)
	}

	return spec, nil
}

// RefreshSpec refreshes a cached spec
func (m *LibOpenAPICacheManager) RefreshSpec(specPath string, ttl time.Duration) (*v3.Document, error) {
	// Generate a cache key for the spec
	key := m.generateCacheKey(specPath)

	// Delete the cached spec
	if err := m.Store.Delete(key); err != nil {
		log.Warn("Failed to delete cached spec", "error", err)
	}

	// Load the spec from the source
	log.Info("Refreshing OpenAPI spec", "path", specPath)
	spec, err := LoadLibOpenAPISpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	// Cache the spec
	if err := m.Store.Set(key, spec, ttl); err != nil {
		log.Warn("Failed to cache spec", "error", err)
	}

	return spec, nil
}

// ClearCache clears the entire cache
func (m *LibOpenAPICacheManager) ClearCache() error {
	return m.Store.Clear()
}

// generateCacheKey generates a cache key for a spec path
func (m *LibOpenAPICacheManager) generateCacheKey(specPath string) string {
	// Use a hash of the spec path as the key
	hash := sha256.Sum256([]byte(specPath))
	return hex.EncodeToString(hash[:])
}

// LoadLibOpenAPISpec loads an OpenAPI specification from a file or URL using libopenapi
func LoadLibOpenAPISpec(specPath string) (*v3.Document, error) {
	parser := openapi.NewLibOpenAPISpecParser()
	return parser.ParseSpec(specPath)
}

// IsLibOpenAPIURL checks if a string is a URL
func IsLibOpenAPIURL(s string) bool {
	return s != "" && (s[:7] == "http://" || s[:8] == "https://")
}

// DownloadLibOpenAPISpec downloads an OpenAPI specification from a URL to a file
func DownloadLibOpenAPISpec(url, destPath string) error {
	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Fetch the spec from the URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch spec from URL: %w", err)
	}
	defer resp.Body.Close()

	// Create the destination file
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	// Copy the response body to the destination file
	_, err = dest.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write spec to file: %w", err)
	}

	log.Info("Downloaded spec", "url", url, "path", destPath)
	return nil
}
