package apidoc

import (
	"net/http"
	"sync"
)

var (
	registry     = &EndpointRegistry{endpoints: make([]EndpointConfig, 0)}
	registryLock sync.RWMutex
)

// EndpointRegistry stores all registered endpoints
type EndpointRegistry struct {
	endpoints []EndpointConfig
	info      OpenAPIInfo
	baseURL   string
}

// SetServiceInfo sets global service metadata
func SetServiceInfo(title, version, description, baseURL string) {
	registryLock.Lock()
	defer registryLock.Unlock()

	registry.info = OpenAPIInfo{
		Title:       title,
		Description: description,
		Version:     version,
	}
	registry.baseURL = baseURL
}

// RegisterEndpoint registers an API endpoint with metadata for OpenAPI generation
// NOTE: This only collects metadata - services must register their own HTTP handlers!
// This prevents documentation drift by keeping API metadata close to implementation.
func RegisterEndpoint(cfg EndpointConfig) {
	registryLock.Lock()
	defer registryLock.Unlock()

	registry.endpoints = append(registry.endpoints, cfg)

	// Handler is stored only for reflection (analyzing request/response types)
	// Services are responsible for registering actual HTTP routes with their router
}

// GetEndpoints returns all registered endpoints (for testing/debugging)
func GetEndpoints() []EndpointConfig {
	registryLock.RLock()
	defer registryLock.RUnlock()

	return append([]EndpointConfig{}, registry.endpoints...)
}

// GenerateOpenAPI generates OpenAPI 3.0 spec from registered endpoints
func GenerateOpenAPI() *OpenAPISpec {
	registryLock.RLock()
	defer registryLock.RUnlock()

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info:    registry.info,
		Servers: []OpenAPIServer{
			{URL: registry.baseURL},
		},
		Paths: make(map[string]PathItem),
		Components: &OpenAPIComponents{
			Schemas:         make(map[string]map[string]interface{}),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// Add common security schemes
	spec.Components.SecuritySchemes["mTLS"] = SecurityScheme{
		Type:        "mutualTLS",
		Description: "Mutual TLS authentication with client certificates",
	}

	spec.Components.SecuritySchemes["Bearer"] = SecurityScheme{
		Type:        "http",
		Scheme:      "bearer",
		Description: "JWT Bearer token authentication (ADR-031: OAuth2 scopes validated against AuthorizationElements)",
	}

	// Convert each endpoint to OpenAPI operation
	for _, endpoint := range registry.endpoints {
		pathItem, ok := spec.Paths[endpoint.Path]
		if !ok {
			pathItem = PathItem{}
		}

		operation := endpointToOperation(endpoint, spec.Components.Schemas)

		// Assign operation to correct method
		switch endpoint.Method {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		}

		spec.Paths[endpoint.Path] = pathItem
	}

	return spec
}

// GenerateAPIDescription generates our internal APIDescription format
func GenerateAPIDescription() *APIDescription {
	registryLock.RLock()
	defer registryLock.RUnlock()

	desc := &APIDescription{
		ServiceName: registry.info.Title,
		Version:     registry.info.Version,
		BaseURL:     registry.baseURL,
		Endpoints:   make([]APIEndpoint, 0, len(registry.endpoints)),
	}

	for _, endpoint := range registry.endpoints {
		apiEndpoint := APIEndpoint{
			Method:      endpoint.Method,
			Path:        endpoint.Path,
			Summary:     endpoint.Summary,
			Description: endpoint.Description,
			Tags:        endpoint.Tags,
			Responses:   make(map[string]ResponseSchema),
		}

		// Convert request body
		if endpoint.RequestBody != nil {
			schema := reflectToJSONSchema(endpoint.RequestBody)
			apiEndpoint.RequestBody = &RequestBodySchema{
				ContentType: "application/json",
				Schema:      schema,
				Required:    true,
			}
		}

		// Convert responses
		for statusCode, responseType := range endpoint.Responses {
			statusStr := http.StatusText(statusCode)
			if statusStr == "" {
				statusStr = "Response"
			}

			respSchema := ResponseSchema{
				Description: statusStr,
				ContentType: "application/json",
			}

			if responseType != nil {
				// Check if it's a string (simple description)
				if desc, ok := responseType.(string); ok {
					respSchema.Description = desc
					respSchema.ContentType = "text/plain"
				} else {
					// It's a struct → reflect it
					respSchema.Schema = reflectToJSONSchema(responseType)
				}
			}

			apiEndpoint.Responses[http.StatusText(statusCode)] = respSchema
		}

		desc.Endpoints = append(desc.Endpoints, apiEndpoint)
	}

	return desc
}

// endpointToOperation converts EndpointConfig to OpenAPI Operation
func endpointToOperation(endpoint EndpointConfig, schemas map[string]map[string]interface{}) *Operation {
	op := &Operation{
		Summary:     endpoint.Summary,
		Description: endpoint.Description,
		Tags:        endpoint.Tags,
		Responses:   make(map[string]Response),
	}

	// Add request body if present
	if endpoint.RequestBody != nil {
		schema := reflectToJSONSchema(endpoint.RequestBody)
		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: schema,
				},
			},
		}
	}

	// Add responses
	for statusCode, responseType := range endpoint.Responses {
		statusStr := http.StatusText(statusCode)
		if statusStr == "" {
			statusStr = "Response"
		}

		response := Response{
			Description: statusStr,
		}

		if responseType != nil {
			// Check if it's a string (simple description)
			if desc, ok := responseType.(string); ok {
				response.Description = desc
			} else {
				// It's a struct → reflect it
				schema := reflectToJSONSchema(responseType)
				response.Content = map[string]MediaType{
					"application/json": {
						Schema: schema,
					},
				}
			}
		}

		op.Responses[http.StatusText(statusCode)] = response
	}

	// Add security if specified
	if len(endpoint.Security) > 0 {
		op.Security = make([]map[string][]string, len(endpoint.Security))
		for i, scheme := range endpoint.Security {
			op.Security[i] = map[string][]string{
				scheme: {},
			}
		}
	}

	return op
}
