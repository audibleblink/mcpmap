package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
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
	
	// Skip setting globals for completion commands
	if config == nil {
		return nil
	}

	// Set global state only after successful parsing
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
		cmd.Name() == "__completeNoDesc" {
		return nil, nil // Skip validation for completion commands
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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

