//go:build windows

package cache

import (
	"os"
	"path/filepath"
)

func getCacheDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "mcpmap", "cache")
}
