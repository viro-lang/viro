package main

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/api"
)

type Mode int

const (
	ModeREPL Mode = iota
	ModeScript
	ModeEval
	ModeCheck
	ModeVersion
	ModeHelp
)

func (m Mode) String() string {
	switch m {
	case ModeREPL:
		return "REPL"
	case ModeScript:
		return "Script"
	case ModeEval:
		return "Eval"
	case ModeCheck:
		return "Check"
	case ModeVersion:
		return "Version"
	case ModeHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

func DetectMode(c *api.Config) (Mode, error) {
	modes := []struct {
		condition bool
		mode      Mode
	}{
		{c.ShowVersion, ModeVersion},
		{c.ShowHelp, ModeHelp},
		{c.EvalExpr != "", ModeEval},
		{c.CheckOnly, ModeCheck},
		{!c.CheckOnly && c.ScriptFile != "", ModeScript},
	}

	var detectedMode Mode
	modeCount := 0

	for _, m := range modes {
		if m.condition {
			modeCount++
			detectedMode = m.mode
		}
	}

	if modeCount > 1 {
		return ModeREPL, fmt.Errorf("multiple modes specified; use only one of: --version, --help, -c, or script file")
	}

	if modeCount == 0 {
		return ModeREPL, nil
	}

	return detectedMode, nil
}
