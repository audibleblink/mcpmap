# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Building the Application

```bash
# Ensure dependencies are up to date
go mod tidy

# Build the application
go build -o mcpmap

# Run the compiled binary
./mcpmap

# Use `go doc` for library documentation
go doc $package
```

### Running the Application
```bash
# Run directly from source
go run main.go

# Run with specific MCP server URL
MCP_SERVER_URL=https://your-mcp-server.example.com/mcp/ go run main.go

# Run with custom client name and version
MCP_CLIENT_NAME="custom-client" MCP_CLIENT_VERSION="v2.0.0" go run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...
```

## High-Level Architecture

This repository contains a Go client for the Model Context Protocol (MCP), which allows AI assistants to connect to external data sources and tools.

### Key Components

1. **MCP Client**: Creates a client with implementation metadata.
   ```go
   client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)
   ```

2. **Transport**: Sets up SSE (Server-Sent Events) transport to connect to an MCP server.
   ```go
   transport := mcp.NewSSEClientTransport(serverURL, opts)
   ```

3. **Session Management**: Establishes connections and manages communication sessions.
   ```go
   session, err := client.Connect(ctx, transport)
   defer session.Close()
   ```

4. **Tool Discovery**: Lists available tools from the MCP server.
   ```go
   toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
   ```

5. **Output Handling**: Formats and outputs tool information as JSON.
   ```go
   for _, tool := range toolsRes.Tools {
       js, err := json.Marshal(tool)
       fmt.Fprintln(os.Stdout, string(js))
   }
   ```

## Configuration Options

The application can be configured using environment variables:

| Environment Variable | Description | Default Value |
|----------------------|-------------|---------------|
| `MCP_SERVER_URL` | MCP server endpoint URL | https://snyk-detector-mcp-server.dev.walmart.com/mcp/ |
| `MCP_CLIENT_NAME` | Client implementation name | "mcp-client" |
| `MCP_CLIENT_VERSION` | Client implementation version | "v1.0.0" |

## Dependencies

- Go 1.23.4
- Main dependency: `github.com/modelcontextprotocol/go-sdk v0.2.0`
- Indirect dependency: `github.com/yosida95/uritemplate/v3 v3.0.2`

## Error Handling Approach

- Critical errors that should terminate the application use `log.Fatal()`
- Application logs are separated from output data with `log.SetOutput(os.Stderr)`
- Output data is sent to stdout for easy redirection and processing
