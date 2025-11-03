package main

import (
	"fmt"
	"github.com/marcin-radoszewski/viro/internal/parse"
)

func main() {
	vals, err := parse.Parse(`join "a" "b" "c"`)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	fmt.Printf("Parsed %d values:\n", len(vals))
	for i, v := range vals {
		fmt.Printf("%d: %s (type: %s)\n", i, v.Mold(), v.GetType())
	}
}
