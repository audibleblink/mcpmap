package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL     string
	transportType string = "sse" // Default transport type
)

var rootCmd = &cobra.Command{
	Use:   "mcpmap [--sse=|--http=]<server-uri> [command]",
	Short: "A command-line tool for interacting with MCP servers",
	Long: `mcpmap is a command-line tool for interacting with Model Context Protocol (MCP) servers.
It supports both SSE (Server-Sent Events) and Streamable HTTP transport options.`,
}

func init() {
	// Set up logging to stderr
	log.SetOutput(os.Stderr)

	// Global flags for transport configuration
	rootCmd.PersistentFlags().
		StringVar(&serverURL, "sse", "", "Use SSE transport with the specified server URL")
	rootCmd.PersistentFlags().
		StringVar(&serverURL, "http", "", "Use HTTP transport with the specified server URL")

	// Custom validation for mutually exclusive flags
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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

	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(execCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

