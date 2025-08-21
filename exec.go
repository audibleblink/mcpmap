package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"mcpmap/cache"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var params []string

var execCmd = &cobra.Command{
	Use:   "exec <tool>",
	Short: "Execute a tool on the MCP server with automatic type conversion",
	Long: `Execute a tool on the MCP server with automatic type conversion.

Parameters are automatically converted to their expected types based on the tool's schema.
If schema fetching fails, parameters are treated as strings (backward compatibility).

Examples:
  # Simple types
  mcpmap exec search --param query="user login" --param limit=10
  
  # Boolean values (accepts: true/false, yes/no, 1/0, on/off)
  mcpmap exec toggle --param enabled=true --param verbose=yes
  
  # Arrays (comma-separated or JSON)
  mcpmap exec filter --param tags=red,blue --param ids=[1,2,3]
  
  # Complex objects (JSON required)
  mcpmap exec query --param filter='{"age":{"min":18}}'
  
  # Numbers (integers and floats)
  mcpmap exec calculate --param x=10 --param y=3.14`,
	Args: cobra.ExactArgs(1),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().
		StringArrayVar(&params, "param", []string{}, "Specify a parameter for the tool in format name=value (can be repeated)")

	execCmd.ValidArgsFunction = toolNameCompletion
	execCmd.RegisterFlagCompletionFunc("param", paramCompletion)
}

func runExec(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	toolName := args[0]

	return withSession(ctx, func(session *mcp.ClientSession) error {
		// Try to fetch schema (best-effort)
		var toolParams map[string]any
		schema, err := getToolSchema(ctx, session, toolName)
		if err != nil {
			// Schema fetch failed, warn and fall back to string parsing
			fmt.Fprintf(os.Stderr, "Warning: Could not fetch schema for tool %q: %v\n", toolName, err)
			fmt.Fprintf(os.Stderr, "Warning: Using string-only parameter parsing\n")

			toolParams, err = parseParams(params)
			if err != nil {
				return fmt.Errorf("parse parameters: %w", err)
			}
		} else {
			// Schema available, use schema-based parsing
			toolParams, err = parseParamsWithSchema(params, schema)
			if err != nil {
				return fmt.Errorf("parse parameters with schema: %w", err)
			}
		}

		result, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: toolParams,
		})
		if err != nil {
			return err
		}

		js, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("json marshal result: %w", err)
		}
		fmt.Fprintln(os.Stdout, string(js))
		return nil
	})
}

func parseParams(params []string) (map[string]any, error) {
	result := make(map[string]any)

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format '%s', expected name=value", param)
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if name == "" {
			return nil, fmt.Errorf("parameter name cannot be empty in '%s'", param)
		}

		result[name] = value
	}

	return result, nil
}

// ParameterInfo represents information about a tool parameter
type ParameterInfo struct {
	Name string
	Type string
}

func extractParametersFromSchema(schema any) []ParameterInfo {
	var params []ParameterInfo

	if schema == nil {
		return params
	}

	// Handle jsonschema.Schema type by converting to JSON and back
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return params
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		return params
	}

	properties, ok := schemaMap["properties"].(map[string]any)
	if !ok {
		return params
	}

	for name, propData := range properties {
		paramInfo := ParameterInfo{Name: name, Type: "string"}

		if propMap, ok := propData.(map[string]any); ok {
			if typeVal, exists := propMap["type"]; exists {
				if typeStr, ok := typeVal.(string); ok {
					paramInfo.Type = typeStr
				}
			}
		}

		params = append(params, paramInfo)
	}

	return params
}

func extractServerConfig(cmd *cobra.Command) (serverURL, transportType string) {
	if sseFlag := cmd.Flag("sse"); sseFlag != nil && sseFlag.Changed {
		return sseFlag.Value.String(), "sse"
	}
	if httpFlag := cmd.Flag("http"); httpFlag != nil && httpFlag.Changed {
		return httpFlag.Value.String(), "http"
	}
	return "", ""
}

// withSession creates a session, invokes fn, and ensures the session is closed.
// It returns any error produced during session creation or execution.
func withSession(ctx context.Context, fn func(*mcp.ClientSession) error) error {
	session, err := createSession(ctx, transportType, serverURL, proxyURL, authToken, clientName)
	if err != nil {
		return err
	}
	defer session.Close()
	return fn(session)
}

func toolNameCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	serverURL, transportType := extractServerConfig(cmd)
	if serverURL == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Try cache first
	c := cache.New(serverURL, transportType, authToken, clientName)
	if data, _, _ := c.Load(); data != nil && len(data.Tools) > 0 {
		completions := make([]string, 0, len(data.Tools))
		for _, tool := range data.Tools {
			completions = append(completions, tool.Name)
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Cache miss - query server
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := createSession(ctx, transportType, serverURL, proxyURL, authToken, clientName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer session.Close()

	tools, err := getTools(ctx, session)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Update cache for next time
	go func() {
		cacheData := &cache.CacheData{Tools: tools}
		c.Save(cacheData)
	}()

	completions := make([]string, 0, len(tools))
	for _, tool := range tools {
		completions = append(completions, tool.Name)
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func paramCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	// Need tool name to get parameters
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	serverURL, transportType := extractServerConfig(cmd)
	if serverURL == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	toolName := args[0]

	// Try cache first
	c := cache.New(serverURL, transportType, authToken, clientName)
	if data, _, _ := c.Load(); data != nil && len(data.Tools) > 0 {
		// Find the tool in cached data
		for _, tool := range data.Tools {
			if tool.Name == toolName {
				params := extractParametersFromSchema(tool.InputSchema)
				if len(params) > 0 {
					completions := make([]string, 0, len(params))
					for _, param := range params {
						completions = append(completions, param.Name+"=")
					}
					return completions, cobra.ShellCompDirectiveNoFileComp
				}
				break
			}
		}
	}

	// Cache miss or tool not found - query server
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := createSession(ctx, transportType, serverURL, proxyURL, authToken, clientName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer session.Close()

	params, err := getToolParameters(ctx, session, toolName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Update cache for next time (get all tools to cache them)
	go func() {
		if tools, err := getTools(ctx, session); err == nil {
			cacheData := &cache.CacheData{Tools: tools}
			c.Save(cacheData)
		}
	}()

	completions := make([]string, 0, len(params))
	for _, param := range params {
		completions = append(completions, param.Name+"=")
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
