package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"mcpmap/cache"
)

var (
	serverURL     string
	transportType string
	proxyURL      string
	authToken     string
	clientName    string
)

var rootCmd = &cobra.Command{
	Use:   "mcpmap [--sse=|--http=]<server-uri> [command]",
	Short: "A command-line tool for interacting with MCP servers",
	Long: `mcpmap is a command-line tool for interacting with Model Context Protocol (MCP) servers.
It supports both SSE (Server-Sent Events) and Streamable HTTP transport options.`,
}

func validateFlags(cmd *cobra.Command, args []string) error {
	config, err := parseTransportFlags(cmd)
	if err != nil {
		return err
	}

	if config == nil {
		return nil
	}
	transportType = config.transportType
	serverURL = config.serverURL

	return nil
}

type transportConfig struct {
	transportType string
	serverURL     string
}

func parseTransportFlags(cmd *cobra.Command) (*transportConfig, error) {
	if cmd.Name() == "completion" || cmd.Name() == "__complete" ||
		cmd.Name() == "__completeNoDesc" || cmd.Name() == "cache" ||
		cmd.Name() == "clear" || cmd.Name() == "info" {
		return nil, nil
	}

	sseFlag := cmd.Flag("sse")
	httpFlag := cmd.Flag("http")

	if sseFlag.Changed && httpFlag.Changed {
		return nil, fmt.Errorf("cannot specify both --sse and --http flags")
	}

	if sseFlag.Changed {
		return &transportConfig{"sse", sseFlag.Value.String()}, nil
	}
	if httpFlag.Changed {
		return &transportConfig{"http", httpFlag.Value.String()}, nil
	}

	return nil, fmt.Errorf("must specify either --sse=<url> or --http=<url>")
}

// createCompletionCommand creates the completion command
func createCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}

func init() {
	log.SetOutput(os.Stderr)
	rootCmd.PersistentFlags().
		StringVar(&serverURL, "sse", "", "Use SSE transport with the specified server URL")
	rootCmd.PersistentFlags().
		StringVar(&serverURL, "http", "", "Use HTTP transport with the specified server URL")
	rootCmd.PersistentFlags().
		StringVar(&proxyURL, "proxy", "", "HTTP proxy URL (e.g., http://proxy.example.com:8080)")
	rootCmd.PersistentFlags().
		StringVar(&authToken, "token", "", "Bearer token for authentication")
	rootCmd.PersistentFlags().
		StringVarP(&clientName, "name", "n", "mcpmap", "Client name to send in MCP initialize request")

	rootCmd.PersistentPreRunE = validateFlags
	rootCmd.AddCommand(createCompletionCommand())
	rootCmd.AddCommand(createCacheCommand())
}

// createCacheCommand creates the cache management command
func createCacheCommand() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage mcpmap cache",
		Long:  "Commands to manage the mcpmap cache system for faster tab completion and server metadata access.",
	}

	cacheClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all cache entries",
		Long:  "Remove all cached server metadata to force fresh queries on next access.",
		RunE:  runCacheClear,
	}

	cacheInfoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show cache statistics",
		Long:  "Display information about cached server metadata including file sizes and entry counts.",
		RunE:  runCacheInfo,
	}

	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	return cacheCmd
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	err := cache.ClearAll()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("Cache cleared successfully")
	return nil
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
	info, err := cache.GetCacheInfo()
	if err != nil {
		return fmt.Errorf("failed to get cache info: %w", err)
	}

	if info.TotalFiles == 0 {
		fmt.Println("Cache is empty")
		fmt.Printf("Cache directory: %s\n", info.CacheDir)
		return nil
	}

	fmt.Printf("Cache directory: %s\n", info.CacheDir)
	fmt.Printf("Total files: %d\n", info.TotalFiles)
	fmt.Printf("Total size: %d bytes (%.2f KB)\n", info.TotalSize, float64(info.TotalSize)/1024)
	fmt.Println()

	if len(info.Files) > 0 {
		fmt.Println("Cache entries:")
		for _, file := range info.Files {
			fmt.Printf("  %s:\n", file.Name)
			fmt.Printf("    Size: %d bytes\n", file.Size)
			fmt.Printf("    Modified: %s\n", file.ModTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Tools: %d, Resources: %d, Prompts: %d\n",
				file.ToolsCount, file.ResourcesCount, file.PromptsCount)
			fmt.Println()
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
