package main

import "github.com/marcin-radoszewski/viro/internal/repl"

// Deprecated: retained for backward compatibility; prefer using the internal/repl package directly.
func NewREPL() (*repl.REPL, error) {
	return repl.NewREPL()
}
