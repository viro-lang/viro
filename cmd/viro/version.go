package main

import "fmt"

const (
	Version   = "0.1.0"
	BuildDate = ""
)

func printVersion() {
	if BuildDate != "" {
		fmt.Printf("Viro %s (built %s)\n", Version, BuildDate)
	} else {
		fmt.Printf("Viro %s\n", Version)
	}
}
