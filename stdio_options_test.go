package main

import (
	"testing"
	"time"
)

func TestStdioOptions(t *testing.T) {
	if index := stdioModeIndex([]string{"--connection", "mcp"}); index != -1 {
		t.Fatal("profile value mistaken for transport")
	}
	if opts, err := parseStdioOptions([]string{"--inspect", "--connection", "mcp", "stdio"}); err != nil || opts.connectionID != "mcp" {
		t.Fatalf("trailing transport: %+v, %v", opts, err)
	}
	for _, mode := range []string{"--mcp", "stdio", "--stdio", "/mcp"} {
		opts, err := parseStdioOptions([]string{mode, "--inspect", "--connection", "mcp", "--timeout", "2s"})
		if err != nil || !opts.inspect || opts.connectionID != "mcp" || opts.timeout != 2*time.Second {
			t.Fatalf("%s: %+v, %v", mode, opts, err)
		}
	}
	for _, args := range [][]string{
		{"--mcp", "--connection", "id"}, {"--mcp", "--timeout", "0s"}, {"--mcp", "--timeout", "11m"}, {"stdio", "unexpected"}, {"--mcp", "--unknown"},
	} {
		if _, err := parseStdioOptions(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
