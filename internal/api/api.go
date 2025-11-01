package api

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/profile"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

const (
	defaultTraceMaxSizeMB = 50
)

type RuntimeContext struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
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

type Config struct {
	SandboxRoot      string
	AllowInsecureTLS bool
	Quiet            bool
	Verbose          bool
	ShowVersion      bool
	ShowHelp         bool
	EvalExpr         string
	CheckOnly        bool
	ScriptFile       string
	Args             []string
	NoHistory        bool
	HistoryFile      string
	Prompt           string
	NoWelcome        bool
	TraceOn          bool
	NoPrint          bool
	ReadStdin        bool
	Profile          bool
}

func NewConfig() *Config {
	cwd, _ := os.Getwd()
	return &Config{
		SandboxRoot: cwd,
	}
}

func ConfigFromArgs(args []string) (*Config, error) {
	cfg := NewConfig()

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
	if c.Profile {
		if c.EvalExpr != "" {
			return fmt.Errorf("--profile flag cannot be used with -c expressions; it requires a script file or '-' for stdin")
		}
		if c.ScriptFile == "" {
			return fmt.Errorf("--profile flag requires a script file or '-' for stdin")
		}
	}
	return nil
}

type InputSource interface {
	Load() (string, error)
}

type ExprInputWithContext struct {
	Expr      string
	WithStdin bool
	Stdin     io.Reader
}

func (e *ExprInputWithContext) Load() (string, error) {
	expr := e.Expr

	if e.WithStdin {
		stdinData, err := io.ReadAll(e.Stdin)
		if err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
		expr = string(stdinData) + "\n" + expr
	}

	return expr, nil
}

type FileInput struct {
	Config *Config
	Path   string
	Stdin  io.Reader
}

func (f *FileInput) Load() (string, error) {
	if f.Path == "-" {
		stdin := f.Stdin
		if stdin == nil {
			stdin = os.Stdin
		}
		data, err := io.ReadAll(stdin)
		return string(data), err
	}

	fullPath := f.Path
	if !filepath.IsAbs(f.Path) && f.Config != nil {
		fullPath = filepath.Join(f.Config.SandboxRoot, f.Path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", f.Path, err)
	}

	return string(data), nil
}

func NewFileInput(cfg *Config, path string, stdin io.Reader) InputSource {
	return &FileInput{Config: cfg, Path: path, Stdin: stdin}
}

func Run(ctx *RuntimeContext, cfg *Config) int {
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(ctx.Stderr, "Configuration error: %v\n", err)
		return ExitUsage
	}

	mode := detectMode(cfg)

	switch mode {
	case ModeREPL:
		fmt.Fprintf(ctx.Stderr, "REPL mode not supported in API context\n")
		return ExitError
	case ModeScript, ModeEval, ModeCheck:
		return runExecutionWithContext(cfg, mode, ctx)
	case ModeVersion:
		fmt.Fprintf(ctx.Stdout, "%s\n", "Viro 0.1.0")
		return ExitSuccess
	case ModeHelp:
		fmt.Fprintf(ctx.Stdout, "%s", "Viro help text")
		return ExitSuccess
	default:
		fmt.Fprintf(ctx.Stderr, "Unknown mode: %v\n", mode)
		return ExitUsage
	}
}

func detectMode(cfg *Config) Mode {
	if cfg.ShowVersion {
		return ModeVersion
	}
	if cfg.ShowHelp {
		return ModeHelp
	}
	if cfg.EvalExpr != "" {
		return ModeEval
	}
	if cfg.CheckOnly {
		return ModeCheck
	}
	if cfg.ScriptFile != "" {
		return ModeScript
	}
	return ModeREPL
}

