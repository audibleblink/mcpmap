//go:build linux

package cache

import (
	"os"
	"path/filepath"
)

func getCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "mcpmap")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "mcpmap")
}