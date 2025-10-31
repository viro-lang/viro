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

func detectAndValidateMode(cfg *Config) (Mode, error) {
	modeCount := 0
	var detectedMode Mode

	if cfg.ShowVersion {
		modeCount++
		detectedMode = ModeVersion
	}
	if cfg.ShowHelp {
		modeCount++
		detectedMode = ModeHelp
	}
	if cfg.EvalExpr != "" {
		modeCount++
		detectedMode = ModeEval
	}
	if cfg.CheckOnly {
		modeCount++
		detectedMode = ModeCheck
		if cfg.ScriptFile == "" {
			return ModeCheck, fmt.Errorf("--check flag requires a script file")
		}
	} else if cfg.ScriptFile != "" {
		modeCount++
		detectedMode = ModeScript
	}

	if modeCount > 1 {
		return ModeREPL, fmt.Errorf("multiple modes specified; use only one of: --version, --help, -c, or script file")
	}

	if cfg.ReadStdin && cfg.EvalExpr == "" {
		return ModeREPL, fmt.Errorf("--stdin flag requires -c flag")
	}

	if cfg.NoPrint && cfg.EvalExpr == "" {
		return ModeREPL, fmt.Errorf("--no-print flag requires -c flag")
	}

	if modeCount == 0 {
		return ModeREPL, nil
	}

	return detectedMode, nil
}

func detectMode(cfg *Config) (Mode, error) {
	return detectAndValidateMode(cfg)
}
