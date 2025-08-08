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

## Usage

### Basic Commands

```bash
# List all available tools, resources, and prompts
mcpmap --sse=http://localhost:3000 list

# List only tools
mcpmap --sse=http://localhost:3000 list tools

# Execute a tool
mcpmap --sse=http://localhost:3000 exec read_file --param path=/tmp/example.txt

# Using HTTP transport instead of SSE
mcpmap --http=http://localhost:3000 list tools
```

### Tab Completion

mcpmap supports intelligent tab completion for tool names and parameters. This feature provides real-time completion without caching, ensuring you always get the most up-to-date information from your MCP server.

#### Setup Completion

Generate and install completion scripts for your shell:

**Bash:**
```bash
# Load completion for current session
source <(mcpmap completion bash)

# Install permanently (Linux)
mcpmap completion bash > /etc/bash_completion.d/mcpmap

# Install permanently (macOS with Homebrew)
mcpmap completion bash > $(brew --prefix)/etc/bash_completion.d/mcpmap
```

**Zsh:**
```bash
# Enable completion system (if not already enabled)
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
mcpmap completion zsh > "${fpath[1]}/_mcpmap"

# Restart your shell or source your .zshrc
```

**Fish:**
```bash
# Load completion for current session
mcpmap completion fish | source

# Install permanently
mcpmap completion fish > ~/.config/fish/completions/mcpmap.fish
```

**PowerShell:**
```powershell
# Load completion for current session
mcpmap completion powershell | Out-String | Invoke-Expression

# Install permanently
mcpmap completion powershell > mcpmap.ps1
# Then source this file from your PowerShell profile
```

#### Using Completion

Once installed, you can use tab completion:

**Tool Name Completion:**
```bash
# Press TAB after 'exec' to see available tools
mcpmap --sse=http://localhost:3000 exec <TAB>
# Shows: file_operations  list_directory  read_file  write_file

# Partial matching works too
mcpmap --sse=http://localhost:3000 exec file<TAB>
# Shows: file_operations
```

**Parameter Completion:**
```bash
# Press TAB after '--param' to see available parameters for the tool
mcpmap --sse=http://localhost:3000 exec read_file --param <TAB>
# Shows: path=  encoding=

# You can complete multiple parameters
mcpmap --sse=http://localhost:3000 exec read_file --param path=/tmp/file --param <TAB>
# Shows: encoding=
```

#### Completion Features

- **Real-time Data**: Completion fetches fresh data from the MCP server on each request (no caching)
- **Error Handling**: Network failures result in empty completions without breaking your shell
- **Performance**: 3-second timeout ensures completion doesn't hang
- **Schema Parsing**: Automatically extracts parameter names from tool schemas

## Transport Options

### SSE (Server-Sent Events)
```bash
mcpmap --sse=http://localhost:3000 list
```

### HTTP
```bash
mcpmap --http=http://localhost:3000 list
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

## Development

### Building
```bash
go build -o mcpmap
```

### Testing
```bash
go test ./...
```

### Code Style
- Follow standard Go formatting with `gofmt`
- Use descriptive variable names
- Handle errors appropriately
- Add comments for exported functions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license information here]