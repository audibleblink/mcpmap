package main

import (
	"fmt"

	"mcpmap/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage mcpmap cache",
	Long:  "Commands to manage the mcpmap cache system for faster tab completion and server metadata access.",
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cache entries",
	Long:  "Remove all cached server metadata to force fresh queries on next access.",
	RunE:  runCacheClear,
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cache statistics",
	Long:  "Display information about cached server metadata including file sizes and entry counts.",
	RunE:  runCacheInfo,
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	rootCmd.AddCommand(cacheCmd)
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
