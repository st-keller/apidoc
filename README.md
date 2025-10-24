# apidoc

  **Zero-drift API documentation library for Go microservices**

  A lightweight Go library that prevents documentation drift by generating OpenAPI specs and authorization metadata directly from code.

  ## Features

  - 🔄 **Zero Documentation Drift** - API docs generated from actual endpoint definitions
  - 📝 **OpenAPI 3.0 Generation** - Automatic spec generation with reflection
  - 🔐 **Authorization Metadata** - Built-in support for ZITADEL OAuth2 scopes
  - 🎯 **Type-Safe** - Leverage Go's type system for request/response schemas
  - 🚀 **Zero Dependencies** - Pure Go, no external dependencies

  ## Installation

  ```bash
  go get github.com/st-keller/apidoc

  Quick Start

  package main

  import "github.com/st-keller/apidoc"

  func main() {
      // Set service metadata
      apidoc.SetServiceInfo(
          "my-service",
          "1.0.0",
          "My awesome service",
          "https://localhost:8080",
      )

      // Register endpoints
      apidoc.RegisterEndpoint(apidoc.EndpointConfig{
          Method:      "POST",
          Path:        "/api/v1/users",
          Summary:     "Create User",
          Description: "Creates a new user account",
          Tags:        []string{"users"},
          Security:    []string{"Bearer"},
          RequestBody: CreateUserRequest{},
          Responses: map[int]interface{}{
              201: CreateUserResponse{},
              400: "Invalid request",
          },
      })

      // Generate OpenAPI spec
      spec := apidoc.GenerateOpenAPI()

      // Or generate internal API description format
      apiDesc := apidoc.GenerateAPIDescription()
  }

  Why This Library?

  Traditional API documentation approaches suffer from documentation drift - when code changes but docs don't get updated. This library solves that by:

  1. Single Source of Truth - Document endpoints where you define them
  2. Reflection-Based - Automatically generates schemas from Go types
  3. Co-located - Docs live next to the code they document
  4. Compile-Time Safety - Broken references = compilation errors
