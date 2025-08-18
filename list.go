package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"mcpmap/cache"
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

// fetchAllServerData retrieves tools, resources, and prompts from the server
func fetchAllServerData(ctx context.Context, session *mcp.ClientSession) (*cache.CacheData, error) {
	var tools []*mcp.Tool
	var resources []*mcp.Resource
	var prompts []*mcp.Prompt

	// Fetch tools
	if toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{}); err == nil {
		tools = toolsRes.Tools
	}

	// Fetch resources
	if resourcesRes, err := session.ListResources(ctx, &mcp.ListResourcesParams{}); err == nil {
		resources = resourcesRes.Resources
	}

	// Fetch prompts
	if promptsRes, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{}); err == nil {
		prompts = promptsRes.Prompts
	}

	return &cache.CacheData{
		Tools:     tools,
		Resources: resources,
		Prompts:   prompts,
	}, nil
}

// displayCachedData displays cached data when server is unavailable
func displayCachedData(data *cache.CacheData, args []string) error {
	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	switch listType {
	case "tools":
		if data.Tools != nil {
			items := make([]any, len(data.Tools))
			for i, tool := range data.Tools {
				items[i] = tool
			}
			outputItems(items, "tool")
		}
	case "resources":
		if data.Resources != nil {
			items := make([]any, len(data.Resources))
			for i, resource := range data.Resources {
				items[i] = resource
			}
			outputItems(items, "resource")
		}
	case "prompts":
		if data.Prompts != nil {
			items := make([]any, len(data.Prompts))
			for i, prompt := range data.Prompts {
				items[i] = prompt
			}
			outputItems(items, "prompt")
		}
	case "all":
		if data.Tools != nil {
			items := make([]any, len(data.Tools))
			for i, tool := range data.Tools {
				items[i] = tool
			}
			outputItems(items, "tool")
		}
		if data.Resources != nil {
			items := make([]any, len(data.Resources))
			for i, resource := range data.Resources {
				items[i] = resource
			}
			outputItems(items, "resource")
		}
		if data.Prompts != nil {
			items := make([]any, len(data.Prompts))
			for i, prompt := range data.Prompts {
				items[i] = prompt
			}
			outputItems(items, "prompt")
		}
	default:
		return fmt.Errorf(
			"unknown list type '%s', supported types: tools, resources, prompts",
			listType,
		)
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize cache
	c := cache.New(serverURL, transportType, authToken, clientName)

	// Try cache first for faster response (will still fetch fresh data)
	var cachedData *cache.CacheData
	if data, _, _ := c.Load(); data != nil {
		cachedData = data
	}

	// Query server for fresh data synchronously
	session, err := createSession(ctx, transportType, serverURL, proxyURL, authToken, clientName)
	if err != nil {
		// If cache exists, use it as fallback
		if cachedData != nil {
			fmt.Fprintf(os.Stderr, "Warning: Using cached data (server unavailable)\n")
			return displayCachedData(cachedData, args)
		}
		return fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	freshData, err := fetchAllServerData(ctx, session)
	if err == nil && freshData != nil {
		// Save synchronously so cache file is guaranteed written before exit
		_ = c.Save(freshData)
		cachedData = freshData
	}

	// Display data (fresh if successful, else cached)
	if cachedData == nil {
		return fmt.Errorf("no data available")
	}

	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	switch listType {
	case "tools":
		items := make([]any, len(cachedData.Tools))
		for i, tool := range cachedData.Tools { items[i] = tool }
		outputItems(items, "tool")
	case "resources":
		items := make([]any, len(cachedData.Resources))
		for i, r := range cachedData.Resources { items[i] = r }
		outputItems(items, "resource")
	case "prompts":
		items := make([]any, len(cachedData.Prompts))
		for i, p := range cachedData.Prompts { items[i] = p }
		outputItems(items, "prompt")
	case "all":
		items := make([]any, len(cachedData.Tools))
		for i, tool := range cachedData.Tools { items[i] = tool }
		outputItems(items, "tool")
		items = make([]any, len(cachedData.Resources))
		for i, r := range cachedData.Resources { items[i] = r }
		outputItems(items, "resource")
		items = make([]any, len(cachedData.Prompts))
		for i, p := range cachedData.Prompts { items[i] = p }
		outputItems(items, "prompt")
	default:
		return fmt.Errorf("unknown list type '%s', supported types: tools, resources, prompts", listType)
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

// listItems is a generic function to handle the common pattern of list operations
func listItems[T any](ctx context.Context, session *mcp.ClientSession, 
	fetchFunc func(context.Context, *mcp.ClientSession) ([]T, error), 
	itemType string) {
	
	items, err := fetchFunc(ctx, session)
	if err != nil {
		return
	}

	anyItems := make([]any, len(items))
	for i, item := range items {
		anyItems[i] = item
	}
	outputItems(anyItems, itemType)
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
