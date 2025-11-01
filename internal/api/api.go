package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marcin-radoszewski/viro/internal/bootstrap"
	"github.com/marcin-radoszewski/viro/internal/config"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/profile"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

type RuntimeContext struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Mode = config.Mode

const (
	ModeREPL    = config.ModeREPL
	ModeScript  = config.ModeScript
	ModeEval    = config.ModeEval
	ModeCheck   = config.ModeCheck
	ModeVersion = config.ModeVersion
	ModeHelp    = config.ModeHelp
)

type Config = config.Config

func NewConfig() *Config {
	return config.NewConfig()
}

func ConfigFromArgs(args []string) (*Config, error) {
	return config.ParseSimple(args)
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
	mode := detectMode(cfg)

	switch mode {
	case ModeREPL:
		fmt.Fprintf(ctx.Stderr, "REPL mode not supported in API context\n")
		return ExitError
	case ModeScript, ModeEval, ModeCheck:
		return RunExecutionWithContext(cfg, mode, ctx)
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

func RunExecutionWithContext(cfg *Config, mode Mode, ctx *RuntimeContext) int {
	if cfg.Profile && mode == ModeEval {
		fmt.Fprintf(ctx.Stderr, "Error: --profile flag requires a script file, not -c expression\n")
		return ExitUsage
	}

	var err error
	if cfg.Profile {
		err = bootstrap.InitTrace(true)
	} else {
		err = bootstrap.InitTrace(false)
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
		return HandleErrorWithContext(err)
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

	// Initialize debugger for script execution (same as REPL)
	debug.InitDebugger()

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

func HandleErrorWithContext(err error) int {
	if err == nil {
		return ExitSuccess
	}

	if vErr, ok := err.(*verror.Error); ok {
		return verror.ToExitCode(vErr.Category)
	}

	return ExitError
}

func printErrorToWriter(err error, prefix string, w io.Writer) {
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintln(w, verror.FormatErrorWithContext(vErr))
	} else if prefix != "" {
		fmt.Fprintf(w, "%s error: %v\n", prefix, err)
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
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
