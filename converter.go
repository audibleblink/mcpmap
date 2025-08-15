// converter.go - Type conversion functions for MCP tool parameters
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// TypeConversionError represents an error during type conversion
type TypeConversionError struct {
	Parameter    string
	ExpectedType string
	ActualValue  string
	Hint         string
}

func (e TypeConversionError) Error() string {
	return fmt.Sprintf("parameter %q (type: %s): cannot convert %q\nHint: %s",
		e.Parameter, e.ExpectedType, e.ActualValue, e.Hint)
}

// Error hints for different types
var errorHints = map[string]string{
	"integer": "Use whole numbers like 42 or -10",
	"number":  "Use numbers like 3.14, -0.5, or 42",
	"boolean": "Use true/false, yes/no, 1/0, or on/off",
	"array":   "Use JSON format [1,2,3] or comma-separated: a,b,c",
	"object":  "Use JSON format: {\"key\":\"value\"}",
	"string":  "Use any text value",
}

// Boolean value mappings
var (
	booleanTrueValues  = []string{"true", "yes", "1", "on"}
	booleanFalseValues = []string{"false", "no", "0", "off"}
)

func getConverters() map[string]func(string, *ParameterSchema) (any, error) {
	return map[string]func(string, *ParameterSchema) (any, error){
		"string":  convertString,
		"integer": convertInteger,
		"number":  convertNumber,
		"boolean": convertBoolean,
		"array":   convertArray,
		"object":  convertObject,
		"null":    convertNull,
	}
}

func newTypeError(
	schema *ParameterSchema,
	expectedType, actualValue, hint string,
) TypeConversionError {
	return TypeConversionError{
		Parameter:    schema.Name,
		ExpectedType: expectedType,
		ActualValue:  actualValue,
		Hint:         hint,
	}
}

// convertValue converts a string value to the appropriate type based on schema
func convertValue(value string, schema *ParameterSchema) (any, error) {
	if schema == nil {
		return value, nil
	}

	// Handle empty string with default value
	if value == "" && schema.Default != nil {
		return schema.Default, nil
	}

	if conv, ok := getConverters()[schema.Type]; ok {
		return conv(value, schema)
	}

	// Unknown type, treat as string
	return value, nil
}

// convertString handles string type conversion with format validation
func convertString(value string, schema *ParameterSchema) (any, error) {
	// Remove surrounding quotes if present
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	// Validate enum if specified
	if len(schema.Enum) > 0 {
		if err := validateEnum(value, schema.Enum); err != nil {
			return nil, newTypeError(
				schema,
				fmt.Sprintf("enum %v", schema.Enum),
				value,
				fmt.Sprintf("Must be one of: %v", schema.Enum),
			)
		}
	}

	// Validate format if specified
	if schema.Format != "" {
		if err := validateFormat(value, schema.Format); err != nil {
			return nil, newTypeError(
				schema,
				fmt.Sprintf("string (format: %s)", schema.Format),
				value,
				getFormatHint(schema.Format),
			)
		}
	}

	return value, nil
}

// convertInteger converts string to integer
func convertInteger(value string, schema *ParameterSchema) (any, error) {
	value = strings.TrimSpace(value)

	// Check if it contains decimal point
	if strings.Contains(value, ".") {
		return nil, newTypeError(schema, "integer", value, errorHints["integer"])
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, newTypeError(schema, "integer", value, errorHints["integer"])
	}

	// Validate enum if specified
	if len(schema.Enum) > 0 {
		if err := validateEnum(result, schema.Enum); err != nil {
			return nil, newTypeError(
				schema,
				fmt.Sprintf("integer enum %v", schema.Enum),
				value,
				fmt.Sprintf("Must be one of: %v", schema.Enum),
			)
		}
	}

	return result, nil
}

// convertNumber converts string to float64
func convertNumber(value string, schema *ParameterSchema) (any, error) {
	value = strings.TrimSpace(value)

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, newTypeError(schema, "number", value, errorHints["number"])
	}

	// Validate enum if specified
	if len(schema.Enum) > 0 {
		if err := validateEnum(result, schema.Enum); err != nil {
			return nil, newTypeError(
				schema,
				fmt.Sprintf("number enum %v", schema.Enum),
				value,
				fmt.Sprintf("Must be one of: %v", schema.Enum),
			)
		}
	}

	return result, nil
}

// convertBoolean converts string to boolean using multiple formats
func convertBoolean(value string, schema *ParameterSchema) (any, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if slices.Contains(booleanTrueValues, value) {
		return true, nil
	}

	if slices.Contains(booleanFalseValues, value) {
		return false, nil
	}

	return nil, newTypeError(schema, "boolean", value, errorHints["boolean"])
}

