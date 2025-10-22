package main

import (
	"flag"
	"fmt"
	"os"
)

// CLI flags for Feature 002
var (
	sandboxRoot      = flag.String("sandbox-root", "", "Sandbox root directory for file operations (default: current directory)")
	allowInsecureTLS = flag.Bool("allow-insecure-tls", false, "Allow insecure TLS connections globally (warning: disables certificate verification)")
)

func main() {
	flag.Parse()

	// Resolve sandbox root (default to current directory per FR-006)
	if *sandboxRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		*sandboxRoot = cwd
	}

	// Warn if insecure TLS is enabled (per FR-020)
	if *allowInsecureTLS {
		fmt.Fprintf(os.Stderr, "WARNING: TLS certificate verification disabled globally. Use with caution.\n")
	}

	repl, err := NewREPL()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		os.Exit(1)
	}

	if err := repl.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running REPL: %v\n", err)
		os.Exit(1)
	}
}
