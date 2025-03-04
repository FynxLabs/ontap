package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v2"
)

// OpenAPIVersion represents an OpenAPI specification version
type OpenAPIVersion string

const (
	// OpenAPIV30 represents OpenAPI 3.0.x
	OpenAPIV30 OpenAPIVersion = "3.0"

	// OpenAPIV31 represents OpenAPI 3.1.x
	OpenAPIV31 OpenAPIVersion = "3.1"

	// OpenAPIUnknown represents an unknown OpenAPI version
	OpenAPIUnknown OpenAPIVersion = "unknown"
)

// VersionDetector is the interface for detecting OpenAPI versions
type VersionDetector interface {
	// DetectVersion detects the OpenAPI version from a file or URL
	DetectVersion(specPath string) (OpenAPIVersion, error)

	// DetectVersionFromBytes detects the OpenAPI version from a byte slice
	DetectVersionFromBytes(data []byte) (OpenAPIVersion, error)
}

// DefaultVersionDetector implements VersionDetector
type DefaultVersionDetector struct{}

// NewVersionDetector creates a new DefaultVersionDetector
func NewVersionDetector() *DefaultVersionDetector {
	return &DefaultVersionDetector{}
}

// DetectVersion detects the OpenAPI version from a file or URL
func (d *DefaultVersionDetector) DetectVersion(specPath string) (OpenAPIVersion, error) {
	// Check if the spec is a URL
	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		return d.detectVersionFromURL(specPath)
	}

	// Read the spec file
	data, err := os.ReadFile(specPath)
	if err != nil {
		return OpenAPIUnknown, fmt.Errorf("failed to read spec file: %w", err)
	}

	return d.DetectVersionFromBytes(data)
}

// DetectVersionFromBytes detects the OpenAPI version from a byte slice
func (d *DefaultVersionDetector) DetectVersionFromBytes(data []byte) (OpenAPIVersion, error) {
	// Try to unmarshal as JSON first
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		// Try YAML if JSON fails
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return OpenAPIUnknown, fmt.Errorf("failed to parse spec: %w", err)
		}
	}

	// Check for OpenAPI 3.x
	if openapi, ok := doc["openapi"].(string); ok {
		if strings.HasPrefix(openapi, "3.1") {
			return OpenAPIV31, nil
		} else if strings.HasPrefix(openapi, "3.0") {
			return OpenAPIV30, nil
		} else if strings.HasPrefix(openapi, "3") {
			// Generic v3, assume 3.0 compatibility
			log.Warn("Detected generic OpenAPI 3.x version, assuming 3.0 compatibility", "version", openapi)
			return OpenAPIV30, nil
		}
	}

	return OpenAPIUnknown, fmt.Errorf("unsupported or unrecognized OpenAPI version (only 3.0 and 3.1 are supported)")
}

// detectVersionFromURL detects the OpenAPI version from a URL
func (d *DefaultVersionDetector) detectVersionFromURL(url string) (OpenAPIVersion, error) {
	// Fetch the spec from the URL
	resp, err := http.Get(url)
	if err != nil {
		return OpenAPIUnknown, fmt.Errorf("failed to fetch spec from URL: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return OpenAPIUnknown, fmt.Errorf("failed to read response body: %w", err)
	}

	return d.DetectVersionFromBytes(data)
}
