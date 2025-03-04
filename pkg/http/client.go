package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// Client is the HTTP client for making API requests
type Client struct {
	// BaseURL is the base URL for the API
	BaseURL string

	// Headers are the default headers to include with every request
	Headers map[string]string

	// Auth is the authentication string
	Auth string

	// Timeout is the request timeout
	Timeout time.Duration

	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client

	// Verbose indicates whether to log verbose output
	Verbose bool
}

// NewClient creates a new HTTP client
func NewClient(baseURL, auth string) *Client {
	return &Client{
		BaseURL: baseURL,
		Auth:    auth,
		Headers: map[string]string{
			"User-Agent": "OnTap CLI",
		},
		Timeout: 30 * time.Second,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Verbose: false,
	}
}

// Request represents an HTTP request
type Request struct {
	// Method is the HTTP method
	Method string

	// Path is the URL path
	Path string

	// QueryParams are the query parameters
	QueryParams url.Values

	// Headers are the request headers
	Headers map[string]string

	// Body is the request body
	Body interface{}

	// FormData is the form data
	FormData map[string]string

	// FormFiles are the files to upload
	FormFiles map[string]string

	// Auth is the authentication string
	Auth string

	// DryRun indicates whether to perform a dry run
	DryRun bool
}

// Response represents an HTTP response
type Response struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Headers are the response headers
	Headers http.Header

	// Body is the response body
	Body []byte

	// Request is the original request
	Request *Request

	// Duration is the request duration
	Duration time.Duration
}

// Execute executes an HTTP request
func (c *Client) Execute(req *Request) (*Response, error) {
	// Start the timer
	start := time.Now()

	// Create the URL
	reqURL, err := c.buildURL(req.Path, req.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Create the request body
	var reqBody io.Reader
	var contentType string
	if req.FormData != nil || req.FormFiles != nil {
		reqBody, contentType, err = c.createFormBody(req.FormData, req.FormFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to create form body: %w", err)
		}
	} else if req.Body != nil {
		reqBody, contentType, err = c.createJSONBody(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON body: %w", err)
		}
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest(req.Method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Add authentication
	auth := req.Auth
	if auth == "" {
		auth = c.Auth
	}
	if auth != "" {
		c.addAuth(httpReq, auth)
	}

	// Log the request
	if c.Verbose || req.DryRun {
		c.logRequest(httpReq, req.Body)
	}

	// If this is a dry run, return a dummy response
	if req.DryRun {
		return &Response{
			StatusCode: 0,
			Headers:    nil,
			Body:       nil,
			Request:    req,
			Duration:   time.Since(start),
		}, nil
	}

	// Execute the request
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create the response
	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       respBody,
		Request:    req,
		Duration:   time.Since(start),
	}

	// Log the response
	if c.Verbose {
		c.logResponse(resp)
	}

	return resp, nil
}

// Get executes a GET request
func (c *Client) Get(path string, queryParams url.Values, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:      http.MethodGet,
		Path:        path,
		QueryParams: queryParams,
		Headers:     headers,
	}
	return c.Execute(req)
}

// Post executes a POST request
func (c *Client) Post(path string, body interface{}, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPost,
		Path:    path,
		Body:    body,
		Headers: headers,
	}
	return c.Execute(req)
}

// Put executes a PUT request
func (c *Client) Put(path string, body interface{}, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPut,
		Path:    path,
		Body:    body,
		Headers: headers,
	}
	return c.Execute(req)
}

// Patch executes a PATCH request
func (c *Client) Patch(path string, body interface{}, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodPatch,
		Path:    path,
		Body:    body,
		Headers: headers,
	}
	return c.Execute(req)
}

// Delete executes a DELETE request
func (c *Client) Delete(path string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:  http.MethodDelete,
		Path:    path,
		Headers: headers,
	}
	return c.Execute(req)
}

