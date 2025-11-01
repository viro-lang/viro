package main

import (
	"fmt"
	"github.com/marcin-radoszewski/viro/internal/api"
	"github.com/marcin-radoszewski/viro/internal/config"
)

func main() {
	cfg := config.NewConfig()
	cfg.ScriptFile = "test_stepping.viro"
	
	ctx := &api.RuntimeContext{
		Args:   []string{},
		Stdin:  nil,
		Stdout: nil,
		Stderr: nil,
	}
	
	exitCode := api.Run(ctx, cfg)
	fmt.Printf("Exit code: %d
", exitCode)
}
