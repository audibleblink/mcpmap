package main

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListArgumentValidation(t *testing.T) {
	validTypes := []string{"tools", "resources", "prompts", "all", ""}

	for _, listType := range validTypes {
		t.Run("valid-"+listType, func(t *testing.T) {
			// These should all be valid
			switch listType {
			case "tools", "resources", "prompts", "all", "":
				// Valid - no error expected
			default:
				t.Errorf("unexpected invalid type in valid list: %s", listType)
			}
		})
	}

	// Test invalid type
	t.Run("invalid-type", func(t *testing.T) {
		// In actual usage, this would be caught by cobra args validation
		listType := "invalid"
		validTypes := map[string]bool{"tools": true, "resources": true, "prompts": true, "all": true}
		if !validTypes[listType] && listType != "" {
			// This would error in real usage - expected behavior
		}
	})
}

func TestOutputItems(t *testing.T) {
	h := newTestHelper(t)

	tests := []struct {
		name       string
		items      []any
		prefix     string
		jsonOutput bool
		contains   []string
	}{
		{
			name:     "text output tools",
			items:    []any{&mcp.Tool{Name: "tool1"}, &mcp.Tool{Name: "tool2"}},
			prefix:   "tool",
			contains: []string{"tool:tool1", "tool:tool2"},
		},
		{
			name:     "text output resources",
			items:    []any{&mcp.Resource{URI: "file://test.txt"}},
			prefix:   "resource",
			contains: []string{"resource:file://test.txt"},
		},
		{
			name:       "json output",
			items:      []any{&mcp.Tool{Name: "tool1", Description: "Test tool"}},
			prefix:     "tool",
			jsonOutput: true,
			contains:   []string{`"name":"tool1"`, `"description":"Test tool"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonOutput = tt.jsonOutput

			output := h.captureOutput(func() {
				outputItems(tt.items, tt.prefix)
			})

			h.assertStringContains(output, tt.contains)
		})
	}
}

func TestGetItemName(t *testing.T) {
	tests := []struct {
		name string
		item any
		want string
	}{
		{"tool", &mcp.Tool{Name: "test-tool"}, "test-tool"},
		{"resource", &mcp.Resource{URI: "file://test.txt"}, "file://test.txt"},
		{"prompt", &mcp.Prompt{Name: "test-prompt"}, "test-prompt"},
		{"unknown", "string", "unknown"},
		{"nil", nil, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getItemName(tt.item); got != tt.want {
				t.Errorf("getItemName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListCommandConfiguration(t *testing.T) {
	// Test basic command setup
	if listCmd.Use != "list [resources|tools|prompts]" {
		t.Errorf("unexpected list command Use: %q", listCmd.Use)
	}

	// Check json flag exists
	if listCmd.Flags().Lookup("json") == nil {
		t.Error("json flag not found")
	}
}

func TestJSONMarshaling(t *testing.T) {
	items := []any{
		&mcp.Tool{Name: "test-tool", Description: "A test tool"},
		&mcp.Resource{URI: "file://test.txt", Name: "Test Resource"},
		&mcp.Prompt{Name: "test-prompt", Description: "A test prompt"},
	}

	for _, item := range items {
		t.Run("marshal-"+getItemName(item), func(t *testing.T) {
			jsonBytes, err := json.Marshal(item)
			if err != nil {
				t.Errorf("failed to marshal item to JSON: %v", err)
			}

			// Verify it's valid JSON
			var result map[string]any
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Errorf("generated JSON is invalid: %v", err)
			}
		})
	}
}
