package openapi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// stringValue returns the string value of a pointer or an empty string if nil
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// pointerToUint64 creates a pointer to a uint64 value
func pointerToUint64(value uint64) *uint64 {
	return &value
}

// DefaultSpecParser implements SpecParser
type DefaultSpecParser struct {
	detector VersionDetector
}

// NewSpecParser creates a new DefaultSpecParser
func NewSpecParser() *DefaultSpecParser {
	return &DefaultSpecParser{
		detector: NewVersionDetector(),
	}
}

// ParseSpec parses an OpenAPI specification from a file or URL
func (p *DefaultSpecParser) ParseSpec(specPath string) (*v3.Document, error) {
	// Detect the OpenAPI version
	version, err := p.detector.DetectVersion(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect OpenAPI version: %w", err)
	}

	log.Info("Detected OpenAPI version", "version", version)

	// Parse the spec based on the version
	switch version {
	case OpenAPIV30, OpenAPIV31:
		return p.parseOpenAPIV3(specPath)
	default:
		return nil, fmt.Errorf("unsupported OpenAPI version: %s", version)
	}
}

// parseOpenAPIV3 parses an OpenAPI 3.x specification
func (p *DefaultSpecParser) parseOpenAPIV3(specPath string) (*v3.Document, error) {
	var data []byte

	// Check if the spec is a URL
	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		// Create a configuration with a base URL
		baseURL, err := url.Parse(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		// Fetch the spec from the URL
		resp, err := http.Get(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch spec from URL: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Create a document configuration
		config := &datamodel.DocumentConfiguration{
			AllowRemoteReferences: true,
			BaseURL:               baseURL,
		}

		// Create a new document
		doc, err := libopenapi.NewDocumentWithConfiguration(data, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}

		// Build the V3 model
		model, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			return nil, fmt.Errorf("failed to build model: %v", errs)
		}

		return &model.Model, nil
	} else {
		// Get the absolute path
		absPath, err := filepath.Abs(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Read the spec file
		data, err = os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file: %w", err)
		}

		// Create a document configuration
		config := &datamodel.DocumentConfiguration{
			AllowFileReferences: true,
			BasePath:            filepath.Dir(absPath),
		}

		// Create a new document
		doc, err := libopenapi.NewDocumentWithConfiguration(data, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}

		// Build the V3 model
		model, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			return nil, fmt.Errorf("failed to build model: %v", errs)
		}

		return &model.Model, nil
	}
}

