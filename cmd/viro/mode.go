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
	if cfg.ShowVersion {
		return ModeVersion, nil
	}
	if cfg.ShowHelp {
		return ModeHelp, nil
	}
	if cfg.EvalExpr != "" {
		return ModeEval, nil
	}
	if cfg.CheckOnly {
		if cfg.ScriptFile == "" {
			return ModeCheck, fmt.Errorf("--check requires a script file")
		}
		return ModeCheck, nil
	}
	if cfg.ScriptFile != "" {
		return ModeScript, nil
	}
	return ModeREPL, nil
}
