package main

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
