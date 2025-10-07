package main

import (
	"fmt"
	"os"
)

func main() {
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
