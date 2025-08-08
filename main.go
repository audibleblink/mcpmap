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
)

var rootCmd = &cobra.Command{
	Use:   "mcpmap [--sse=|--http=]<server-uri> [command]",
	Short: "A command-line tool for interacting with MCP servers",
	Long: `mcpmap is a command-line tool for interacting with Model Context Protocol (MCP) servers.
It supports both SSE (Server-Sent Events) and Streamable HTTP transport options.`,
}

func validateFlags(cmd *cobra.Command, args []string) error {
	if cmd.Name() == "completion" || cmd.Name() == "__complete" ||
		cmd.Name() == "__completeNoDesc" {
		return nil
	}

	sseFlag := cmd.Flag("sse")
	httpFlag := cmd.Flag("http")

	if sseFlag.Changed && httpFlag.Changed {
		return fmt.Errorf("cannot specify both --sse and --http flags")
	}

	if sseFlag.Changed {
		transportType = "sse"
		serverURL = sseFlag.Value.String()
	} else if httpFlag.Changed {
		transportType = "http"
		serverURL = httpFlag.Value.String()
	}

	if serverURL == "" {
		return fmt.Errorf("must specify either --sse=<url> or --http=<url>")
	}

	return nil
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

	rootCmd.PersistentPreRunE = validateFlags
	rootCmd.AddCommand(createCompletionCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
