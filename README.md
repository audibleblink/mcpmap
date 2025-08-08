# mcpmap

A command-line tool for interacting with Model Context Protocol (MCP) servers.

![](https://private-user-images.githubusercontent.com/4605783/476181529-b391a8f3-b170-471e-ac90-4bf77ec66a65.png?jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NTQ2ODYxODMsIm5iZiI6MTc1NDY4NTg4MywicGF0aCI6Ii80NjA1NzgzLzQ3NjE4MTUyOS1iMzkxYThmMy1iMTcwLTQ3MWUtYWM5MC00YmY3N2VjNjZhNjUucG5nP1gtQW16LUFsZ29yaXRobT1BV1M0LUhNQUMtU0hBMjU2JlgtQW16LUNyZWRlbnRpYWw9QUtJQVZDT0RZTFNBNTNQUUs0WkElMkYyMDI1MDgwOCUyRnVzLWVhc3QtMSUyRnMzJTJGYXdzNF9yZXF1ZXN0JlgtQW16LURhdGU9MjAyNTA4MDhUMjA0NDQzWiZYLUFtei1FeHBpcmVzPTMwMCZYLUFtei1TaWduYXR1cmU9OWZkODA3NTc1YTYwMGE3YjE3Mzk0MDk0Nzc1YzkwMGMzMTZmZGY2ZDFhZDc5MGU5MWIwMmVmMWU1ZTUwNWVmNyZYLUFtei1TaWduZWRIZWFkZXJzPWhvc3QifQ.2tN9C48Wu7N1DXQA4m3HFNrbilnlqUDrflLokfXcsmw)
## Features

- **Multiple Transport Support**: Connect to MCP servers using SSE (Server-Sent Events) or HTTP transport
- **Proxy Support**: Route HTTP requests through an HTTP proxy server
- **Tool Execution**: Execute tools on MCP servers with parameters
- **Resource Listing**: List available resources, tools, and prompts
- **Tab Completion**: Smart tab completion for tool names and parameters

## Installation

```bash
go build -o mcpmap
```

## Examples

```bash
# Connect to a local MCP server and list all tools
mcpmap --sse=http://localhost:3000 list tools

# Execute a file reading tool with tab completion
mcpmap --sse=http://localhost:3000 exec read_file --param path=/etc/hosts

# List resources with JSON output
mcpmap --sse=http://localhost:3000 list resources --json

# Execute a tool with multiple parameters
mcpmap --sse=http://localhost:3000 exec write_file --param path=/tmp/output.txt --param content="Hello World"

# Use HTTP transport with a proxy server
mcpmap --http=http://localhost:8080 --proxy=http://proxy.example.com:8080 list tools
```
