package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// testHelper provides common test utilities
type testHelper struct {
	t *testing.T
}

func newTestHelper(t *testing.T) *testHelper {
	return &testHelper{t: t}
}

// createCmdWithFlags creates a command with standard transport flags
func (h *testHelper) createCmdWithFlags() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("sse", "", "")
	cmd.Flags().String("http", "", "")
	return cmd
}

// setTransportFlag sets either sse or http flag on a command
func (h *testHelper) setTransportFlag(cmd *cobra.Command, transport, url string) {
	switch transport {
	case "sse":
		cmd.Flags().Set("sse", url)
	case "http":
		cmd.Flags().Set("http", url)
	}
}

// captureOutput captures stdout during function execution
func (h *testHelper) captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	fn()
	w.Close()

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// assertStringContains checks if output contains expected strings
func (h *testHelper) assertStringContains(output string, expected []string) {
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			h.t.Errorf("expected output to contain %q, but got: %q", exp, output)
		}
	}
}

// Mock session for testing parameter extraction
type mockSession struct {
	tools []*mcp.Tool
	err   error
}

func (m *mockSession) ListTools(
	ctx context.Context,
	params *mcp.ListToolsParams,
) (*mcp.ListToolsResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mcp.ListToolsResult{Tools: m.tools}, nil
}

func (m *mockSession) Close() error { return nil }

// Implement other required methods to satisfy the interface
func (m *mockSession) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSession) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSession) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSession) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSession) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSession) SetLevel(ctx context.Context, params *mcp.SetLevelParams) error {
	return errors.New("not implemented")
}