// convertArray converts string to array, supporting both JSON and CSV formats
func convertArray(value string, schema *ParameterSchema) (any, error) {
	value = strings.TrimSpace(value)

	// Try JSON format first
	if isJSONArray(value) {
		var result []any
		if err := json.Unmarshal([]byte(value), &result); err != nil {
			return nil, newTypeError(
				schema,
				"array",
				value,
				"Invalid JSON array format. "+errorHints["array"],
			)
		}

		// Convert array items if schema is provided
		if schema.Items != nil {
			convertedResult := make([]any, len(result))
			for i, item := range result {
				// Convert item to string first, then apply schema conversion
				itemStr := fmt.Sprintf("%v", item)
				converted, err := convertValue(itemStr, schema.Items)
				if err != nil {
					return nil, fmt.Errorf("array item %d: %w", i, err)
				}
				convertedResult[i] = converted
			}
			return convertedResult, nil
		}

		return result, nil
	}

	// Try comma-separated format
	if value == "" {
		return []any{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]any, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		// Convert item if schema is provided
		if schema.Items != nil {
			converted, err := convertValue(part, schema.Items)
			if err != nil {
				return nil, fmt.Errorf("array item %d: %w", i, err)
			}
			result[i] = converted
		} else {
			result[i] = part
		}
	}

	return result, nil
}

// convertObject converts JSON string to object
func convertObject(value string, schema *ParameterSchema) (any, error) {
	value = strings.TrimSpace(value)

	var result map[string]any
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, newTypeError(schema, "object", value, errorHints["object"])
	}

	// Convert object properties if schema is provided
	if schema.Properties != nil {
		convertedResult := make(map[string]any)
		for key, val := range result {
			if propSchema, exists := schema.Properties[key]; exists {
				// Convert value to string first, then apply schema conversion
				valStr := fmt.Sprintf("%v", val)
				converted, err := convertValue(valStr, propSchema)
				if err != nil {
					return nil, fmt.Errorf("object property %q: %w", key, err)
				}
				convertedResult[key] = converted
			} else {
				convertedResult[key] = val
			}
		}
		return convertedResult, nil
	}

	return result, nil
}

// convertNull handles null values
func convertNull(value string, schema *ParameterSchema) (any, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" {
		return nil, nil
	}

	return nil, newTypeError(schema, "null", value, "Use empty string or 'null'")
}

// isJSONArray checks if a string looks like a JSON array
func isJSONArray(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

// validateEnum checks if a value is in the allowed enum values
func validateEnum(value any, enum []any) error {
	if slices.Contains(enum, value) {
		return nil
	}
	return fmt.Errorf("value not in enum")
}

// validateFormat validates string format (basic implementation)
func validateFormat(value, format string) error {
	switch format {
	case "email":
		if !strings.Contains(value, "@") {
			return fmt.Errorf("invalid email format")
		}
	case "uri", "url":
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("invalid URL format")
		}
	case "date-time":
		// Basic check for ISO 8601 format
		if !strings.Contains(value, "T") && !strings.Contains(value, "-") {
			return fmt.Errorf("invalid date-time format")
		}
	}
	return nil
}

// getFormatHint returns helpful hints for format validation
func getFormatHint(format string) string {
	switch format {
	case "email":
		return "Use email format: user@example.com"
	case "uri", "url":
		return "Use URL format: https://example.com"
	case "date-time":
		return "Use ISO 8601 format: 2024-01-01T12:00:00Z"
	default:
		return fmt.Sprintf("Must match format: %s", format)
	}
}

// parseParamsWithSchema parses parameters using schema-based type conversion
func parseParamsWithSchema(params []string, schema *ToolSchema) (map[string]any, error) {
	result := make(map[string]any)
	var warnings []string

	// Parse all parameters
	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format '%s', expected name=value", param)
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if name == "" {
			return nil, fmt.Errorf("parameter name cannot be empty in '%s'", param)
		}

		// Get parameter schema
		paramSchema, exists := schema.Parameters[name]
		if !exists {
			// Parameter not in schema - warn but continue
			warnings = append(warnings, fmt.Sprintf("parameter %q not found in schema", name))
			result[name] = value
			continue
		}

		// Convert value using schema
		converted, err := convertValue(value, paramSchema)
		if err != nil {
			return nil, err
		}

		result[name] = converted
	}

	// Print warnings to stderr
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}

	// Validate required parameters
	if err := validateRequired(result, schema); err != nil {
		return nil, err
	}

	return result, nil
}

// ensure verbose flag propagates to schema extraction logic
