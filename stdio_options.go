package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

type stdioOptions struct {
	inspect      bool
	connectionID string
	timeout      time.Duration
}

func parseStdioOptions(args []string) (stdioOptions, error) {
	var opts stdioOptions
	flags := flag.NewFlagSet("dbridge --mcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&opts.inspect, "inspect", false, "pin a connection and restrict MCP to inspection")
	flags.StringVar(&opts.connectionID, "connection", "", "connection profile ID (requires --inspect)")
	flags.DurationVar(&opts.timeout, "timeout", 30*time.Second, "maximum MCP operation duration")
	filtered := make([]string, 0, len(args))
	modeIndex := stdioModeIndex(args)
	for i, arg := range args {
		if i != modeIndex {
			filtered = append(filtered, arg)
		}
	}
	if err := flags.Parse(filtered); err != nil {
		return opts, err
	}
	if flags.NArg() != 0 {
		return opts, fmt.Errorf("unexpected STDIO argument: %s", flags.Arg(0))
	}
	if opts.connectionID != "" && !opts.inspect {
		return opts, fmt.Errorf("--connection requires --inspect")
	}
	if opts.timeout <= 0 || opts.timeout > 10*time.Minute {
		return opts, fmt.Errorf("--timeout must be greater than zero and at most 10m")
	}
	return opts, nil
}

// Do not mistake a profile ID or timeout argument for a transport selector.
func stdioModeIndex(args []string) int {
	for i := 0; i < len(args); i++ {
		switch strings.ToLower(strings.TrimSpace(args[i])) {
		case "--connection", "-connection", "--timeout", "-timeout":
			i++
		case "--mcp", "-mcp", "mcp", "stdio", "--stdio", "-stdio", "/mcp":
			return i
		}
	}
	return -1
}
