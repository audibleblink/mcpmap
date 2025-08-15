# mcpmap

A command-line tool for interacting with Model Context Protocol (MCP) servers.

![](https://private-user-images.githubusercontent.com/4605783/476181529-b391a8f3-b170-471e-ac90-4bf77ec66a65.png?jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NTQ2ODYxODMsIm5iZiI6MTc1NDY4NTg4MywicGF0aCI6Ii80NjA1NzgzLzQ3NjE4MTUyOS1iMzkxYThmMy1iMTcwLTQ3MWUtYWM5MC00YmY3N2VjNjZhNjUucG5nP1gtQW16LUFsZ29yaXRobT1BV1M0LUhNQUMtU0hBMjU2JlgtQW16LUNyZWRlbnRpYWw9QUtJQVZDT0RZTFNBNTNQUUs0WkElMkYyMDI1MDgwOCUyRnVzLWVhc3QtMSUyRnMzJTJGYXdzNF9yZXF1ZXN0JlgtQW16LURhdGU9MjAyNTA4MDhUMjA0NDQzWiZYLUFtei1FeHBpcmVzPTMwMCZYLUFtei1TaWduYXR1cmU9OWZkODA3NTc1YTYwMGE3YjE3Mzk0MDk0Nzc1YzkwMGMzMTZmZGY2ZDFhZDc5MGU5MWIwMmVmMWU1ZTUwNWVmNyZYLUFtei1TaWduZWRIZWFkZXJzPWhvc3QifQ.2tN9C48Wu7N1DXQA4m3HFNrbilnlqUDrflLokfXcsmw)
## Features

- **Multiple Transport Support**: Connect to MCP servers using SSE (Server-Sent Events) or HTTP transport
- **Authentication Support**: Bearer token authentication for secure MCP servers
- **Proxy Support**: Route HTTP requests through an HTTP proxy server
- **Smart Type Conversion**: Automatically convert CLI parameters to correct types based on tool schemas
- **Tool Execution**: Execute tools on MCP servers with typed parameters
- **Resource Listing**: List available resources, tools, and prompts
- **Tab Completion**: Smart tab completion for tool names and parameters
- **Graceful Fallback**: Falls back to string parameters when schema is unavailable

## Installation

```bash
go build -o mcpmap
```

## Examples

### Basic Usage

```bash
# Connect to a local MCP server and list all tools
mcpmap --sse=http://localhost:3000 list tools

# Execute a file reading tool with tab completion
mcpmap --sse=http://localhost:3000 exec read_file --param path=/etc/hosts

# List resources with JSON output
mcpmap --sse=http://localhost:3000 list resources --json
```

### Type Conversion Examples

```bash
# Automatic type conversion based on tool schema
mcpmap --sse=http://localhost:3000 exec search --param query="user data" --param limit=50

# Boolean parameters (accepts: true/false, yes/no, 1/0, on/off)
mcpmap --sse=http://localhost:3000 exec toggle --param enabled=true --param debug=no

# Array parameters (comma-separated or JSON)
mcpmap --sse=http://localhost:3000 exec filter --param tags=red,blue,green
mcpmap --sse=http://localhost:3000 exec filter --param ids=[1,2,3,4]

# Complex object parameters (JSON format)
mcpmap --sse=http://localhost:3000 exec query --param filter='{"age":{"min":18,"max":65}}'

# Numeric parameters (integers and floats)
mcpmap --sse=http://localhost:3000 exec calculate --param x=42 --param y=3.14159
```

### Advanced Usage

```bash
# Use HTTP transport with a proxy server
mcpmap --http=http://localhost:8080 --proxy=http://proxy.example.com:8080 list tools

# Connect to an authenticated MCP server
mcpmap --sse=https://mcp.sentry.dev/sse --token=your-bearer-token list tools

# Execute a tool with multiple typed parameters
mcpmap --sse=http://localhost:3000 exec process_data \
  --param input_file="/path/to/data.csv" \
  --param batch_size=1000 \
  --param parallel=true \
  --param options='{"format":"json","compress":true}'
```

## Type Conversion

mcpmap automatically converts CLI parameters to their expected types based on the tool's JSON schema:

| Type | CLI Input Examples | Converted To | Notes |
|------|-------------------|--------------|-------|
| **string** | `hello`, `"hello world"` | string | Handles quoted strings |
| **integer** | `42`, `-10` | int64 | Rejects decimals |
| **number** | `3.14`, `-0.5`, `42` | float64 | Accepts integers |
| **boolean** | `true`, `yes`, `1`, `on` | bool | Case-insensitive |
| **array** | `[1,2,3]`, `a,b,c` | []any | JSON or CSV format |
| **object** | `{"key":"value"}` | map[string]any | JSON required |

### Error Handling

When type conversion fails, mcpmap provides helpful error messages:

```bash
$ mcpmap exec search --param limit=abc
Error: parameter "limit" (type: integer): cannot convert "abc"
Hint: Use whole numbers like 42 or -10

$ mcpmap exec toggle --param enabled=maybe
Error: parameter "enabled" (type: boolean): cannot convert "maybe"
Hint: Use true/false, yes/no, 1/0, or on/off
```
