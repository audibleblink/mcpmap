package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCreateTransport(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		proxy     string
		token     string
		client    string
		wantErr   bool
		wantType  string
	}{
		{
			name:      "sse transport",
			transport: "sse",
			url:       "http://localhost:3000",
			client:    "test-client",
			wantErr:   false,
			wantType:  "*mcp.SSEClientTransport",
		},
		{
			name:      "http transport",
			transport: "http",
			url:       "http://localhost:8080",
			client:    "test-client",
			wantErr:   false,
			wantType:  "*mcp.StreamableClientTransport",
		},
		{
			name:      "streamable transport",
			transport: "streamable",
			url:       "http://localhost:8080",
			client:    "test-client",
			wantErr:   false,
			wantType:  "*mcp.StreamableClientTransport",
		},
		{
			name:      "streamable-http transport",
			transport: "streamable-http",
			url:       "http://localhost:8080",
			client:    "test-client",
			wantErr:   false,
			wantType:  "*mcp.StreamableClientTransport",
		},
		{
			name:      "case insensitive",
			transport: "SSE",
			url:       "http://localhost:3000",
			client:    "test-client",
			wantErr:   false,
			wantType:  "*mcp.SSEClientTransport",
		},
		{
			name:      "invalid transport",
			transport: "invalid",
			url:       "http://localhost:3000",
			client:    "test-client",
			wantErr:   true,
		},
		{
			name:      "empty transport",
			transport: "",
			url:       "http://localhost:3000",
			client:    "test-client",
			wantErr:   true,
		},
		{
			name:      "empty client name",
			transport: "sse",
			url:       "http://localhost:3000",
			client:    "",
			wantErr:   false,
			wantType:  "*mcp.SSEClientTransport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, tt.proxy, tt.token, tt.client)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if transport == nil {
				t.Error("expected non-nil transport")
				return
			}

			if tt.wantType != "" {
				if got := reflect.TypeOf(transport).String(); got != tt.wantType {
					t.Errorf("expected type %q, got %q", tt.wantType, got)
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
		errMsg    string
	}{
		{
			name:      "http with valid proxy",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "http://proxy.example.com:8080",
			wantErr:   false,
		},
		{
			name:      "http with invalid proxy URL",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "://invalid-url",
			wantErr:   true,
			errMsg:    "invalid proxy URL",
		},
		{
			name:      "http with malformed proxy",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "not-a-url",
			wantErr:   false, // url.Parse actually accepts this
		},
		{
			name:      "sse with valid proxy",
			transport: "sse",
			url:       "http://localhost:3000",
			proxy:     "http://proxy.example.com:8080",
			wantErr:   false,
		},
		{
			name:      "sse with invalid proxy",
			transport: "sse",
			url:       "http://localhost:3000",
			proxy:     "://invalid-url",
			wantErr:   true,
			errMsg:    "invalid proxy URL",
		},
		{
			name:      "http without proxy",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, tt.proxy, "", "test-client")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if transport == nil {
				t.Error("expected non-nil transport")
			}
		})
	}
}

func TestAuthenticationConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		token     string
		wantErr   bool
	}{
		{
			name:      "http with auth token",
			transport: "http",
			url:       "http://localhost:8080",
			token:     "test-token",
			wantErr:   false,
		},
		{
			name:      "sse with auth token",
			transport: "sse",
			url:       "http://localhost:3000",
			token:     "test-token",
			wantErr:   false,
		},
		{
			name:      "http without auth token",
			transport: "http",
			url:       "http://localhost:8080",
			token:     "",
			wantErr:   false,
		},
		{
			name:      "http with empty token",
			transport: "http",
			url:       "http://localhost:8080",
			token:     "   ",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, "", tt.token, "test-client")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if transport == nil {
				t.Error("expected non-nil transport")
			}
		})
	}
}

func TestProxyAndAuthenticationTogether(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		proxy     string
		token     string
		wantErr   bool
	}{
		{
			name:      "http with both proxy and auth",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "http://proxy.example.com:8080",
			token:     "test-token",
			wantErr:   false,
		},
		{
			name:      "sse with both proxy and auth",
			transport: "sse",
			url:       "http://localhost:3000",
			proxy:     "http://proxy.example.com:8080",
			token:     "test-token",
			wantErr:   false,
		},
		{
			name:      "http with invalid proxy but valid auth",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "://invalid",
			token:     "test-token",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := createTransport(tt.transport, tt.url, tt.proxy, tt.token, "test-client")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if transport != nil {
					t.Error("expected nil transport on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if transport == nil {
				t.Error("expected non-nil transport")
			}
		})
	}
}

