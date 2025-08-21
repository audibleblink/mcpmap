package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Cache provides file-based caching for MCP server metadata
type // Cache provides file-based caching for MCP server metadata (tools, resources, prompts).
// Implementations should be safe for concurrent reads and tolerate concurrent writes.
Cache interface {
	// Load retrieves cached data, returns (data, isFresh, error)
	// isFresh is always true since we don't use TTL
	Load() (*CacheData, bool, error)

	// Save stores data to cache
	Save(data *CacheData) error

	// Delete removes this cache entry
	Delete() error
}

// CacheData represents the cached MCP server information
type // CacheData holds the MCP server metadata cached for faster subsequent access.
CacheData struct {
	Tools     []*mcp.Tool     `json:"tools"`
	Resources []*mcp.Resource `json:"resources"`
	Prompts   []*mcp.Prompt   `json:"prompts"`
}

// cacheFile represents the structure of the cache file on disk
type cacheFile struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"server_info"`
	Data *CacheData `json:"data"`
}

// fileCache implements Cache using filesystem storage
type fileCache struct {
	cacheKey string
	cacheDir string
	filePath string
}

// New creates a cache instance for the given server configuration
// New returns a filesystem-backed Cache keyed by the supplied server connection parameters.
func New(serverURL, transportType, authToken, clientName string) Cache {
	cacheKey := generateCacheKey(serverURL, transportType, authToken, clientName)
	cacheDir := getCacheDir()
	filePath := filepath.Join(cacheDir, cacheKey+".json")

	return &fileCache{
		cacheKey: cacheKey,
		cacheDir: cacheDir,
		filePath: filePath,
	}
}

// ensureDir creates the cache directory if it doesn't exist
func (fc *fileCache) ensureDir() error {
	if err := os.MkdirAll(fc.cacheDir, 0700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	return nil
}

// generateCacheKey creates a unique cache key from server configuration
func generateCacheKey(serverURL, transportType, authToken, clientName string) string {
	h := sha256.New()
	h.Write([]byte(serverURL))
	h.Write([]byte(transportType))
	h.Write([]byte(authToken))
	h.Write([]byte(clientName))
	return hex.EncodeToString(h.Sum(nil))[:16] // First 16 chars
}

// Load retrieves cached data from disk
// Load retrieves cached data from disk. The isFresh return value is always true
// on a successful hit because the current implementation has no TTL or staleness checks.
func (fc *fileCache) Load() (*CacheData, bool, error) {
	// Ensure cache directory exists
	if err := fc.ensureDir(); err != nil {
		return nil, false, err
	}

	// Read file
	data, err := os.ReadFile(fc.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil // Cache miss
		}
		return nil, false, fmt.Errorf("read cache file: %w", err)
	}

	// Parse JSON
	var cf cacheFile
	if err := json.Unmarshal(data, &cf); err != nil {
		// Corrupted cache, delete and return miss
		os.Remove(fc.filePath)
		return nil, false, nil
	}

	// Version check
	if cf.Version != 1 {
		// Old version, delete and return miss
		os.Remove(fc.filePath)
		return nil, false, nil
	}

	// isFresh is always true since we don't implement TTL
	return cf.Data, true, nil
}

// Save stores data to cache using atomic writes
func (fc *fileCache) Save(data *CacheData) error {
	// Ensure cache directory exists with secure permissions
	if err := fc.ensureDir(); err != nil {
		return err
	}

	cf := cacheFile{
		Version:   1,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache data: %w", err)
	}

	// Write atomically via temp file
	tmpFile := fc.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0600); err != nil {
		return fmt.Errorf("write temp cache file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, fc.filePath); err != nil {
		os.Remove(tmpFile) // Cleanup
		return fmt.Errorf("rename cache file: %w", err)
	}

	return nil
}

// Delete removes the cache file
func (fc *fileCache) Delete() error {
	err := os.Remove(fc.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete cache file: %w", err)
	}
	return nil
}

// readCacheCounts reads a cache file and returns the count of tools, resources, and prompts
func readCacheCounts(filePath string) (tools, resources, prompts int) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, 0, 0
	}
	
	var cf cacheFile
	if err := json.Unmarshal(data, &cf); err != nil || cf.Data == nil {
		return 0, 0, 0
	}
	
	return len(cf.Data.Tools), len(cf.Data.Resources), len(cf.Data.Prompts)
}

// ClearAll removes all cache files from the cache directory
// ClearAll removes all cache entry files from the user cache directory.
// It ignores missing directories and returns an error if any file removal fails.
func ClearAll() error {
	cacheDir := getCacheDir()
	
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil // Nothing to clear
	}
	
	// Read directory contents
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("read cache directory: %w", err)
	}
	
	// Remove all .json files (cache files)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filePath := filepath.Join(cacheDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("remove cache file %s: %w", entry.Name(), err)
			}
		}
	}
	
	return nil
}

// CacheInfo represents information about the cache
type // CacheInfo summarizes the contents of the cache directory including total counts and per-file metadata.
CacheInfo struct {
	CacheDir   string      `json:"cache_dir"`
	TotalFiles int         `json:"total_files"`
	TotalSize  int64       `json:"total_size_bytes"`
	Files      []FileInfo  `json:"files"`
}

// FileInfo represents information about a single cache file
type // FileInfo describes a single cache file's size, modification time, and contained item counts.
FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size_bytes"`
	ModTime      time.Time `json:"modified_time"`
	ToolsCount   int       `json:"tools_count"`
	ResourcesCount int     `json:"resources_count"`
	PromptsCount int       `json:"prompts_count"`
}

// GetCacheInfo returns information about all cache files
// GetCacheInfo returns information about all cache files in the cache directory.
// Files that cannot be read or parsed are skipped silently.
func GetCacheInfo() (*CacheInfo, error) {
	cacheDir := getCacheDir()
	
	info := &CacheInfo{
		CacheDir: cacheDir,
		Files:    []FileInfo{},
	}
	
	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return info, nil // Empty cache info
	}
	
	// Read directory contents
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("read cache directory: %w", err)
	}
	
	// Process each cache file
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		
		filePath := filepath.Join(cacheDir, entry.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue // Skip files we can't stat
		}
		
		// Get item counts using helper
		toolsCount, resourcesCount, promptsCount := readCacheCounts(filePath)
		
		cacheFileInfo := FileInfo{
			Name:           entry.Name(),
			Size:           fileInfo.Size(),
			ModTime:        fileInfo.ModTime(),
			ToolsCount:     toolsCount,
			ResourcesCount: resourcesCount,
			PromptsCount:   promptsCount,
		}
		
		info.Files = append(info.Files, cacheFileInfo)
		info.TotalFiles++
		info.TotalSize += fileInfo.Size()
	}
	
	return info, nil
}