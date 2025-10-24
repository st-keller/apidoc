package apidoc

import "net/http"

// EndpointConfig defines an API endpoint with metadata for automatic OpenAPI generation
type EndpointConfig struct {
	Method      string                 // HTTP method: "GET", "POST", "PUT", "DELETE", etc.
	Path        string                 // URL path: "/api/resource"
	Handler     http.HandlerFunc       // The actual handler function
	Summary     string                 // Short one-line description
	Description string                 // Detailed description (can be multi-line)
	Tags        []string               // Grouping tags (e.g., ["certificates", "admin"])
	RequestBody interface{}            // Struct type for request body (will be reflected)
	Responses   map[int]interface{}    // Status code → response type (struct or string)
	Security    []string               // Security schemes (e.g., ["mTLS", "Bearer"])
}

// OpenAPISpec represents a minimal OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI string                `json:"openapi"` // "3.0.0"
	Info    OpenAPIInfo           `json:"info"`
	Servers []OpenAPIServer       `json:"servers,omitempty"`
	Paths   map[string]PathItem   `json:"paths"`
	Components *OpenAPIComponents `json:"components,omitempty"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPIServer describes API server
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem describes operations available on a single path
type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

// Operation describes a single API operation
type Operation struct {
	Summary     string                        `json:"summary,omitempty"`
	Description string                        `json:"description,omitempty"`
	Tags        []string                      `json:"tags,omitempty"`
	OperationID string                        `json:"operationId,omitempty"`
	RequestBody *RequestBody                  `json:"requestBody,omitempty"`
	Responses   map[string]Response           `json:"responses"`
	Security    []map[string][]string         `json:"security,omitempty"`
}

// RequestBody describes request body
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Required    bool                  `json:"required,omitempty"`
	Content     map[string]MediaType  `json:"content"`
}

// MediaType describes content type
type MediaType struct {
	Schema  map[string]interface{} `json:"schema"`
	Example interface{}            `json:"example,omitempty"`
}

// Response describes a single response
type Response struct {
	Description string                `json:"description"`
	Content     map[string]MediaType  `json:"content,omitempty"`
}

// OpenAPIComponents contains reusable schemas
type OpenAPIComponents struct {
	Schemas         map[string]map[string]interface{} `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme         `json:"securitySchemes,omitempty"`
}

// SecurityScheme describes authentication method
type SecurityScheme struct {
	Type        string `json:"type"` // "http", "apiKey", "oauth2", "openIdConnect", "mutualTLS"
	Scheme      string `json:"scheme,omitempty"` // "bearer", "basic", etc.
	Description string `json:"description,omitempty"`
}

// APIDescription is our internal format for Introspection
// (similar to OpenAPI but simpler for our needs)
type APIDescription struct {
	ServiceName string        `json:"service_name"`
	Version     string        `json:"version"`
	BaseURL     string        `json:"base_url"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

// APIEndpoint describes a single endpoint (internal format)
type APIEndpoint struct {
	Method      string                    `json:"method"`
	Path        string                    `json:"path"`
	Summary     string                    `json:"summary"`
	Description string                    `json:"description"`
	RequestBody *RequestBodySchema        `json:"request_body,omitempty"`
	Responses   map[string]ResponseSchema `json:"responses"`
	Tags        []string                  `json:"tags,omitempty"`
}

// RequestBodySchema describes request body (internal format)
type RequestBodySchema struct {
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema"`
	Required    bool                   `json:"required"`
	Example     interface{}            `json:"example,omitempty"`
}

// ResponseSchema describes response (internal format)
type ResponseSchema struct {
	Description string                 `json:"description"`
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
}
