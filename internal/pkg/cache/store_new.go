package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// LibOpenAPICacheStore is the interface for caching OpenAPI specs using libopenapi
type LibOpenAPICacheStore interface {
	// Get retrieves a cached spec
	Get(key string) (*LibOpenAPICacheEntry, error)

	// Set stores a spec in the cache
	Set(key string, spec *v3.Document, ttl time.Duration) error

	// Delete removes a spec from the cache
	Delete(key string) error

	// Clear clears the entire cache
	Clear() error

	// GetCachePath returns the path to the cache file for a key
	GetCachePath(key string) string
}

// LibOpenAPICacheEntry represents a cached OpenAPI spec using libopenapi
type LibOpenAPICacheEntry struct {
	// Spec is the OpenAPI spec
	Spec *v3.Document

	// CreatedAt is the time the entry was created
	CreatedAt time.Time

	// ExpiresAt is the time the entry expires
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry is expired
func (e *LibOpenAPICacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// LibOpenAPIFileSystemCacheStore implements LibOpenAPICacheStore using the file system
type LibOpenAPIFileSystemCacheStore struct {
	// CacheDir is the directory where cache files are stored
	CacheDir string

	// Memory cache for faster access
	memoryCache map[string]*LibOpenAPICacheEntry

	// Mutex for thread safety
	mutex sync.RWMutex
}

// NewLibOpenAPIFileSystemCacheStore creates a new LibOpenAPIFileSystemCacheStore
func NewLibOpenAPIFileSystemCacheStore(cacheDir string) (*LibOpenAPIFileSystemCacheStore, error) {
	// Create the cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &LibOpenAPIFileSystemCacheStore{
		CacheDir:    cacheDir,
		memoryCache: make(map[string]*LibOpenAPICacheEntry),
	}, nil
}

// Get retrieves a cached spec
func (s *LibOpenAPIFileSystemCacheStore) Get(key string) (*LibOpenAPICacheEntry, error) {
	s.mutex.RLock()
	// Check memory cache first
	if entry, ok := s.memoryCache[key]; ok {
		s.mutex.RUnlock()
		if entry.IsExpired() {
			// Remove expired entry
			if err := s.Delete(key); err != nil {
				log.Warn("Failed to delete expired cache entry", "key", key, "error", err)
			}
			return nil, fmt.Errorf("cache entry expired")
		}
		return entry, nil
	}
	s.mutex.RUnlock()

	// Check file system
	cachePath := s.GetCachePath(key)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Unmarshal the cache entry
	var entry LibOpenAPICacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if the entry is expired
	if entry.IsExpired() {
		// Remove expired entry
		if err := s.Delete(key); err != nil {
			log.Warn("Failed to delete expired cache entry", "key", key, "error", err)
		}
		return nil, fmt.Errorf("cache entry expired")
	}

	// Add to memory cache
	s.mutex.Lock()
	s.memoryCache[key] = &entry
	s.mutex.Unlock()

	return &entry, nil
}

// Set stores a spec in the cache
func (s *LibOpenAPIFileSystemCacheStore) Set(key string, spec *v3.Document, ttl time.Duration) error {
	// Create the cache entry
	now := time.Now()
	entry := &LibOpenAPICacheEntry{
		Spec:      spec,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	// Add to memory cache
	s.mutex.Lock()
	s.memoryCache[key] = entry
	s.mutex.Unlock()

	// Marshal the cache entry
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write to file
	cachePath := s.GetCachePath(key)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Info("Cached OpenAPI spec", "key", key, "path", cachePath, "ttl", ttl)
	return nil
}

// Delete removes a spec from the cache
func (s *LibOpenAPIFileSystemCacheStore) Delete(key string) error {
	// Remove from memory cache
	s.mutex.Lock()
	delete(s.memoryCache, key)
	s.mutex.Unlock()

	// Remove from file system
	cachePath := s.GetCachePath(key)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	log.Info("Removed cached OpenAPI spec", "key", key)
	return nil
}

// Clear clears the entire cache
func (s *LibOpenAPIFileSystemCacheStore) Clear() error {
	// Clear memory cache
	s.mutex.Lock()
	s.memoryCache = make(map[string]*LibOpenAPICacheEntry)
	s.mutex.Unlock()

	// Clear file system cache
	if err := os.RemoveAll(s.CacheDir); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	// Recreate the cache directory
	if err := os.MkdirAll(s.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	log.Info("Cleared cache", "dir", s.CacheDir)
	return nil
}

// GetCachePath returns the path to the cache file for a key
func (s *LibOpenAPIFileSystemCacheStore) GetCachePath(key string) string {
	// Create a safe filename from the key
	safeKey := filepath.Base(key)
	return filepath.Join(s.CacheDir, safeKey+".json")
}
