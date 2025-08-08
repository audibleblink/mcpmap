package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func createTransport(transportType, serverURL, proxyURL string) (mcp.Transport, error) {
	switch strings.ToLower(transportType) {
	case "streamable", "streamable-http", "http":
		var baseTransport http.RoundTripper = http.DefaultTransport
		
		// Configure proxy if provided
		if proxyURL != "" {
			proxyURLParsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL: %w", err)
			}
			baseTransport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}

		// Wrap with session-aware round tripper
		sessionTransport := newSessionAwareRoundTripper(baseTransport)
		
		httpClient := &http.Client{
			Transport: sessionTransport,
		}

		return mcp.NewStreamableClientTransport(serverURL, &mcp.StreamableClientTransportOptions{
			HTTPClient: httpClient,
		}), nil
	case "sse":
		var baseTransport http.RoundTripper = http.DefaultTransport
		
		// Configure proxy if provided
		if proxyURL != "" {
			proxyURLParsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL: %w", err)
			}
			baseTransport = &http.Transport{
				Proxy: http.ProxyURL(proxyURLParsed),
			}
		}

		// Wrap with session-aware round tripper
		sessionTransport := newSessionAwareRoundTripper(baseTransport)
		
		httpClient := &http.Client{
			Transport: sessionTransport,
		}

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

// sessionAwareRoundTripper wraps an http.RoundTripper to handle MCP session IDs
type sessionAwareRoundTripper struct {
	transport http.RoundTripper
	sessionID string
	firstCall bool
}

// newSessionAwareRoundTripper creates a new session-aware round tripper
func newSessionAwareRoundTripper(transport http.RoundTripper) *sessionAwareRoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &sessionAwareRoundTripper{
		transport: transport,
		firstCall: true,
	}
}

// RoundTrip implements http.RoundTripper interface
func (s *sessionAwareRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add session ID to request headers if we have one
	if s.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", s.sessionID)
	}

	// Make the request
	resp, err := s.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Check for session ID in response headers (only on first successful response)
	if s.firstCall {
		if mcpSessionID := resp.Header.Get("Mcp-Session-Id"); mcpSessionID != "" {
			s.sessionID = mcpSessionID
		}
		s.firstCall = false
	}

	return resp, nil
}

func createSession(
	ctx context.Context,
	transportType, serverURL, proxyURL string,
) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcpmap", Version: "v1.0.0"}, nil)
	transport, err := createTransport(transportType, serverURL, proxyURL)
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