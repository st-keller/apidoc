package apidoc

import (
	"fmt"
	"reflect"
	"strings"
)

// reflectToJSONSchema converts a Go type to JSON Schema using reflection
func reflectToJSONSchema(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{"type": "object"}
	}

	t := reflect.TypeOf(v)

	// Dereference pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return typeToSchema(t)
}

// typeToSchema converts reflect.Type to JSON Schema
func typeToSchema(t reflect.Type) map[string]interface{} {
	switch t.Kind() {
	case reflect.Struct:
		return structToSchema(t)
	case reflect.Slice, reflect.Array:
		return arrayToSchema(t)
	case reflect.Map:
		return mapToSchema(t)
	case reflect.String:
		return map[string]interface{}{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	default:
		return map[string]interface{}{"type": "object"}
	}
}

// structToSchema converts a struct to JSON Schema
func structToSchema(t reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse json tag (e.g., "field_name,omitempty")
		fieldName := strings.Split(jsonTag, ",")[0]

		// Get openapi tag for metadata
		openAPITag := field.Tag.Get("openapi")
		fieldSchema := typeToSchema(field.Type)

		// Parse openapi tag and enhance schema
		if openAPITag != "" {
			parseOpenAPITag(openAPITag, fieldSchema, &required, fieldName)
		}

		// Add description from doc tag
		if docTag := field.Tag.Get("doc"); docTag != "" {
			fieldSchema["description"] = docTag
		}

		schema["properties"].(map[string]interface{})[fieldName] = fieldSchema
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// arrayToSchema converts array/slice to JSON Schema
func arrayToSchema(t reflect.Type) map[string]interface{} {
	itemType := t.Elem()
	return map[string]interface{}{
		"type":  "array",
		"items": typeToSchema(itemType),
	}
}

// mapToSchema converts map to JSON Schema
func mapToSchema(t reflect.Type) map[string]interface{} {
	valueType := t.Elem()
	return map[string]interface{}{
		"type":                 "object",
		"additionalProperties": typeToSchema(valueType),
	}
}

// parseOpenAPITag parses openapi:"..." tag and enhances schema
// Format: "required,enum=val1|val2,min=1,max=100,default=50,description=..."
func parseOpenAPITag(tag string, schema map[string]interface{}, required *[]string, fieldName string) {
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			*required = append(*required, fieldName)
			continue
		}

		// Handle key=value pairs
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "enum":
			// enum=val1|val2|val3
			values := strings.Split(value, "|")
			schema["enum"] = values

		case "min", "minimum":
			// For numbers: minimum value
			schema["minimum"] = parseNumber(value)

		case "max", "maximum":
			// For numbers: maximum value
			schema["maximum"] = parseNumber(value)

		case "minLength":
			// For strings: minimum length
			schema["minLength"] = parseNumber(value)

		case "maxLength":
			// For strings: maximum length
			schema["maxLength"] = parseNumber(value)

		case "pattern":
			// For strings: regex pattern
			schema["pattern"] = value

		case "default":
			// Default value
			schema["default"] = parseValue(value, schema["type"].(string))

		case "description":
			// Description
			schema["description"] = value

		case "format":
			// Format hint (e.g., "date-time", "email", "uri")
			schema["format"] = value

		case "example":
			// Example value
			schema["example"] = parseValue(value, schema["type"].(string))
		}
	}
}

// parseNumber parses string to appropriate number type
func parseNumber(s string) interface{} {
	// Try parsing as integer first
	if n, err := parseIntHelper(s); err == nil {
		return n
	}

	// Fallback to float
	if f, err := parseFloatHelper(s); err == nil {
		return f
	}

	return s
}

func parseIntHelper(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseFloatHelper(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// parseValue parses string value based on type
func parseValue(s string, typ string) interface{} {
	switch typ {
	case "integer":
		var i int
		return i
	case "number":
		var f float64
		return f
	case "boolean":
		return s == "true"
	default:
		return s
	}
}
