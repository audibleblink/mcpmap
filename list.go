package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"mcpmap/cache"
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

// loadServerData loads data from cache first, then tries server, with fallback to cache
func loadServerData(ctx context.Context) (*cache.CacheData, error) {
	return loadServerDataWithConfig(ctx, serverURL, transportType, authToken, clientName)
}

// loadServerDataWithConfig loads data with specific server configuration
func loadServerDataWithConfig(ctx context.Context, srvURL, transport, token, client string) (*cache.CacheData, error) {
	c := cache.New(srvURL, transport, token, client)

	var cachedData *cache.CacheData
	if data, _, _ := c.Load(); data != nil {
		cachedData = data
	}
	session, err := createSession(ctx, transport, srvURL, proxyURL, token, client)
	if err != nil {
		if cachedData != nil {
			fmt.Fprintf(os.Stderr, "Warning: Using cached data (server unavailable)\n")
			return cachedData, nil
		}
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	freshData, err := fetchAllServerData(ctx, session)
	if err == nil && freshData != nil {
		_ = c.Save(freshData)
		return freshData, nil
	}

	if cachedData != nil {
		return cachedData, nil
	}

	return nil, fmt.Errorf("no data available")
}

// displayData outputs the specified data type from cache data
func displayData(data *cache.CacheData, listType string) error {
	switch listType {
	case "tools":
		outputSlice(data.Tools, "tool")
	case "resources":
		outputSlice(data.Resources, "resource")
	case "prompts":
		outputSlice(data.Prompts, "prompt")
	case "all":
		outputSlice(data.Tools, "tool")
		outputSlice(data.Resources, "resource")
		outputSlice(data.Prompts, "prompt")
	default:
		return fmt.Errorf("unknown list type '%s', supported types: tools, resources, prompts", listType)
	}
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	data, err := loadServerData(ctx)
	if err != nil {
		return err
	}

	listType := "all"
	if len(args) > 0 {
		listType = args[0]
	}

	return displayData(data, listType)
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

// outputSlice converts a typed slice to []any and outputs it
func outputSlice[T any](items []T, itemType string) {
	converted := make([]any, len(items))
	for i, item := range items {
		converted[i] = item
	}
	outputItems(converted, itemType)
}
