package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func createTransport(transportType, serverURL string) (mcp.Transport, error) {
	switch strings.ToLower(transportType) {
	case "streamable", "streamable-http", "http":
		return mcp.NewStreamableClientTransport(serverURL, &mcp.StreamableClientTransportOptions{
			HTTPClient: &http.Client{},
		}), nil
	case "sse":
		return mcp.NewSSEClientTransport(serverURL, &mcp.SSEClientTransportOptions{}), nil
	default:
		return nil, fmt.Errorf(
			"unknown transport type '%s', supported types: sse, streamable-http",
			transportType,
		)
	}
}

func createSession(
	ctx context.Context,
	transportType, serverURL string,
) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcpmap", Version: "v1.0.0"}, nil)
	transport, err := createTransport(transportType, serverURL)
	if err != nil {
		return nil, err
	}

	session, err := client.Connect(ctx, transport)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func getTools(ctx context.Context, session *mcp.ClientSession) ([]*mcp.Tool, error) {
	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, err
	}

	return toolsRes.Tools, nil
}

func getToolParameters(
	ctx context.Context,
	session *mcp.ClientSession,
	toolName string,
) ([]ParameterInfo, error) {
	tools, err := getTools(ctx, session)
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if tool.Name == toolName {
			return extractParametersFromSchema(tool.InputSchema), nil
		}
	}

	return nil, fmt.Errorf("tool %q not found", toolName)
}
