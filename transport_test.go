package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateTransport(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		wantErr   bool
		wantType  string
	}{
		{"sse transport", "sse", "http://localhost:3000", false, "*mcp.SSEClientTransport"},
		{"http transport", "http", "http://localhost:8080", false, "*mcp.StreamableClientTransport"},
		{"streamable transport", "streamable", "http://localhost:8080", false, "*mcp.StreamableClientTransport"},
		{"case insensitive", "SSE", "http://localhost:3000", false, "*mcp.SSEClientTransport"},
		{"invalid transport", "invalid", "http://localhost:3000", true, ""},
		{"empty transport", "", "http://localhost:3000", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, "")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if transport == nil {
					t.Error("expected non-nil transport")
				}
				if tt.wantType != "" {
					if got := reflect.TypeOf(transport).String(); got != tt.wantType {
						t.Errorf("expected type %q, got %q", tt.wantType, got)
					}
				}
			}
		})
	}
}

func TestCreateTransportWithProxy(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		proxy     string
		wantErr   bool
	}{
		{"http with valid proxy", "http", "http://localhost:8080", "http://proxy.example.com:8080", false},
		{"http with invalid proxy", "http", "http://localhost:8080", "://invalid-url", true},
		{"sse with proxy (ignored)", "sse", "http://localhost:3000", "http://proxy.example.com:8080", false},
		{"http without proxy", "http", "http://localhost:8080", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, tt.proxy)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if transport == nil {
					t.Error("expected non-nil transport")
				}
			}
		})
	}
}

func TestProxyConfiguration(t *testing.T) {
	// Test that proxy is properly configured in HTTP client
	transport, err := createTransport("http", "http://localhost:8080", "http://proxy.example.com:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// This test verifies that the transport was created successfully with proxy
	// The actual proxy functionality would be tested in integration tests
	if transport == nil {
		t.Error("expected non-nil transport")
	}
}

func TestCreateSessionFailureScenarios(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		timeout   time.Duration
	}{
		{"invalid transport", "invalid", "http://localhost:3000", time.Second},
		{"malformed URL", "sse", "not-a-url", time.Second},
		{"unreachable server", "sse", "http://localhost:99999", time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			session, err := createSession(ctx, tt.transport, tt.url, "")
			if err == nil {
				t.Error("expected error but got none")
			}
			if session != nil {
				session.Close()
				t.Error("expected nil session on error")
			}
		})
	}
}

// Simple mock for testing parameter extraction
type mockSession struct {
	tools []*mcp.Tool
	err   error
}

func (m *mockSession) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mcp.ListToolsResult{Tools: m.tools}, nil
}

func (m *mockSession) Close() error { return nil }

func TestExtractParametersFromSchema(t *testing.T) {
	tests := []struct {
		name   string
		schema any
		want   []ParameterInfo
	}{
		{"nil schema", nil, []ParameterInfo{}},
		{"empty schema", map[string]any{}, []ParameterInfo{}},
		{
			"simple parameters",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "number"},
				},
			},
			[]ParameterInfo{
				{Name: "name", Type: "string"},
				{Name: "age", Type: "number"},
			},
		},
		{
			"missing type defaults to string",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"param": map[string]any{"description": "test"},
				},
			},
			[]ParameterInfo{{Name: "param", Type: "string"}},
		},
		{"invalid schema", "not-a-map", []ParameterInfo{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractParametersFromSchema(tt.schema)

			if len(got) != len(tt.want) {
				t.Errorf("expected %d parameters, got %d", len(tt.want), len(got))
				return
			}

			// Convert to maps for comparison (order doesn't matter)
			gotMap := make(map[string]string)
			for _, p := range got {
				gotMap[p.Name] = p.Type
			}

			wantMap := make(map[string]string)
			for _, p := range tt.want {
				wantMap[p.Name] = p.Type
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("expected %v, got %v", wantMap, gotMap)
			}
		})
	}
}
