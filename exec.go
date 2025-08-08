package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var params []string

var execCmd = &cobra.Command{
	Use:   "exec <tool>",
	Short: "Execute a tool on the MCP server",
	Long:  `Execute a tool on the MCP server with the specified parameters.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExec,
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

	session, err := createSession(ctx, transportType, serverURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	toolParams, err := parseParams(params)
	if err != nil {
		return fmt.Errorf("parse parameters: %w", err)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: toolParams,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	js, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("json marshal result: %w", err)
	}
	fmt.Fprintln(os.Stdout, string(js))

	return nil
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

	schemaMap, ok := schema.(map[string]any)
	if !ok {
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := createSession(ctx, transportType, serverURL)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer session.Close()

	tools, err := getTools(ctx, session)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := createSession(ctx, transportType, serverURL)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer session.Close()

	toolName := args[0]
	params, err := getToolParameters(ctx, session, toolName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	completions := make([]string, 0, len(params))
	for _, param := range params {
		completions = append(completions, param.Name+"=")
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