func runExecutionWithContext(cfg *Config, mode Mode, ctx *RuntimeContext) int {
	var err error
	if cfg.Profile {
		err = trace.InitTraceSilent()
	} else {
		err = trace.InitTrace("", defaultTraceMaxSizeMB)
	}

	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error initializing trace: %v\n", err)
		return ExitInternal
	}

	var profiler *profile.Profiler
	if cfg.Profile && trace.GlobalTraceSession != nil {
		profiler = profile.NewProfiler()
		profile.EnableProfilingWithTrace(trace.GlobalTraceSession, profiler)
	}

	var input InputSource

	switch mode {
	case ModeCheck:
		input = NewFileInput(cfg, cfg.ScriptFile, ctx.Stdin)
	case ModeEval:
		input = &ExprInputWithContext{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin, Stdin: ctx.Stdin}
	case ModeScript:
		input = NewFileInput(cfg, cfg.ScriptFile, ctx.Stdin)
	}

	var args []string
	if mode == ModeScript {
		args = cfg.Args
	} else {
		args = []string{}
	}

	printResult := (mode == ModeEval && !cfg.NoPrint)
	parseOnly := (mode == ModeCheck)

	exitCode := executeViroCodeWithContext(cfg, input, args, printResult, parseOnly, ctx)

	if profiler != nil {
		profiler.Disable()
		if !cfg.Quiet {
			report := profiler.GetReport()
			report.FormatText(ctx.Stderr)
		}
	}

	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Close()
	}

	return exitCode
}

func executeViroCodeWithContext(cfg *Config, input InputSource, args []string, printResult bool, parseOnly bool, ctx *RuntimeContext) int {
	content, err := input.Load()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error loading input: %v\n", err)
		return ExitError
	}

	values, err := parse.Parse(content)
	if err != nil {
		printErrorToWriter(err, "Parse", ctx.Stderr)
		return ExitSyntax
	}

	if parseOnly {
		if cfg.Verbose {
			fmt.Fprintf(ctx.Stdout, "✓ Syntax valid\n")
			fmt.Fprintf(ctx.Stdout, "  Parsed %d expressions\n", len(values))
		}
		return ExitSuccess
	}

	evaluator := setupEvaluatorWithContext(cfg, ctx)
	initializeSystemObjectInEvaluator(evaluator, args)

	result, err := evaluator.DoBlock(values)
	if err != nil {
		printErrorToWriter(err, "Runtime", ctx.Stderr)
		return handleErrorWithContext(err)
	}

	if printResult && !cfg.Quiet {
		fmt.Fprintln(ctx.Stdout, result.Form())
	}

	return ExitSuccess
}

func setupEvaluatorWithContext(cfg *Config, ctx *RuntimeContext) *eval.Evaluator {
	evaluator := eval.NewEvaluator()

	if cfg.Quiet {
		evaluator.SetOutputWriter(io.Discard)
	} else {
		evaluator.SetOutputWriter(ctx.Stdout)
	}
	evaluator.SetErrorWriter(ctx.Stderr)
	evaluator.SetInputReader(ctx.Stdin)

	rootFrame := evaluator.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, evaluator)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	return evaluator
}

func initializeSystemObjectInEvaluator(evaluator *eval.Evaluator, args []string) {
	viroArgs := make([]core.Value, len(args))
	for i, arg := range args {
		viroArgs[i] = value.NewStringValue(arg)
	}

	argsBlock := value.NewBlockValue(viroArgs)

	ownedFrame := frame.NewFrame(frame.FrameObject, -1)
	ownedFrame.Bind("args", argsBlock)

	systemObj := value.NewObject(ownedFrame)

	rootFrame := evaluator.GetFrameByIndex(0)
	rootFrame.Bind("system", systemObj)
}

func handleErrorWithContext(err error) int {
	if err == nil {
		return ExitSuccess
	}

	if vErr, ok := err.(*verror.Error); ok {
		return categoryToExitCode(vErr.Category)
	}

	return ExitError
}

func printErrorToWriter(err error, prefix string, w io.Writer) {
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintf(w, "%v", vErr)
	} else if prefix != "" {
		fmt.Fprintf(w, "%s error: %v\n", prefix, err)
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}

func categoryToExitCode(cat verror.ErrorCategory) int {
	switch cat {
	case verror.ErrSyntax:
		return ExitSyntax
	case verror.ErrAccess:
		return ExitAccess
	case verror.ErrInternal:
		return ExitInternal
	default:
		return ExitError
	}
}

const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitSyntax    = 2
	ExitAccess    = 3
	ExitUsage     = 64
	ExitInternal  = 70
	ExitInterrupt = 130
)
