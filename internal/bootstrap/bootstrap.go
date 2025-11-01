package bootstrap

import (
	"io"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// InitTrace initializes the global trace session.
// This should be called once at application startup.
func InitTrace(profile bool) error {
	return InitTraceWithOutput(profile, "")
}

// InitTraceWithOutput initializes the global trace session with custom output.
// If output is empty, uses default stderr. If output is not empty, treats it as a file path.
func InitTraceWithOutput(profile bool, output string) error {
	if profile {
		return trace.InitTraceSilent()
	}
	return trace.InitTrace(output, 50) // default 50MB max size
}

// InitDebugger initializes the global debugger.
// This should be called once at application startup.
func InitDebugger() {
	debug.InitDebugger()
}

// InitGlobalServices initializes global services like trace and debugger.
// This should be called once at application startup.
// Deprecated: Use InitTrace(false) or InitTraceWithOutput(false, output) and InitDebugger separately for better control.
func InitGlobalServices(profile bool) error {
	if err := InitTrace(profile); err != nil {
		return err
	}
	InitDebugger()
	return nil
}

// NewEvaluatorWithNatives creates a new evaluator and registers all native functions.
// The evaluator is configured with the provided I/O streams.
// If quiet is true, stdout is set to io.Discard.
func NewEvaluatorWithNatives(stdout, stderr io.Writer, stdin io.Reader, quiet bool) *eval.Evaluator {
	evaluator := eval.NewEvaluator()

	if quiet {
		evaluator.SetOutputWriter(io.Discard)
	} else if stdout != nil {
		evaluator.SetOutputWriter(stdout)
	}
	if stderr != nil {
		evaluator.SetErrorWriter(stderr)
	}
	if stdin != nil {
		evaluator.SetInputReader(stdin)
	}

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

// InjectSystemArgs creates and injects the system object with command-line arguments
// into the evaluator's root frame.
func InjectSystemArgs(evaluator core.Evaluator, args []string) {
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
