package main

import "fmt"

const (
	Version   = "0.1.0"
	BuildDate = ""
)

func getVersionString() string {
	if BuildDate != "" {
		return fmt.Sprintf("Viro %s (built %s)", Version, BuildDate)
	}
	return fmt.Sprintf("Viro %s", Version)
}

func printVersion() {
	fmt.Println(getVersionString())
}
