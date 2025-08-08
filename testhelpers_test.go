package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// testHelper provides common test utilities
type testHelper struct {
	t *testing.T
}

func newTestHelper(t *testing.T) *testHelper {
	return &testHelper{t: t}
}

// createCmdWithFlags creates a command with standard transport flags
func (h *testHelper) createCmdWithFlags() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("sse", "", "")
	cmd.Flags().String("http", "", "")
	return cmd
}

// setTransportFlag sets either sse or http flag on a command
func (h *testHelper) setTransportFlag(cmd *cobra.Command, transport, url string) {
	switch transport {
	case "sse":
		cmd.Flags().Set("sse", url)
	case "http":
		cmd.Flags().Set("http", url)
	}
}

// captureOutput captures stdout during function execution
func (h *testHelper) captureOutput(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	fn()
	w.Close()

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// assertStringContains checks if output contains expected strings
func (h *testHelper) assertStringContains(output string, expected []string) {
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			h.t.Errorf("expected output to contain %q, but got: %q", exp, output)
		}
	}
}
