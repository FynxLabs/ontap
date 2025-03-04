package config

import (
	"encoding/json"
	"time"
)

// Config represents the main configuration structure for OnTap
type Config struct {
	APIs map[string]APIConfig `yaml:"apis" json:"apis"`
}

// APIConfig represents the configuration for a single API
type APIConfig struct {
	// APISpec is the path or URL to the OpenAPI specification
	APISpec string `yaml:"apispec" json:"apispec"`

	// Auth is the authentication string (format depends on auth type)
	// Examples:
	// - Basic auth: "username:password"
	// - Bearer token: "Bearer token123"
	// - API key: "key123"
	Auth string `yaml:"auth" json:"auth"`

	// URL is the base URL for the API
	URL string `yaml:"url" json:"url"`

	// CacheTTL is the time-to-live for the cached OpenAPI spec
	// Format: time.Duration string (e.g., "24h", "30m")
	CacheTTL Duration `yaml:"cache_ttl" json:"cache_ttl" default:"24h"`

	// DefaultOutput is the default output format for this API
	DefaultOutput string `yaml:"output" json:"output" default:"json"`

	// Headers are additional headers to include with every request
	Headers map[string]string `yaml:"headers" json:"headers"`
}

// Duration is a wrapper around time.Duration for YAML/JSON marshaling
type Duration struct {
	time.Duration
}

// String returns the string representation of the duration
func (d Duration) String() string {
	return d.Duration.String()
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