// GetEndpoints returns a list of endpoints from an OpenAPI document
func (p *DefaultSpecParser) GetEndpoints(doc *v3.Document) ([]Endpoint, error) {
	if doc == nil {
		return nil, fmt.Errorf("OpenAPI document is nil")
	}

	endpoints := []Endpoint{}

	// Iterate over all paths
	for pathPairs := doc.Paths.PathItems.First(); pathPairs != nil; pathPairs = pathPairs.Next() {
		path := pathPairs.Key()
		pathItem := pathPairs.Value()

		// Process GET operations
		if pathItem.Get != nil {
			endpoint, err := p.createEndpoint(path, "GET", pathItem.Get)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "GET", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process POST operations
		if pathItem.Post != nil {
			endpoint, err := p.createEndpoint(path, "POST", pathItem.Post)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "POST", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process PUT operations
		if pathItem.Put != nil {
			endpoint, err := p.createEndpoint(path, "PUT", pathItem.Put)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "PUT", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process DELETE operations
		if pathItem.Delete != nil {
			endpoint, err := p.createEndpoint(path, "DELETE", pathItem.Delete)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "DELETE", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process PATCH operations
		if pathItem.Patch != nil {
			endpoint, err := p.createEndpoint(path, "PATCH", pathItem.Patch)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "PATCH", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process OPTIONS operations
		if pathItem.Options != nil {
			endpoint, err := p.createEndpoint(path, "OPTIONS", pathItem.Options)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "OPTIONS", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process HEAD operations
		if pathItem.Head != nil {
			endpoint, err := p.createEndpoint(path, "HEAD", pathItem.Head)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "HEAD", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}

		// Process TRACE operations
		if pathItem.Trace != nil {
			endpoint, err := p.createEndpoint(path, "TRACE", pathItem.Trace)
			if err != nil {
				log.Warn("Failed to create endpoint", "path", path, "method", "TRACE", "error", err)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}
	}

	return endpoints, nil
}

// GetEndpoint returns a specific endpoint from an OpenAPI document
func (p *DefaultSpecParser) GetEndpoint(doc *v3.Document, path, method string) (*Endpoint, error) {
	if doc == nil {
		return nil, fmt.Errorf("OpenAPI document is nil")
	}

	// Find the path item
	var pathItem *v3.PathItem
	for pathPairs := doc.Paths.PathItems.First(); pathPairs != nil; pathPairs = pathPairs.Next() {
		if pathPairs.Key() == path {
			pathItem = pathPairs.Value()
			break
		}
	}

	if pathItem == nil {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	// Get the operation based on the method
	var operation *v3.Operation
	switch strings.ToUpper(method) {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	case "OPTIONS":
		operation = pathItem.Options
	case "HEAD":
		operation = pathItem.Head
	case "TRACE":
		operation = pathItem.Trace
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if operation == nil {
		return nil, fmt.Errorf("method not found: %s %s", method, path)
	}

	// Create an endpoint
	return p.createEndpoint(path, method, operation)
}

// createEndpoint creates an Endpoint from an OpenAPI operation
func (p *DefaultSpecParser) createEndpoint(path, method string, operation *v3.Operation) (*Endpoint, error) {
	// Check if operation is deprecated
	deprecated := false
	if operation.Deprecated != nil {
		deprecated = *operation.Deprecated
	}

	// Create the endpoint
	endpoint := &Endpoint{
		Path:        path,
		Method:      method,
		OperationID: operation.OperationId,
		Summary:     operation.Summary,
		Description: operation.Description,
		Tags:        operation.Tags,
		Deprecated:  deprecated,
		Parameters:  []Parameter{},
		Responses:   map[string]Response{},
		Security:    []map[string][]string{},
	}

	// Add parameters
	for _, param := range operation.Parameters {
		parameter, err := p.createParameter(param)
		if err != nil {
			log.Warn("Failed to create parameter", "name", param.Name, "error", err)
			continue
		}

		endpoint.Parameters = append(endpoint.Parameters, *parameter)
	}

	// Add request body
	if operation.RequestBody != nil {
		requestBody, err := p.createRequestBody(operation.RequestBody)
		if err != nil {
			log.Warn("Failed to create request body", "error", err)
		} else {
			endpoint.RequestBody = requestBody
		}
	}

	// Add responses
	for codePairs := operation.Responses.Codes.First(); codePairs != nil; codePairs = codePairs.Next() {
		code := codePairs.Key()
		response := codePairs.Value()

		resp, err := p.createResponse(response)
		if err != nil {
			log.Warn("Failed to create response", "statusCode", code, "error", err)
			continue
		}

		endpoint.Responses[code] = *resp
	}

	// Add security
	if operation.Security != nil {
		for _, securityRequirement := range operation.Security {
			securityMap := make(map[string][]string)

			// Iterate through the security requirements
			for reqPairs := securityRequirement.Requirements.First(); reqPairs != nil; reqPairs = reqPairs.Next() {
				name := reqPairs.Key()
				scopes := reqPairs.Value()
				securityMap[name] = scopes
			}

			endpoint.Security = append(endpoint.Security, securityMap)
		}
	}

	return endpoint, nil
}

// createParameter creates a Parameter from an OpenAPI parameter
func (p *DefaultSpecParser) createParameter(param *v3.Parameter) (*Parameter, error) {
	if param == nil {
		return nil, fmt.Errorf("parameter is nil")
	}

	// Get required value
	required := false
	if param.Required != nil {
		required = *param.Required
	}

	// Get deprecated value
	deprecated := param.Deprecated

	// Create the parameter
	parameter := &Parameter{
		Name:        param.Name,
		In:          param.In,
		Description: param.Description,
		Required:    required,
		Deprecated:  deprecated,
		Example:     param.Example,
	}

	// Add schema
	if param.Schema != nil {
		schema, err := p.createSchemaFromProxy(param.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create schema: %w", err)
		}

		parameter.Schema = schema
	}

	return parameter, nil
}

// createSchemaFromProxy creates a Schema from an OpenAPI schema proxy
func (p *DefaultSpecParser) createSchemaFromProxy(schemaProxy *base.SchemaProxy) (*Schema, error) {
	if schemaProxy == nil {
		return nil, fmt.Errorf("schema proxy is nil")
	}

	// Get the actual schema
	schema := schemaProxy.Schema()
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	return p.createSchema(schema)
}

// createSchema creates a Schema from an OpenAPI schema
func (p *DefaultSpecParser) createSchema(schema *base.Schema) (*Schema, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	// Create the schema
	s := &Schema{
		Type:        strings.Join(schema.Type, ","), // Convert []string to string
		Format:      schema.Format,
		Description: schema.Description,
		Default:     schema.Default,
		Pattern:     schema.Pattern,
		Example:     schema.Example,
		Required:    schema.Required,
	}

	// Convert enum values
	if schema.Enum != nil {
		for _, enum := range schema.Enum {
			if enum != nil {
				s.Enum = append(s.Enum, enum)
			}
		}
	}

	// Handle numeric constraints
	if schema.Minimum != nil {
		min := *schema.Minimum
		s.Minimum = &min
	}
	if schema.Maximum != nil {
		max := *schema.Maximum
		s.Maximum = &max
	}

	// Handle string constraints
	if schema.MinLength != nil {
		minLength := uint64(*schema.MinLength)
		s.MinLength = &minLength
	}
	if schema.MaxLength != nil {
		maxLength := uint64(*schema.MaxLength)
		s.MaxLength = &maxLength
	}

	// Add properties
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		s.Properties = map[string]*Schema{}

		for propPairs := schema.Properties.First(); propPairs != nil; propPairs = propPairs.Next() {
			propName := propPairs.Key()
			propSchema := propPairs.Value()

			propSchemaObj, err := p.createSchemaFromProxy(propSchema)
			if err != nil {
				log.Warn("Failed to create property schema", "name", propName, "error", err)
				continue
			}

			s.Properties[propName] = propSchemaObj
		}
	}

	// Add items for array types
	if schema.Items != nil && schema.Items.A != nil {
		// Get the schema from the dynamic value
		itemsSchema, err := p.createSchemaFromProxy(schema.Items.A)
		if err != nil {
			log.Warn("Failed to create items schema", "error", err)
		} else {
			s.Items = itemsSchema
		}
	}

	return s, nil
}

// createRequestBody creates a RequestBody from an OpenAPI request body
func (p *DefaultSpecParser) createRequestBody(requestBody *v3.RequestBody) (*RequestBody, error) {
	if requestBody == nil {
		return nil, fmt.Errorf("request body is nil")
	}

	// Get required value
	required := false
	if requestBody.Required != nil {
		required = *requestBody.Required
	}

	// Create the request body
	rb := &RequestBody{
		Description: requestBody.Description,
		Required:    required,
		Content:     map[string]*MediaType{},
	}

	// Add content
	for contentPairs := requestBody.Content.First(); contentPairs != nil; contentPairs = contentPairs.Next() {
		mediaType := contentPairs.Key()
		content := contentPairs.Value()

		mt, err := p.createMediaType(content)
		if err != nil {
			log.Warn("Failed to create media type", "mediaType", mediaType, "error", err)
			continue
		}

		rb.Content[mediaType] = mt
	}

	return rb, nil
}

// createMediaType creates a MediaType from an OpenAPI media type
func (p *DefaultSpecParser) createMediaType(mediaType *v3.MediaType) (*MediaType, error) {
	if mediaType == nil {
		return nil, fmt.Errorf("media type is nil")
	}

	// Create the media type
	mt := &MediaType{
		Example: mediaType.Example,
	}

	// Add schema
	if mediaType.Schema != nil {
		schema, err := p.createSchemaFromProxy(mediaType.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to create schema: %w", err)
		}

		mt.Schema = schema
	}

	return mt, nil
}

// createResponse creates a Response from an OpenAPI response
func (p *DefaultSpecParser) createResponse(response *v3.Response) (*Response, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	// Create the response
	r := &Response{
		Description: response.Description,
		Content:     map[string]*MediaType{},
		Headers:     map[string]*Schema{},
	}

	// Add content
	for contentPairs := response.Content.First(); contentPairs != nil; contentPairs = contentPairs.Next() {
		mediaType := contentPairs.Key()
		content := contentPairs.Value()

		mt, err := p.createMediaType(content)
		if err != nil {
			log.Warn("Failed to create media type", "mediaType", mediaType, "error", err)
			continue
		}

		r.Content[mediaType] = mt
	}

	// Add headers
	for headerPairs := response.Headers.First(); headerPairs != nil; headerPairs = headerPairs.Next() {
		name := headerPairs.Key()
		header := headerPairs.Value()

		if header.Schema == nil {
			continue
		}

		schema, err := p.createSchemaFromProxy(header.Schema)
		if err != nil {
			log.Warn("Failed to create header schema", "name", name, "error", err)
			continue
		}

		r.Headers[name] = schema
	}

	return r, nil
}

// LoadSpec loads an OpenAPI specification from a file or URL
func LoadSpec(specPath string) (*v3.Document, error) {
	parser := NewSpecParser()
	return parser.ParseSpec(specPath)
}

// IsURL checks if a string is a URL
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// DownloadSpec downloads an OpenAPI specification from a URL to a file
func DownloadSpec(url, destPath string) error {
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
