// schema.go - Schema-based type conversion for MCP tool parameters
package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// ParameterSchema represents the schema for a single parameter
type ParameterSchema struct {
	Name        string                      `json:"name"`
	Type        string                      `json:"type"`
	Required    bool                        `json:"required"`
	Default     any                         `json:"default,omitempty"`
	Enum        []any                       `json:"enum,omitempty"`
	Format      string                      `json:"format,omitempty"`
	Items       *ParameterSchema            `json:"items,omitempty"`      // For arrays
	Properties  map[string]*ParameterSchema `json:"properties,omitempty"` // For objects
	Description string                      `json:"description,omitempty"`
}

// ToolSchema represents the complete schema for a tool
type ToolSchema struct {
	Parameters map[string]*ParameterSchema `json:"parameters"`
	Required   []string                    `json:"required"`
}

// extractFullSchema extracts a complete tool schema from the MCP tool schema
func extractFullSchema(schema any) (*ToolSchema, error) {
	if schema == nil {
		return &ToolSchema{
			Parameters: make(map[string]*ParameterSchema),
			Required:   []string{},
		}, nil
	}

	if jsonSchema, ok := schema.(*jsonschema.Schema); ok {
		toolSchema := &ToolSchema{
			Parameters: make(map[string]*ParameterSchema),
			Required:   jsonSchema.Required,
		}

		for name, propSchema := range jsonSchema.Properties {
			paramSchema := extractParameterSchemaFromJSON(name, propSchema, jsonSchema.Required)
			toolSchema.Parameters[name] = paramSchema
		}

		return toolSchema, nil
	}

	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema is not a valid object")
	}

	toolSchema := &ToolSchema{
		Parameters: make(map[string]*ParameterSchema),
		Required:   []string{},
	}

	if requiredField, exists := schemaMap["required"]; exists {
		if requiredSlice, ok := requiredField.([]any); ok {
			for _, req := range requiredSlice {
				if reqStr, ok := req.(string); ok {
					toolSchema.Required = append(toolSchema.Required, reqStr)
				}
			}
		}
	}

	if propertiesField, exists := schemaMap["properties"]; exists {
		if propertiesMap, ok := propertiesField.(map[string]any); ok {
			for name, propSchema := range propertiesMap {
				paramSchema := extractParameterSchema(name, propSchema, toolSchema.Required)
				toolSchema.Parameters[name] = paramSchema
			}
		}
	}

	return toolSchema, nil
}

type schemaData struct {
	Type        string
	Description string
	Format      string
	Default     any
	Enum        []any
}

func buildParameterSchema(name string, data schemaData, required []string) *ParameterSchema {
	return &ParameterSchema{
		Name:        name,
		Type:        data.Type,
		Required:    contains(required, name),
		Description: data.Description,
		Default:     data.Default,
		Enum:        data.Enum,
		Format:      data.Format,
	}
}

func extractComplexTypes(param *ParameterSchema, schemaMap map[string]any) {
	if param.Type == "array" {
		if itemsField, exists := schemaMap["items"]; exists {
			param.Items = extractParameterSchema("", itemsField, []string{})
		}
	}

	if param.Type == "object" {
		if propertiesField, exists := schemaMap["properties"]; exists {
			if propertiesMap, ok := propertiesField.(map[string]any); ok {
				param.Properties = make(map[string]*ParameterSchema)
				for propName, propSchema := range propertiesMap {
					param.Properties[propName] = extractParameterSchema(
						propName,
						propSchema,
						[]string{},
					)
				}
			}
		}
	}
}

func extractComplexTypesFromJSON(param *ParameterSchema, schema *jsonschema.Schema) {
	if param.Type == "array" && schema.Items != nil {
		param.Items = extractParameterSchemaFromJSON("", schema.Items, []string{})
	}

	if param.Type == "object" && len(schema.Properties) > 0 {
		param.Properties = make(map[string]*ParameterSchema)
		for propName, propSchema := range schema.Properties {
			param.Properties[propName] = extractParameterSchemaFromJSON(
				propName,
				propSchema,
				[]string{},
			)
		}
	}
}

func extractParameterSchema(name string, schema any, required []string) *ParameterSchema {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return buildParameterSchema(name, schemaData{Type: "string"}, required)
	}

	data := schemaData{
		Type: "string",
	}

	if typeField, exists := schemaMap["type"]; exists {
		if t, ok := typeField.(string); ok {
			data.Type = t
		}
	}

	if descField, exists := schemaMap["description"]; exists {
		if d, ok := descField.(string); ok {
			data.Description = d
		}
	}

	if formatField, exists := schemaMap["format"]; exists {
		if f, ok := formatField.(string); ok {
			data.Format = f
		}
	}

	if defaultField, exists := schemaMap["default"]; exists {
		data.Default = defaultField
	}

	if enumField, exists := schemaMap["enum"]; exists {
		if e, ok := enumField.([]any); ok {
			data.Enum = e
		}
	}

	param := buildParameterSchema(name, data, required)
	extractComplexTypes(param, schemaMap)

	return param
}

// extractParameterSchemaFromJSON extracts schema from *jsonschema.Schema
func extractParameterSchemaFromJSON(
	name string,
	schema *jsonschema.Schema,
	required []string,
) *ParameterSchema {
	if schema == nil {
		return buildParameterSchema(name, schemaData{Type: "string"}, required)
	}

	var defaultVal any
	if len(schema.Default) > 0 {
		json.Unmarshal(schema.Default, &defaultVal)
	}

	data := schemaData{
		Type:        schema.Type,
		Description: schema.Description,
		Format:      schema.Format,
		Default:     defaultVal,
		Enum:        schema.Enum,
	}

	param := buildParameterSchema(name, data, required)
	extractComplexTypesFromJSON(param, schema)

	return param
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// validateRequired checks if all required parameters are present
func validateRequired(params map[string]any, schema *ToolSchema) error {
	var missing []string

	for _, required := range schema.Required {
		if _, exists := params[required]; !exists {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required parameters: %v", missing)
	}

	return nil
}
