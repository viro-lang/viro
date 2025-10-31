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
		validator func() error
	}{
		{cfg.ShowVersion, ModeVersion, nil},
		{cfg.ShowHelp, ModeHelp, nil},
		{cfg.EvalExpr != "", ModeEval, nil},
		{cfg.CheckOnly, ModeCheck, func() error {
			if cfg.ScriptFile == "" {
				return fmt.Errorf("--check flag requires a script file")
			}
			return nil
		}},
		{!cfg.CheckOnly && cfg.ScriptFile != "", ModeScript, nil},
	}

	for _, m := range modes {
		if m.condition {
			modeCount++
			detectedMode = m.mode
			if m.validator != nil {
				if err := m.validator(); err != nil {
					return detectedMode, err
				}
			}
		}
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
