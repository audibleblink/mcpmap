package main

import (
	"context"
	"encoding/json"
	"mcpmap/cache"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// loadTestFixture loads a JSON fixture file for testing
func loadTestFixture(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "fixtures", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load test fixture %s: %v", filename, err)
	}
	return data
}

// TestLoadServerDataWithConfig tests the core data loading functionality
func TestLoadServerDataWithConfig(t *testing.T) {
	// This test would require a mock server, so we'll test the cache logic instead
	ctx := context.Background()

	// Test with invalid server (should fail gracefully)
	_, err := loadServerDataWithConfig(ctx, "invalid://server", "http", "", "test-client")
	if err == nil {
		t.Error("Expected error for invalid server URL")
	}
}

// TestDisplayData tests the data display logic with real fixture data
func TestDisplayData(t *testing.T) {
	// Load fixture data
	fixtureData := loadTestFixture(t, "list_tools.json")

	var tools []mcp.Tool
	if err := json.Unmarshal(fixtureData, &tools); err != nil {
		t.Fatalf("Failed to unmarshal fixture data: %v", err)
	}

	// Create cache data with proper pointer types
	toolPtrs := make([]*mcp.Tool, len(tools))
	for i := range tools {
		toolPtrs[i] = &tools[i]
	}

	cacheData := &cache.CacheData{
		Tools:     toolPtrs,
		Resources: []*mcp.Resource{},
		Prompts:   []*mcp.Prompt{},
	}

	tests := []struct {
		name     string
		listType string
		wantErr  bool
	}{
		{"list tools", "tools", false},
		{"list resources", "resources", false},
		{"list prompts", "prompts", false},
		{"list all", "all", false},
		{"invalid type", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := displayData(cacheData, tt.listType)
			if (err != nil) != tt.wantErr {
				t.Errorf("displayData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseParamsWithSchema tests parameter parsing with real schema
func TestParseParamsWithSchema(t *testing.T) {
	// Load fixture data to get real schema
	fixtureData := loadTestFixture(t, "list_tools.json")

	var tools []mcp.Tool
	if err := json.Unmarshal(fixtureData, &tools); err != nil {
		t.Fatalf("Failed to unmarshal fixture data: %v", err)
	}

	// Extract schema from the first tool
	if len(tools) == 0 {
		t.Skip("No tools in fixture data")
	}

	schema, err := extractFullSchema(tools[0].InputSchema)
	if err != nil {
		t.Fatalf("Failed to extract schema: %v", err)
	}

	tests := []struct {
		name    string
		params  []string
		wantErr bool
	}{
		{
			name:    "valid string parameter",
			params:  []string{"libraryName=react"},
			wantErr: false,
		},
		{
			name:    "missing required parameter",
			params:  []string{},
			wantErr: true,
		},
		{
			name:    "invalid parameter format",
			params:  []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseParamsWithSchema(tt.params, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParamsWithSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConvertValueWithRealSchema tests type conversion with real schema data
func TestConvertValueWithRealSchema(t *testing.T) {
	// Load fixture data
	fixtureData := loadTestFixture(t, "list_tools.json")

	var tools []mcp.Tool
	if err := json.Unmarshal(fixtureData, &tools); err != nil {
		t.Fatalf("Failed to unmarshal fixture data: %v", err)
	}

	// Find a tool with number parameter (get-library-docs has tokens parameter)
	var numberSchema *ParameterSchema
	for _, tool := range tools {
		if tool.Name == "get-library-docs" {
			schema, err := extractFullSchema(tool.InputSchema)
			if err != nil {
				continue
			}
			if tokensParam, exists := schema.Parameters["tokens"]; exists {
				numberSchema = tokensParam
				break
			}
		}
	}

	if numberSchema == nil {
		t.Skip("No number parameter found in fixture data")
	}

	tests := []struct {
		name     string
		value    string
		expected any
		wantErr  bool
	}{
		{"valid number", "1000", float64(1000), false},
		{"invalid number", "abc", nil, true},
	}

	// Only test default if it exists
	if numberSchema.Default != nil {
		tests = append(tests, struct {
			name     string
			value    string
			expected any
			wantErr  bool
		}{"empty with default", "", numberSchema.Default, false})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertValue(tt.value, numberSchema)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("convertValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestValidateFormat tests format validation
func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		format  string
		wantErr bool
	}{
		{"valid email", "user@example.com", "email", false},
		{"invalid email", "invalid", "email", true},
		{"valid url", "https://example.com", "url", false},
		{"invalid url", "not-a-url", "url", true},
		{"valid date-time", "2024-01-01T12:00:00Z", "date-time", false},
		{"invalid date-time", "invalid", "date-time", true},
		{"unknown format", "value", "unknown", false}, // Should not error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormat(tt.value, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetFormatHint tests format hint generation
func TestGetFormatHint(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"email", "Use email format: user@example.com"},
		{"url", "Use URL format: https://example.com"},
		{"date-time", "Use ISO 8601 format: 2024-01-01T12:00:00Z"},
		{"unknown", "Must match format: unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := getFormatHint(tt.format)
			if result != tt.expected {
				t.Errorf("getFormatHint() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestValidateEnum tests enum validation
func TestValidateEnum(t *testing.T) {
	enum := []any{"option1", "option2", 42}

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid string", "option1", false},
		{"valid number", 42, false},
		{"invalid value", "invalid", true},
		{"wrong type", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnum(tt.value, enum)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConvertObject tests object conversion
func TestConvertObject(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "object"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid object", `{"key":"value"}`, false},
		{"invalid json", `{invalid}`, true},
		{"empty object", `{}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertObject(tt.value, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConvertNull tests null conversion
func TestConvertNull(t *testing.T) {
	schema := &ParameterSchema{Name: "test", Type: "null"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty string", "", false},
		{"null string", "null", false},
		{"invalid value", "not-null", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertNull(tt.value, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertNull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != nil {
				t.Errorf("convertNull() = %v, want nil", result)
			}
		})
	}
}

// TestValidateRequired tests required parameter validation
func TestValidateRequired(t *testing.T) {
	schema := &ToolSchema{
		Required: []string{"required1", "required2"},
		Parameters: map[string]*ParameterSchema{
			"required1": {Name: "required1", Required: true},
			"required2": {Name: "required2", Required: true},
			"optional":  {Name: "optional", Required: false},
		},
	}

	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{
			name:    "all required present",
			params:  map[string]any{"required1": "value1", "required2": "value2"},
			wantErr: false,
		},
		{
			name:    "missing required",
			params:  map[string]any{"required1": "value1"},
			wantErr: true,
		},
		{
			name: "with optional",
			params: map[string]any{
				"required1": "value1",
				"required2": "value2",
				"optional":  "value3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequired(tt.params, schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFullServerDataIntegration tests with complete server data including tools, resources, and prompts
func TestFullServerDataIntegration(t *testing.T) {
	// Load comprehensive fixture data
	fixtureData := loadTestFixture(t, "full_server_data.json")

	var serverData struct {
		Tools     []mcp.Tool     `json:"tools"`
		Resources []mcp.Resource `json:"resources"`
		Prompts   []mcp.Prompt   `json:"prompts"`
	}

	if err := json.Unmarshal(fixtureData, &serverData); err != nil {
		t.Fatalf("Failed to unmarshal server data: %v", err)
	}

	// Convert to cache data format
	toolPtrs := make([]*mcp.Tool, len(serverData.Tools))
	for i := range serverData.Tools {
		toolPtrs[i] = &serverData.Tools[i]
	}
	resourcePtrs := make([]*mcp.Resource, len(serverData.Resources))
	for i := range serverData.Resources {
		resourcePtrs[i] = &serverData.Resources[i]
	}
	promptPtrs := make([]*mcp.Prompt, len(serverData.Prompts))
	for i := range serverData.Prompts {
		promptPtrs[i] = &serverData.Prompts[i]
	}

	cacheData := &cache.CacheData{
		Tools:     toolPtrs,
		Resources: resourcePtrs,
		Prompts:   promptPtrs,
	}

	// Test all display types with real data
	tests := []struct {
		name     string
		listType string
		wantErr  bool
	}{
		{"display tools", "tools", false},
		{"display resources", "resources", false},
		{"display prompts", "prompts", false},
		{"display all", "all", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := displayData(cacheData, tt.listType)
			if (err != nil) != tt.wantErr {
				t.Errorf("displayData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
