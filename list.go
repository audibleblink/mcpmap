package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var jsonOutput bool

var listCmd = &cobra.Command{
	Use:   "list [resources|tools|prompts]",
	Short: "List available resources, tools, or prompts from the MCP server",
	Long:  `List available resources, tools, or prompts from the MCP server. If no type is specified, all types will be listed.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in raw JSON format")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	session, err := createSession(ctx, transportType, serverURL, proxyURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	switch listType {
	case "tools":
		listTools(ctx, session)
	case "resources":
		listResources(ctx, session)
	case "prompts":
		listPrompts(ctx, session)
	case "all":
		listTools(ctx, session)
		listResources(ctx, session)
		listPrompts(ctx, session)
	default:
		return fmt.Errorf(
			"unknown list type '%s', supported types: tools, resources, prompts",
			listType,
		)
	}

	return nil
}

func outputItems(items []any, prefix string) {
	if jsonOutput {
		for _, item := range items {
			if js, err := json.Marshal(item); err == nil {
				fmt.Fprintln(os.Stdout, string(js))
			}
		}
	} else {
		for _, item := range items {
			fmt.Printf("%s:%s\n", prefix, getItemName(item))
		}
	}
}

func getItemName(item any) string {
	switch v := item.(type) {
	case *mcp.Tool:
		return v.Name
	case *mcp.Resource:
		return v.URI
	case *mcp.Prompt:
		return v.Name
	default:
		return "unknown"
	}
}

func listTools(ctx context.Context, session *mcp.ClientSession) {
	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return
	}

	items := make([]any, len(toolsRes.Tools))
	for i, tool := range toolsRes.Tools {
		items[i] = tool
	}
	outputItems(items, "tool")
}

func listResources(ctx context.Context, session *mcp.ClientSession) {
	resourcesRes, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		return
	}

	items := make([]any, len(resourcesRes.Resources))
	for i, resource := range resourcesRes.Resources {
		items[i] = resource
	}
	outputItems(items, "resource")
}

func listPrompts(ctx context.Context, session *mcp.ClientSession) {
	promptsRes, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		return
	}

	items := make([]any, len(promptsRes.Prompts))
	for i, prompt := range promptsRes.Prompts {
		items[i] = prompt
	}
	outputItems(items, "prompt")
}
