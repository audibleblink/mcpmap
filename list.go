package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

var listCmd = &cobra.Command{
	Use:   "list [resources|tools|prompts]",
	Short: "List available resources, tools, or prompts from the MCP server",
	Long:  `List available resources, tools, or prompts from the MCP server. If no type is specified, all types will be listed.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in raw JSON format")
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
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

	// Determine what to list
	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	switch listType {
	case "tools":
		return listTools(ctx, session)
	case "resources":
		return listResources(ctx, session)
	case "prompts":
		return listPrompts(ctx, session)
	case "all":
		if err := listTools(ctx, session); err != nil {
			return err
		}
		if err := listResources(ctx, session); err != nil {
			return err
		}
		return listPrompts(ctx, session)
	default:
		return fmt.Errorf("unknown list type '%s', supported types: tools, resources, prompts", listType)
	}
}

func listTools(ctx context.Context, session *mcp.ClientSession) error {
	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	if jsonOutput {
		for _, tool := range toolsRes.Tools {
			js, err := json.Marshal(tool)
			if err != nil {
				return fmt.Errorf("json marshal tool: %w", err)
			}
			fmt.Fprintln(os.Stdout, string(js))
		}
	} else {
		for _, tool := range toolsRes.Tools {
			fmt.Println(tool.Name)
		}
	}
	return nil
}

func listResources(ctx context.Context, session *mcp.ClientSession) error {
	resourcesRes, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		return fmt.Errorf("list resources: %w", err)
	}

	if jsonOutput {
		for _, resource := range resourcesRes.Resources {
			js, err := json.Marshal(resource)
			if err != nil {
				return fmt.Errorf("json marshal resource: %w", err)
			}
			fmt.Fprintln(os.Stdout, string(js))
		}
	} else {
		for _, resource := range resourcesRes.Resources {
			fmt.Println(resource.Name)
		}
	}
	return nil
}

func listPrompts(ctx context.Context, session *mcp.ClientSession) error {
	promptsRes, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		return fmt.Errorf("list prompts: %w", err)
	}

	if jsonOutput {
		for _, prompt := range promptsRes.Prompts {
			js, err := json.Marshal(prompt)
			if err != nil {
				return fmt.Errorf("json marshal prompt: %w", err)
			}
			fmt.Fprintln(os.Stdout, string(js))
		}
	} else {
		for _, prompt := range promptsRes.Prompts {
			fmt.Println(prompt.Name)
		}
	}
	return nil
}