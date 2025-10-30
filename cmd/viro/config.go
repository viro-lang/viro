package main

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

	NoHistory   bool
	HistoryFile string
	Prompt      string
	NoWelcome   bool

	NoPrint   bool
	ReadStdin bool
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
	sandboxRoot := flag.String("sandbox-root", "", "Sandbox root directory for file operations (default: current directory)")
	allowInsecureTLS := flag.Bool("allow-insecure-tls", false, "Allow insecure TLS connections globally (warning: disables certificate verification)")
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	version := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Show help information")
	evalExpr := flag.String("c", "", "Evaluate expression and print result")
	check := flag.Bool("check", false, "Check syntax only (don't execute)")

	noHistory := flag.Bool("no-history", false, "Disable command history in REPL")
	historyFile := flag.String("history-file", "", "History file location")
	prompt := flag.String("prompt", "", "Custom REPL prompt")
	noWelcome := flag.Bool("no-welcome", false, "Skip welcome message in REPL")

	noPrint := flag.Bool("no-print", false, "Don't print result of evaluation")
	stdin := flag.Bool("stdin", false, "Read additional input from stdin")

	flag.Parse()

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

	c.NoPrint = *noPrint
	c.ReadStdin = *stdin

	if c.SandboxRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
		c.SandboxRoot = cwd
	}

	args := flag.Args()
	if len(args) > 0 {
		c.ScriptFile = args[0]
	}

	return nil
}

func (c *Config) Validate() error {
	modeCount := 0
	if c.ShowVersion {
		modeCount++
	}
	if c.ShowHelp {
		modeCount++
	}
	if c.EvalExpr != "" {
		modeCount++
	}
	if c.CheckOnly && c.ScriptFile == "" {
		return fmt.Errorf("--check flag requires a script file")
	}
	if c.ScriptFile != "" {
		modeCount++
	}

	if modeCount > 1 {
		return fmt.Errorf("multiple modes specified; use only one of: --version, --help, -c, or script file")
	}

	if c.ReadStdin && c.EvalExpr == "" {
		return fmt.Errorf("--stdin flag requires -c flag")
	}

	if c.NoPrint && c.EvalExpr == "" {
		return fmt.Errorf("--no-print flag requires -c flag")
	}

	return nil
}
