package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func createTransport(transportType, serverURL, proxyURL, authToken string) (mcp.Transport, error) {
	switch strings.ToLower(transportType) {
	case "streamable", "streamable-http", "http":
		var httpClient *http.Client
		if proxyURL != "" || authToken != "" {
			transport := &http.Transport{}
			if proxyURL != "" {
				proxyURLParsed, err := url.Parse(proxyURL)
				if err != nil {
					return nil, fmt.Errorf("invalid proxy URL: %w", err)
				}
				transport.Proxy = http.ProxyURL(proxyURLParsed)
			}
			httpClient = &http.Client{Transport: transport}

			// Add authentication if token is provided
			if authToken != "" {
				httpClient.Transport = &authTransport{
					base:  transport,
					token: authToken,
				}
			}
		} else {
			httpClient = &http.Client{}
		}
		return mcp.NewStreamableClientTransport(serverURL, &mcp.StreamableClientTransportOptions{
			HTTPClient: httpClient,
		}), nil
	case "sse":
		var httpClient *http.Client
		if proxyURL != "" || authToken != "" {
			transport := &http.Transport{}
			if proxyURL != "" {
				proxyURLParsed, err := url.Parse(proxyURL)
				if err != nil {
					return nil, fmt.Errorf("invalid proxy URL: %w", err)
				}
				transport.Proxy = http.ProxyURL(proxyURLParsed)
			}
			httpClient = &http.Client{Transport: transport}

			// Add authentication if token is provided
			if authToken != "" {
				httpClient.Transport = &authTransport{
					base:  transport,
					token: authToken,
				}
			}
		} else {
			httpClient = &http.Client{}
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

// authTransport wraps an http.RoundTripper to add authentication headers
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(reqClone)
}

func createSession(
	ctx context.Context,
	transportType, serverURL, proxyURL, authToken string,
) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "mcpmap", Version: "v1.0.0"}, nil)
	transport, err := createTransport(transportType, serverURL, proxyURL, authToken)
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
