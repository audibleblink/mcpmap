# mcpmap

A command-line tool for interacting with Model Context Protocol (MCP) servers.

## Features

- **Multiple Transport Support**: Connect to MCP servers using SSE (Server-Sent Events) or HTTP transport
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
```