func TestAuthTransportRoundTrip(t *testing.T) {
	// Create a test server to verify auth headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create auth transport
	baseTransport := &http.Transport{}
	authTrans := &authTransport{
		base:  baseTransport,
		token: "test-token",
	}

	// Create request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Make request through auth transport
	client := &http.Client{Transport: authTrans}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateSessionFailureScenarios(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		url       string
		proxy     string
		token     string
		timeout   time.Duration
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "invalid transport",
			transport: "invalid",
			url:       "http://localhost:3000",
			timeout:   time.Second,
			wantErr:   true,
			errMsg:    "unknown transport type",
		},
		{
			name:      "invalid proxy URL",
			transport: "http",
			url:       "http://localhost:8080",
			proxy:     "://invalid",
			timeout:   time.Second,
			wantErr:   true,
			errMsg:    "invalid proxy URL",
		},
		{
			name:      "unreachable server with short timeout",
			transport: "sse",
			url:       "http://localhost:99999",
			timeout:   100 * time.Millisecond,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			session, err := createSession(ctx, tt.transport, tt.url, tt.proxy, tt.token, "test-client")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				if session != nil {
					session.Close()
					t.Error("expected nil session on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if session != nil {
				session.Close()
			}
		})
	}
}

// Mock session is not needed for these tests since getToolParameters
// requires a real *mcp.ClientSession. Parameter extraction is tested
// separately via extractParametersFromSchema.

func TestExtractParametersFromSchema(t *testing.T) {
	tests := []struct {
		name   string
		schema any
		want   []ParameterInfo
	}{
		{
			name:   "nil schema",
			schema: nil,
			want:   []ParameterInfo{},
		},
		{
			name:   "empty schema",
			schema: map[string]any{},
			want:   []ParameterInfo{},
		},
		{
			name: "simple parameters",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
					"age":  map[string]any{"type": "number"},
				},
			},
			want: []ParameterInfo{
				{Name: "name", Type: "string"},
				{Name: "age", Type: "number"},
			},
		},
		{
			name: "missing type defaults to string",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"param": map[string]any{"description": "test"},
				},
			},
			want: []ParameterInfo{{Name: "param", Type: "string"}},
		},
		{
			name: "complex types",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"items":   map[string]any{"type": "array"},
					"config":  map[string]any{"type": "object"},
					"enabled": map[string]any{"type": "boolean"},
					"count":   map[string]any{"type": "integer"},
				},
			},
			want: []ParameterInfo{
				{Name: "items", Type: "array"},
				{Name: "config", Type: "object"},
				{Name: "enabled", Type: "boolean"},
				{Name: "count", Type: "integer"},
			},
		},
		{
			name:   "invalid schema type",
			schema: "not-a-map",
			want:   []ParameterInfo{},
		},
		{
			name: "schema without properties",
			schema: map[string]any{
				"type": "object",
			},
			want: []ParameterInfo{},
		},
		{
			name: "properties not a map",
			schema: map[string]any{
				"type":       "object",
				"properties": "invalid",
			},
			want: []ParameterInfo{},
		},
		{
			name: "property with non-string type",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"param": map[string]any{"type": 123},
				},
			},
			want: []ParameterInfo{{Name: "param", Type: "string"}},
		},
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

func TestExtractParametersFromJSONSchema(t *testing.T) {
	// Test with actual JSON schema that would come from MCP
	jsonSchema := `{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum results",
				"default": 10
			},
			"filters": {
				"type": "object",
				"description": "Search filters"
			}
		},
		"required": ["query"]
	}`

	var schema map[string]any
	err := json.Unmarshal([]byte(jsonSchema), &schema)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON schema: %v", err)
	}

	got := extractParametersFromSchema(schema)

	expected := map[string]string{
		"query":   "string",
		"limit":   "integer",
		"filters": "object",
	}

	gotMap := make(map[string]string)
	for _, p := range got {
		gotMap[p.Name] = p.Type
	}

	if !reflect.DeepEqual(gotMap, expected) {
		t.Errorf("expected %v, got %v", expected, gotMap)
	}
}

// Note: getToolParameters requires a real *mcp.ClientSession, so we test
// the parameter extraction logic separately and leave integration testing
// for higher-level tests that can create real sessions.
