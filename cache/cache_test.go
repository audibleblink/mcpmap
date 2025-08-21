package cache

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTestCache creates a cache instance for testing with a temporary directory
func createTestCache(t *testing.T) (Cache, func()) {
	tmpDir := t.TempDir()

	// Set environment variable to use temp directory
	oldXDG := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tmpDir)

	cleanup := func() {
		if oldXDG == "" {
			os.Unsetenv("XDG_CACHE_HOME")
		} else {
			os.Setenv("XDG_CACHE_HOME", oldXDG)
		}
	}

	return New("test-url", "http", "token", "client"), cleanup
}

// createTestData creates sample cache data for testing
func createTestData() *CacheData {
	return &CacheData{
		Tools: []*mcp.Tool{
			{Name: "test-tool-1", Description: "Test tool 1"},
			{Name: "test-tool-2", Description: "Test tool 2"},
		},
		Resources: []*mcp.Resource{
			{URI: "test://resource1", Name: "Resource 1"},
			{URI: "test://resource2", Name: "Resource 2"},
		},
		Prompts: []*mcp.Prompt{
			{Name: "test-prompt-1", Description: "Test prompt 1"},
			{Name: "test-prompt-2", Description: "Test prompt 2"},
		},
	}
}

func TestCacheOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"SaveAndLoad", testSaveAndLoad},
		{"CacheMiss", testCacheMiss},
		{"Delete", testDelete},
		{"CorruptedCache", testCorruptedCache},
		{"ConcurrentAccess", testConcurrentAccess},
		{"CacheKeyGeneration", testCacheKeyGeneration},
		{"PlatformPaths", testPlatformPaths},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testSaveAndLoad(t *testing.T) {
	cache, cleanup := createTestCache(t)
	defer cleanup()

	testData := createTestData()

	// Save data
	err := cache.Save(testData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load data
	loadedData, isFresh, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !isFresh {
		t.Error("Expected isFresh to be true")
	}

	if loadedData == nil {
		t.Fatal("Loaded data is nil")
	}

	// Verify tools
	if len(loadedData.Tools) != len(testData.Tools) {
		t.Errorf("Expected %d tools, got %d", len(testData.Tools), len(loadedData.Tools))
	}

	for i, tool := range loadedData.Tools {
		if tool.Name != testData.Tools[i].Name {
			t.Errorf("Tool %d name mismatch: expected %s, got %s", i, testData.Tools[i].Name, tool.Name)
		}
	}

	// Verify resources
	if len(loadedData.Resources) != len(testData.Resources) {
		t.Errorf("Expected %d resources, got %d", len(testData.Resources), len(loadedData.Resources))
	}

	// Verify prompts
	if len(loadedData.Prompts) != len(testData.Prompts) {
		t.Errorf("Expected %d prompts, got %d", len(testData.Prompts), len(loadedData.Prompts))
	}
}

func testCacheMiss(t *testing.T) {
	cache, cleanup := createTestCache(t)
	defer cleanup()

	// Load from non-existent cache
	data, isFresh, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if data != nil {
		t.Error("Expected nil data for cache miss")
	}

	if isFresh {
		t.Error("Expected isFresh to be false for cache miss")
	}
}

func testDelete(t *testing.T) {
	cache, cleanup := createTestCache(t)
	defer cleanup()

	testData := createTestData()

	// Save data
	err := cache.Save(testData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify data exists
	data, _, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if data == nil {
		t.Fatal("Expected data to exist before delete")
	}

	// Delete cache
	err = cache.Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify data is gone
	data, isFresh, err := cache.Load()
	if err != nil {
		t.Fatalf("Load after delete failed: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data after delete")
	}
	if isFresh {
		t.Error("Expected isFresh to be false after delete")
	}
}

func testCorruptedCache(t *testing.T) {
	cache, cleanup := createTestCache(t)
	defer cleanup()

	// Get the file path by creating a fileCache instance
	fc := cache.(*fileCache)

	// Create cache directory
	err := os.MkdirAll(fc.cacheDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Write corrupted JSON
	corruptedData := []byte("{invalid json")
	err = os.WriteFile(fc.filePath, corruptedData, 0600)
	if err != nil {
		t.Fatalf("Failed to write corrupted cache: %v", err)
	}

	// Load should handle corruption gracefully
	data, isFresh, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if data != nil {
		t.Error("Expected nil data for corrupted cache")
	}

	if isFresh {
		t.Error("Expected isFresh to be false for corrupted cache")
	}

	// Verify corrupted file was cleaned up
	if _, err := os.Stat(fc.filePath); !os.IsNotExist(err) {
		t.Error("Expected corrupted cache file to be deleted")
	}
}

func testConcurrentAccess(t *testing.T) {
	cache, cleanup := createTestCache(t)
	defer cleanup()

	testData := createTestData()

	var wg sync.WaitGroup

	// First save some data
	err := cache.Save(testData)
	if err != nil {
		t.Fatalf("Initial save failed: %v", err)
	}

	// Concurrent loads (these should be safe)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := cache.Load(); err != nil {
				t.Errorf("Concurrent load error: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify final state
	data, _, err := cache.Load()
	if err != nil {
		t.Fatalf("Final load failed: %v", err)
	}
	if data == nil {
		t.Error("Expected data to exist after concurrent access")
	}
}

func testCacheKeyGeneration(t *testing.T) {
	tests := []struct {
		name       string
		serverURL  string
		transport  string
		authToken  string
		clientName string
		expectSame bool
	}{
		{
			name:       "identical configs",
			serverURL:  "http://localhost:8080",
			transport:  "http",
			authToken:  "token123",
			clientName: "client1",
			expectSame: true,
		},
		{
			name:       "different URLs",
			serverURL:  "http://localhost:8081",
			transport:  "http",
			authToken:  "token123",
			clientName: "client1",
			expectSame: false,
		},
		{
			name:       "different tokens",
			serverURL:  "http://localhost:8080",
			transport:  "http",
			authToken:  "token456",
			clientName: "client1",
			expectSame: false,
		},
	}

	baseKey := generateCacheKey("http://localhost:8080", "http", "token123", "client1")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateCacheKey(tt.serverURL, tt.transport, tt.authToken, tt.clientName)

			if tt.expectSame {
				if key != baseKey {
					t.Errorf("Expected same key, got different: %s vs %s", baseKey, key)
				}
			} else {
				if key == baseKey {
					t.Errorf("Expected different key, got same: %s", key)
				}
			}

			// Verify key length
			if len(key) != 16 {
				t.Errorf("Expected key length 16, got %d", len(key))
			}
		})
	}
}

func testPlatformPaths(t *testing.T) {
	// Test that getCacheDir returns a valid path
	cacheDir := getCacheDir()

	if cacheDir == "" {
		t.Error("getCacheDir returned empty string")
	}

	if !filepath.IsAbs(cacheDir) {
		t.Errorf("getCacheDir should return absolute path, got: %s", cacheDir)
	}

	// Test with XDG_CACHE_HOME set
	oldXDG := os.Getenv("XDG_CACHE_HOME")
	testXDG := "/tmp/test-cache"
	os.Setenv("XDG_CACHE_HOME", testXDG)

	defer func() {
		if oldXDG == "" {
			os.Unsetenv("XDG_CACHE_HOME")
		} else {
			os.Setenv("XDG_CACHE_HOME", oldXDG)
		}
	}()

	xdgCacheDir := getCacheDir()
	expectedPath := filepath.Join(testXDG, "mcpmap")

	if xdgCacheDir != expectedPath {
		t.Errorf("Expected XDG cache dir %s, got %s", expectedPath, xdgCacheDir)
	}
}

// Benchmark tests
func BenchmarkCacheSave(b *testing.B) {
	tmpDir := b.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cache := New("test-url", "http", "token", "client")
	testData := createTestData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Save(testData)
	}
}

func BenchmarkCacheLoad(b *testing.B) {
	tmpDir := b.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cache := New("test-url", "http", "token", "client")
	testData := createTestData()
	cache.Save(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Load()
	}
}