// PostForm executes a POST request with form data
func (c *Client) PostForm(path string, formData map[string]string, formFiles map[string]string, headers map[string]string) (*Response, error) {
	req := &Request{
		Method:    http.MethodPost,
		Path:      path,
		FormData:  formData,
		FormFiles: formFiles,
		Headers:   headers,
	}
	return c.Execute(req)
}

// buildURL builds the full URL for a request
func (c *Client) buildURL(path string, queryParams url.Values) (string, error) {
	// Parse the base URL
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Parse the path
	pathURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Resolve the path against the base URL
	fullURL := baseURL.ResolveReference(pathURL)

	// Add query parameters
	if queryParams != nil {
		fullURL.RawQuery = queryParams.Encode()
	}

	return fullURL.String(), nil
}

// createJSONBody creates a JSON request body
func (c *Client) createJSONBody(body interface{}) (io.Reader, string, error) {
	// Marshal the body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return bytes.NewReader(jsonBody), "application/json", nil
}

// createFormBody creates a multipart form request body
func (c *Client) createFormBody(formData map[string]string, formFiles map[string]string) (io.Reader, string, error) {
	// Create a buffer to write the form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	for k, v := range formData {
		if err := writer.WriteField(k, v); err != nil {
			return nil, "", fmt.Errorf("failed to write form field: %w", err)
		}
	}

	// Add form files
	for k, v := range formFiles {
		// Check if the value is a file path
		if strings.HasPrefix(v, "@") {
			filePath := v[1:]
			file, err := os.Open(filePath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			// Create a form file
			part, err := writer.CreateFormFile(k, filepath.Base(filePath))
			if err != nil {
				return nil, "", fmt.Errorf("failed to create form file: %w", err)
			}

			// Copy the file to the form
			if _, err := io.Copy(part, file); err != nil {
				return nil, "", fmt.Errorf("failed to copy file: %w", err)
			}
		} else {
			// Add as a regular form field
			if err := writer.WriteField(k, v); err != nil {
				return nil, "", fmt.Errorf("failed to write form field: %w", err)
			}
		}
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close form writer: %w", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

// addAuth adds authentication to a request
func (c *Client) addAuth(req *http.Request, auth string) {
	// Check if the auth is a basic auth string (username:password)
	if strings.Contains(auth, ":") {
		parts := strings.SplitN(auth, ":", 2)
		req.SetBasicAuth(parts[0], parts[1])
	} else if strings.HasPrefix(auth, "Bearer ") {
		// Bearer token
		req.Header.Set("Authorization", auth)
	} else if strings.HasPrefix(auth, "Basic ") {
		// Basic auth token
		req.Header.Set("Authorization", auth)
	} else {
		// API key
		req.Header.Set("X-API-Key", auth)
	}
}

// logRequest logs a request
func (c *Client) logRequest(req *http.Request, body interface{}) {
	log.Info("Request", "method", req.Method, "url", req.URL.String())
	log.Info("Request Headers", "headers", req.Header)
	if body != nil {
		jsonBody, _ := json.MarshalIndent(body, "", "  ")
		log.Info("Request Body", "body", string(jsonBody))
	}
}

// logResponse logs a response
func (c *Client) logResponse(resp *Response) {
	log.Info("Response", "status", resp.StatusCode, "duration", resp.Duration)
	log.Info("Response Headers", "headers", resp.Headers)
	if len(resp.Body) > 0 {
		// Try to pretty-print JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, resp.Body, "", "  "); err == nil {
			log.Info("Response Body", "body", prettyJSON.String())
		} else {
			log.Info("Response Body", "body", string(resp.Body))
		}
	}
}

// ParseBasicAuth parses a basic auth string
func ParseBasicAuth(auth string) (username, password string, ok bool) {
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", false
	}
	return cs[:s], cs[s+1:], true
}

// ParseBearerToken parses a bearer token
func ParseBearerToken(auth string) (token string, ok bool) {
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", false
	}
	return auth[7:], true
}
