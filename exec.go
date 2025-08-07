package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	params []string
)

var execCmd = &cobra.Command{
	Use:   "exec <tool>",
	Short: "Execute a tool on the MCP server",
	Long:  `Execute a tool on the MCP server with the specified parameters.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExec,
}

func init() {
	execCmd.Flags().StringArrayVar(&params, "param", []string{}, "Specify a parameter for the tool in format name=value (can be repeated)")
}

func runExec(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	toolName := args[0]
	
	// Create client and transport
	client := mcp.NewClient(&mcp.Implementation{Name: "mcpmap", Version: "v1.0.0"}, nil)
	transport, err := createTransport(transportType, serverURL)
	if err != nil {
		return err
	}

	// Connect to server
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer session.Close()

	// Parse parameters
	toolParams, err := parseParams(params)
	if err != nil {
		return fmt.Errorf("parse parameters: %w", err)
	}

	// Execute the tool
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: toolParams,
	})
	if err != nil {
		return fmt.Errorf("call tool: %w", err)
	}

	// Output the result
	js, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("json marshal result: %w", err)
	}
	fmt.Fprintln(os.Stdout, string(js))

	return nil
}

func parseParams(params []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
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