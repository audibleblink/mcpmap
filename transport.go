package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createHTTPClient creates an HTTP client with optional proxy and authentication
func createHTTPClient(proxyURL, authToken string) (*http.Client, error) {
	if proxyURL == "" && authToken == "" {
		return &http.Client{}, nil
	}

	transport := &http.Transport{}

	if proxyURL != "" {
		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURLParsed)
	}

	httpClient := &http.Client{Transport: transport}

	// Add authentication if token is provided
	if authToken != "" {
		httpClient.Transport = &authTransport{
			base:  transport,
			token: authToken,
		}
	}

	return httpClient, nil
}

func createTransport(
	transportType, serverURL, proxyURL, authToken, clientName string,
) (mcp.Transport, error) {
	_ = clientName
	httpClient, err := createHTTPClient(proxyURL, authToken)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(transportType) {
	case "streamable", "streamable-http", "http":
		return mcp.NewStreamableClientTransport(serverURL, &mcp.StreamableClientTransportOptions{
			HTTPClient: httpClient,
		}), nil
	case "sse":
		return mcp.NewSSEClientTransport(serverURL, &mcp.SSEClientTransportOptions{
			HTTPClient: httpClient,
		}), nil
	default:
		return nil, fmt.Errorf(
			"unknown transport type '%s', supported types: sse, streamable-http",
			transportType,
		)
	}
}

// authTransport wraps an http.RoundTripper to add authentication headers
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(reqClone)
}

func createSession(
	ctx context.Context,
	transportType, serverURL, proxyURL, authToken, clientName string,
) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcpmap", Version: "v1.0.0"}, nil)
	transport, err := createTransport(transportType, serverURL, proxyURL, authToken, clientName)
	if err != nil {
		return nil, err
	}

	session, err := client.Connect(ctx, transport)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// getToolSchema fetches the schema for a specific tool with timeout
func getToolSchema(ctx context.Context, session *mcp.ClientSession, toolName string) (*ToolSchema, error) {
	schemaCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	toolsRes, err := session.ListTools(schemaCtx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	for _, tool := range toolsRes.Tools {
		if tool.Name == toolName {
			if tool.InputSchema == nil {
				return &ToolSchema{
					Parameters: make(map[string]*ParameterSchema),
					Required:   []string{},
				}, nil
			}

			schema, err := extractFullSchema(tool.InputSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to extract schema for tool %q: %w", toolName, err)
			}

			return schema, nil
		}
	}

	return nil, fmt.Errorf("tool %q not found", toolName)
}
