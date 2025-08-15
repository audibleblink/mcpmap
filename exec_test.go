package main

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name    string
		params  []string
		want    map[string]any
		wantErr bool
		errMsg  string
	}{
		{"empty", []string{}, map[string]any{}, false, ""},
		{"single param", []string{"name=value"}, map[string]any{"name": "value"}, false, ""},
		{"multiple params", []string{"name=John", "age=30"}, map[string]any{"name": "John", "age": "30"}, false, ""},
		{"equals in value", []string{"url=http://example.com?key=value"}, map[string]any{"url": "http://example.com?key=value"}, false, ""},
		{"with spaces", []string{" name = John Doe "}, map[string]any{"name": "John Doe"}, false, ""},
		{"no equals", []string{"invalid"}, nil, true, "invalid parameter format"},
		{"empty name", []string{"=value"}, nil, true, "parameter name cannot be empty"},
		{"empty value", []string{"name="}, map[string]any{"name": ""}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParams(tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestExtractServerConfig(t *testing.T) {
	h := newTestHelper(t)

	tests := []struct {
		name      string
		transport string
		url       string
		wantURL   string
		wantType  string
	}{
		{"sse flag", "sse", "http://localhost:3000", "http://localhost:3000", "sse"},
		{"http flag", "http", "http://localhost:8080", "http://localhost:8080", "http"},
		{"no flags", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := h.createCmdWithFlags()
			if tt.transport != "" {
				h.setTransportFlag(cmd, tt.transport, tt.url)
			}

			gotURL, gotType := extractServerConfig(cmd)

			if gotURL != tt.wantURL {
				t.Errorf("expected URL %q, got %q", tt.wantURL, gotURL)
			}
			if gotType != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, gotType)
			}
		})
	}
}

func TestCompletionFunctions(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)
		args       []string
		toComplete string
		serverURL  string
		wantEmpty  bool
	}{
		{"tool completion no server", toolNameCompletion, []string{}, "tool", "", true},
		{"tool completion with args", toolNameCompletion, []string{"existing"}, "tool", "http://localhost:3000", true},
		{"param completion no args", paramCompletion, []string{}, "param", "http://localhost:3000", true},
		{"param completion no server", paramCompletion, []string{"tool"}, "param", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			if tt.serverURL != "" {
				cmd.Flags().String("sse", "", "")
				cmd.Flags().Set("sse", tt.serverURL)
			}

			completions, directive := tt.fn(cmd, tt.args, tt.toComplete)

			if tt.wantEmpty && len(completions) > 0 {
				t.Errorf("expected no completions but got %v", completions)
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("expected NoFileComp directive, got %v", directive)
			}
		})
	}
}

func TestExecCommandConfiguration(t *testing.T) {
	if execCmd.Use != "exec <tool>" {
		t.Errorf("unexpected exec command Use: %q", execCmd.Use)
	}

	if execCmd.Short != "Execute a tool on the MCP server with automatic type conversion" {
		t.Errorf("unexpected exec command Short: %q", execCmd.Short)
	}

	// Check param flag exists
	if execCmd.Flags().Lookup("param") == nil {
		t.Error("param flag not found")
	}
}

func TestParameterInfo(t *testing.T) {
	param := ParameterInfo{Name: "test", Type: "string"}

	if param.Name != "test" {
		t.Errorf("expected Name 'test', got %q", param.Name)
	}
	if param.Type != "string" {
		t.Errorf("expected Type 'string', got %q", param.Type)
	}
}
