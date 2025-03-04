package openapi

import (
	"net/http"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Endpoint represents an API endpoint
type Endpoint struct {
	// Path is the URL path of the endpoint
	Path string

	// Method is the HTTP method (GET, POST, etc.)
	Method string

	// OperationID is the unique identifier for the operation
	OperationID string

	// Summary is a brief description of the endpoint
	Summary string

	// Description is a detailed description of the endpoint
	Description string

	// Parameters is a list of parameters for the endpoint
	Parameters []Parameter

	// RequestBody is the request body schema
	RequestBody *RequestBody

	// Responses is a map of response codes to response schemas
	Responses map[string]Response

	// Tags is a list of tags for the endpoint
	Tags []string

	// Security is a list of security requirements for the endpoint
	Security []map[string][]string

	// Deprecated indicates if the endpoint is deprecated
	Deprecated bool
}

// Parameter represents an API parameter
type Parameter struct {
	// Name is the name of the parameter
	Name string

	// In is the location of the parameter (path, query, header, cookie)
	In string

	// Description is a description of the parameter
	Description string

	// Required indicates if the parameter is required
	Required bool

	// Schema is the schema of the parameter
	Schema *Schema

	// Example is an example value for the parameter
	Example interface{}

	// Deprecated indicates if the parameter is deprecated
	Deprecated bool
}

// Schema represents a JSON Schema
type Schema struct {
	// Type is the type of the schema (string, number, integer, boolean, array, object)
	Type string

	// Format is the format of the schema (date-time, email, etc.)
	Format string

	// Description is a description of the schema
	Description string

	// Default is the default value for the schema
	Default interface{}

	// Enum is a list of allowed values for the schema
	Enum []interface{}

	// Minimum is the minimum value for the schema
	Minimum *float64

	// Maximum is the maximum value for the schema
	Maximum *float64

	// MinLength is the minimum length for the schema
	MinLength *uint64

	// MaxLength is the maximum length for the schema
	MaxLength *uint64

	// Pattern is a regex pattern for the schema
	Pattern string

	// Properties is a map of property names to schemas (for object types)
	Properties map[string]*Schema

	// Items is the schema for array items (for array types)
	Items *Schema

	// Required is a list of required properties (for object types)
	Required []string

	// Example is an example value for the schema
	Example interface{}
}

// RequestBody represents a request body
type RequestBody struct {
	// Description is a description of the request body
	Description string

	// Required indicates if the request body is required
	Required bool

	// Content is a map of media types to schemas
	Content map[string]*MediaType
}

// MediaType represents a media type
type MediaType struct {
	// Schema is the schema of the media type
	Schema *Schema

	// Example is an example value for the media type
	Example interface{}
}

// Response represents an API response
type Response struct {
	// Description is a description of the response
	Description string

	// Content is a map of media types to schemas
	Content map[string]*MediaType

	// Headers is a map of header names to schemas
	Headers map[string]*Schema
}

// SpecParser is the interface for parsing OpenAPI specifications
type SpecParser interface {
	// ParseSpec parses an OpenAPI specification from a file or URL
	ParseSpec(specPath string) (*v3.Document, error)

	// GetEndpoints returns a list of endpoints from an OpenAPI document
	GetEndpoints(doc *v3.Document) ([]Endpoint, error)

	// GetEndpoint returns a specific endpoint from an OpenAPI document
	GetEndpoint(doc *v3.Document, path, method string) (*Endpoint, error)
}

// SecurityRequirement represents a security requirement
type SecurityRequirement struct {
	// Name is the name of the security scheme
	Name string

	// Scopes is a list of scopes required for the security scheme
	Scopes []string
}

// HTTPMethod represents an HTTP method
type HTTPMethod string

// HTTP methods
const (
	MethodGet     HTTPMethod = http.MethodGet
	MethodPost    HTTPMethod = http.MethodPost
	MethodPut     HTTPMethod = http.MethodPut
	MethodDelete  HTTPMethod = http.MethodDelete
	MethodPatch   HTTPMethod = http.MethodPatch
	MethodHead    HTTPMethod = http.MethodHead
	MethodOptions HTTPMethod = http.MethodOptions
	MethodTrace   HTTPMethod = http.MethodTrace
)

// ParameterLocation represents the location of a parameter
type ParameterLocation string

// Parameter locations
const (
	ParameterInPath   ParameterLocation = "path"
	ParameterInQuery  ParameterLocation = "query"
	ParameterInHeader ParameterLocation = "header"
	ParameterInCookie ParameterLocation = "cookie"
)

// SchemaType represents a schema type
type SchemaType string

// Schema types
const (
	SchemaTypeString  SchemaType = "string"
	SchemaTypeNumber  SchemaType = "number"
	SchemaTypeInteger SchemaType = "integer"
	SchemaTypeBoolean SchemaType = "boolean"
	SchemaTypeArray   SchemaType = "array"
	SchemaTypeObject  SchemaType = "object"
)
