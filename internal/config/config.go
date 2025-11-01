package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	SandboxRoot      string
	AllowInsecureTLS bool
	Quiet            bool
	Verbose          bool

	ShowVersion bool
	ShowHelp    bool
	EvalExpr    string
	CheckOnly   bool
	ScriptFile  string
	Args        []string

	NoHistory   bool
	HistoryFile string
	Prompt      string
	NoWelcome   bool
	TraceOn     bool

	NoPrint   bool
	ReadStdin bool
	Profile   bool
}

func NewConfig() *Config {
	return &Config{
		SandboxRoot: "",
		HistoryFile: "",
		Prompt:      "",
	}
}

func (c *Config) LoadFromEnv() error {
	if root := os.Getenv("VIRO_SANDBOX_ROOT"); root != "" {
		c.SandboxRoot = root
	}

	if tls := os.Getenv("VIRO_ALLOW_INSECURE_TLS"); tls == "1" || tls == "true" {
		c.AllowInsecureTLS = true
	}

	if history := os.Getenv("VIRO_HISTORY_FILE"); history != "" {
		c.HistoryFile = history
	}

	return nil
}

func (c *Config) LoadFromFlags() error {
	return c.LoadFromFlagsWithArgs(os.Args[1:])
}

func (c *Config) LoadFromFlagsWithArgs(args []string) error {
	fs := flag.NewFlagSet("viro", flag.ContinueOnError)

	sandboxRoot := fs.String("sandbox-root", "", "Sandbox root directory for file operations (default: current directory)")
	allowInsecureTLS := fs.Bool("allow-insecure-tls", false, "Allow insecure TLS connections globally (warning: disables certificate verification)")
	quiet := fs.Bool("quiet", false, "Suppress non-error output")
	verbose := fs.Bool("verbose", false, "Enable verbose output")

	version := fs.Bool("version", false, "Show version information")
	help := fs.Bool("help", false, "Show help information")
	evalExpr := fs.String("c", "", "Evaluate expression and print result")
	check := fs.Bool("check", false, "Check syntax only (don't execute)")

	noHistory := fs.Bool("no-history", false, "Disable command history in REPL")
	historyFile := fs.String("history-file", "", "History file location")
	prompt := fs.String("prompt", "", "Custom REPL prompt")
	noWelcome := fs.Bool("no-welcome", false, "Skip welcome message in REPL")
	traceOn := fs.Bool("trace", false, "Start REPL with tracing enabled")

	noPrint := fs.Bool("no-print", false, "Don't print result of evaluation")
	stdin := fs.Bool("stdin", false, "Read additional input from stdin")
	profileFlag := fs.Bool("profile", false, "Show execution profile after script execution")

	parsed := splitCommandLineArgs(args)

	var flagArgs []string
	if parsed.ReplArgsIdx >= 0 {
		flagArgs = args[:parsed.ReplArgsIdx]
		c.Args = args[parsed.ReplArgsIdx+1:]
		c.ScriptFile = ""
	} else if parsed.ScriptIdx >= 0 {
		flagArgs = args[:parsed.ScriptIdx]
		parsed.ScriptArgs = args[parsed.ScriptIdx:]
	} else {
		flagArgs = args
		parsed.ScriptArgs = nil
	}

	if err := fs.Parse(flagArgs); err != nil {
		return err
	}

	if *sandboxRoot != "" {
		c.SandboxRoot = *sandboxRoot
	}
	c.AllowInsecureTLS = c.AllowInsecureTLS || *allowInsecureTLS
	c.Quiet = *quiet
	c.Verbose = *verbose

	c.ShowVersion = *version
	c.ShowHelp = *help
	c.EvalExpr = *evalExpr
	c.CheckOnly = *check

	c.NoHistory = *noHistory
	if *historyFile != "" {
		c.HistoryFile = *historyFile
	}
	if *prompt != "" {
		c.Prompt = *prompt
	}
	c.NoWelcome = *noWelcome
	c.TraceOn = *traceOn

	c.NoPrint = *noPrint
	c.ReadStdin = *stdin
	c.Profile = *profileFlag

	if parsed.ReplArgsIdx < 0 && len(parsed.ScriptArgs) > 0 {
		c.ScriptFile = parsed.ScriptArgs[0]
		c.Args = parsed.ScriptArgs[1:]
	}

	return nil
}

func (c *Config) ApplyDefaults() error {
	if c.SandboxRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
		c.SandboxRoot = cwd
	}
	return nil
}

func (c *Config) Validate() error {
	if c.CheckOnly && c.ScriptFile == "" {
		return fmt.Errorf("--check flag requires a script file")
	}
	if c.ReadStdin && c.EvalExpr == "" {
		return fmt.Errorf("--stdin flag requires -c flag")
	}
	if c.NoPrint && c.EvalExpr == "" {
		return fmt.Errorf("--no-print flag requires -c flag")
	}
	if c.Profile && c.ScriptFile == "" {
		return fmt.Errorf("--profile flag requires a script file")
	}
	return nil
}

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

func (c *Config) DetectMode() (Mode, error) {
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

func ParseSimple(args []string) (*Config, error) {
	cwd, _ := os.Getwd()
	cfg := &Config{
		SandboxRoot: cwd,
	}

	fs := flag.NewFlagSet("viro", flag.ContinueOnError)

	sandboxRoot := fs.String("sandbox-root", "", "")
	allowInsecureTLS := fs.Bool("allow-insecure-tls", false, "")
	quiet := fs.Bool("quiet", false, "")
	verbose := fs.Bool("verbose", false, "")
	version := fs.Bool("version", false, "")
	help := fs.Bool("help", false, "")
	evalExpr := fs.String("c", "", "")
	check := fs.Bool("check", false, "")
	noHistory := fs.Bool("no-history", false, "")
	historyFile := fs.String("history-file", "", "")
	prompt := fs.String("prompt", "", "")
	noWelcome := fs.Bool("no-welcome", false, "")
	traceOn := fs.Bool("trace", false, "")
	noPrint := fs.Bool("no-print", false, "")
	stdin := fs.Bool("stdin", false, "")
	profileFlag := fs.Bool("profile", false, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *sandboxRoot != "" {
		cfg.SandboxRoot = *sandboxRoot
	}
	cfg.AllowInsecureTLS = *allowInsecureTLS
	cfg.Quiet = *quiet
	cfg.Verbose = *verbose
	cfg.ShowVersion = *version
	cfg.ShowHelp = *help
	cfg.EvalExpr = *evalExpr
	cfg.CheckOnly = *check
	cfg.NoHistory = *noHistory
	if *historyFile != "" {
		cfg.HistoryFile = *historyFile
	}
	if *prompt != "" {
		cfg.Prompt = *prompt
	}
	cfg.NoWelcome = *noWelcome
	cfg.TraceOn = *traceOn
	cfg.NoPrint = *noPrint
	cfg.ReadStdin = *stdin
	cfg.Profile = *profileFlag

	positionalArgs := fs.Args()
	if len(positionalArgs) > 0 {
		cfg.ScriptFile = positionalArgs[0]
		if len(positionalArgs) > 1 {
			cfg.Args = positionalArgs[1:]
		}
	}

	return cfg, nil
}
