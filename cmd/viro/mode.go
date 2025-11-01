package main

import "fmt"

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

func detectMode(cfg *Config) (Mode, error) {
	modeCount := 0
	var detectedMode Mode

	modes := []struct {
		condition bool
		mode      Mode
	}{
		{cfg.ShowVersion, ModeVersion},
		{cfg.ShowHelp, ModeHelp},
		{cfg.EvalExpr != "", ModeEval},
		{cfg.CheckOnly, ModeCheck},
		{!cfg.CheckOnly && cfg.ScriptFile != "", ModeScript},
	}

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
