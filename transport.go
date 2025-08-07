package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTransport creates the appropriate transport based on the transport type and server URL
func createTransport(transportType, serverURL string) (mcp.Transport, error) {
	switch strings.ToLower(transportType) {
	case "streamable", "streamable-http", "http":
		streamableOpts := &mcp.StreamableClientTransportOptions{
			HTTPClient: &http.Client{},
		}
		transport := mcp.NewStreamableClientTransport(serverURL, streamableOpts)
		if transport == nil {
			return nil, fmt.Errorf("failed to create StreamableClientTransport")
		}
		return transport, nil
	case "sse":
		sseOpts := &mcp.SSEClientTransportOptions{}
		transport := mcp.NewSSEClientTransport(serverURL, sseOpts)
		if transport == nil {
			return nil, fmt.Errorf("failed to create SSEClientTransport")
		}
		return transport, nil
	default:
		return nil, fmt.Errorf("unknown transport type '%s', supported types: sse, streamable-http", transportType)
	}
}