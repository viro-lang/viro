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
	Args        []string

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

	noPrint := fs.Bool("no-print", false, "Don't print result of evaluation")
	stdin := fs.Bool("stdin", false, "Read additional input from stdin")

	args := os.Args[1:]
	scriptIdx := -1

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-c" {
			if i+1 < len(args) {
				i++
			}
			continue
		}

		if arg == "--sandbox-root" || arg == "--history-file" || arg == "--prompt" {
			if i+1 < len(args) {
				i++
			}
			continue
		}

		if !hasPrefix(arg, "-") {
			scriptIdx = i
			break
		}
	}

	var flagArgs []string
	var scriptArgs []string

	if scriptIdx >= 0 {
		flagArgs = args[:scriptIdx]
		scriptArgs = args[scriptIdx:]
	} else {
		flagArgs = args
		scriptArgs = nil
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

	c.NoPrint = *noPrint
	c.ReadStdin = *stdin

	if c.SandboxRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
		c.SandboxRoot = cwd
	}

	if len(scriptArgs) > 0 {
		c.ScriptFile = scriptArgs[0]
		c.Args = scriptArgs[1:]
	}

	return nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
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
