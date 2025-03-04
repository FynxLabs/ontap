package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

// AuthType represents an authentication type
type AuthType string

const (
	// AuthTypeBasic represents basic authentication
	AuthTypeBasic AuthType = "basic"

	// AuthTypeBearer represents bearer token authentication
	AuthTypeBearer AuthType = "bearer"

	// AuthTypeAPIKey represents API key authentication
	AuthTypeAPIKey AuthType = "apikey"

	// AuthTypeNone represents no authentication
	AuthTypeNone AuthType = "none"
)

// AuthProvider is the interface for providing authentication
type AuthProvider interface {
	// GetAuthHeader returns the authentication header
	GetAuthHeader() (string, string)

	// GetAuthType returns the authentication type
	GetAuthType() AuthType

	// GetAuthString returns the authentication string
	GetAuthString() string
}

// BasicAuth represents basic authentication
type BasicAuth struct {
	// Username is the username
	Username string

	// Password is the password
	Password string
}

// NewBasicAuth creates a new BasicAuth
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

// NewBasicAuthFromString creates a new BasicAuth from a string
func NewBasicAuthFromString(auth string) (*BasicAuth, error) {
	// Check if the auth is a basic auth string (username:password)
	if strings.Contains(auth, ":") {
		parts := strings.SplitN(auth, ":", 2)
		return NewBasicAuth(parts[0], parts[1]), nil
	}

	// Check if the auth is a basic auth token
	if strings.HasPrefix(auth, "Basic ") {
		username, password, ok := ParseBasicAuth(auth)
		if !ok {
			return nil, fmt.Errorf("invalid basic auth token")
		}
		return NewBasicAuth(username, password), nil
	}

	return nil, fmt.Errorf("invalid basic auth string")
}

// GetAuthHeader returns the authentication header
func (a *BasicAuth) GetAuthHeader() (string, string) {
	return "Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(a.Username+":"+a.Password))
}

// GetAuthType returns the authentication type
func (a *BasicAuth) GetAuthType() AuthType {
	return AuthTypeBasic
}

// GetAuthString returns the authentication string
func (a *BasicAuth) GetAuthString() string {
	return a.Username + ":" + a.Password
}

// BearerAuth represents bearer token authentication
type BearerAuth struct {
	// Token is the bearer token
	Token string
}

// NewBearerAuth creates a new BearerAuth
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		Token: token,
	}
}

// NewBearerAuthFromString creates a new BearerAuth from a string
func NewBearerAuthFromString(auth string) (*BearerAuth, error) {
	// Check if the auth is a bearer token
	if strings.HasPrefix(auth, "Bearer ") {
		token, ok := ParseBearerToken(auth)
		if !ok {
			return nil, fmt.Errorf("invalid bearer token")
		}
		return NewBearerAuth(token), nil
	}

	// Assume the auth is a raw token
	return NewBearerAuth(auth), nil
}

// GetAuthHeader returns the authentication header
func (a *BearerAuth) GetAuthHeader() (string, string) {
	return "Authorization", "Bearer " + a.Token
}

// GetAuthType returns the authentication type
func (a *BearerAuth) GetAuthType() AuthType {
	return AuthTypeBearer
}

// GetAuthString returns the authentication string
func (a *BearerAuth) GetAuthString() string {
	return "Bearer " + a.Token
}

// APIKeyAuth represents API key authentication
type APIKeyAuth struct {
	// Key is the API key
	Key string

	// Name is the name of the API key header
	Name string

	// In is the location of the API key (header, query)
	In string
}

// NewAPIKeyAuth creates a new APIKeyAuth
func NewAPIKeyAuth(key, name, in string) *APIKeyAuth {
	if name == "" {
		name = "X-API-Key"
	}
	if in == "" {
		in = "header"
	}
	return &APIKeyAuth{
		Key:  key,
		Name: name,
		In:   in,
	}
}

// NewAPIKeyAuthFromString creates a new APIKeyAuth from a string
func NewAPIKeyAuthFromString(auth string) (*APIKeyAuth, error) {
	return NewAPIKeyAuth(auth, "", ""), nil
}

// GetAuthHeader returns the authentication header
func (a *APIKeyAuth) GetAuthHeader() (string, string) {
	return a.Name, a.Key
}

// GetAuthType returns the authentication type
func (a *APIKeyAuth) GetAuthType() AuthType {
	return AuthTypeAPIKey
}

// GetAuthString returns the authentication string
func (a *APIKeyAuth) GetAuthString() string {
	return a.Key
}

// NoAuth represents no authentication
type NoAuth struct{}

// NewNoAuth creates a new NoAuth
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// GetAuthHeader returns the authentication header
func (a *NoAuth) GetAuthHeader() (string, string) {
	return "", ""
}

// GetAuthType returns the authentication type
func (a *NoAuth) GetAuthType() AuthType {
	return AuthTypeNone
}

// GetAuthString returns the authentication string
func (a *NoAuth) GetAuthString() string {
	return ""
}

// DetectAuthType detects the authentication type from a string
func DetectAuthType(auth string) AuthType {
	if auth == "" {
		return AuthTypeNone
	}

	if strings.Contains(auth, ":") {
		return AuthTypeBasic
	}

	if strings.HasPrefix(auth, "Basic ") {
		return AuthTypeBasic
	}

	if strings.HasPrefix(auth, "Bearer ") {
		return AuthTypeBearer
	}

	// Check if it's an environment variable
	if strings.HasPrefix(auth, "${") && strings.HasSuffix(auth, "}") {
		envVar := auth[2 : len(auth)-1]
		envValue := os.Getenv(envVar)
		if envValue != "" {
			return DetectAuthType(envValue)
		}
	}

	// Default to API key
	return AuthTypeAPIKey
}

// NewAuthProvider creates a new AuthProvider based on the authentication type
func NewAuthProvider(auth string) (AuthProvider, error) {
	// Check if it's an environment variable
	if strings.HasPrefix(auth, "${") && strings.HasSuffix(auth, "}") {
		envVar := auth[2 : len(auth)-1]
		envValue := os.Getenv(envVar)
		if envValue == "" {
			log.Warn("Environment variable not set", "var", envVar)
			return NewNoAuth(), nil
		}
		auth = envValue
	}

	// Detect the authentication type
	authType := DetectAuthType(auth)

	// Create the appropriate auth provider
	switch authType {
	case AuthTypeBasic:
		return NewBasicAuthFromString(auth)
	case AuthTypeBearer:
		return NewBearerAuthFromString(auth)
	case AuthTypeAPIKey:
		return NewAPIKeyAuthFromString(auth)
	case AuthTypeNone:
		return NewNoAuth(), nil
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", authType)
	}
}

// AddAuthToRequest adds authentication to a request
func AddAuthToRequest(req *http.Request, auth string) error {
	// Create an auth provider
	provider, err := NewAuthProvider(auth)
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Add the auth header
	name, value := provider.GetAuthHeader()
	if name != "" && value != "" {
		req.Header.Set(name, value)
	}

	return nil
}
