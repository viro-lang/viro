// Package repl implements the Read-Eval-Print Loop for Viro.
//
// The REPL provides an interactive interface for evaluating Viro expressions.
// It uses the github.com/chzyer/readline library for command history, line
// editing, and multi-line input support.
//
// Features:
//   - Command history: Persistent across sessions (~/.viro_history)
//   - Multi-line input: Automatic detection of incomplete expressions
//   - Error recovery: Displays error and continues accepting input
//   - Interrupts: Ctrl+C cancels evaluation without exiting
//   - Exit commands: 'quit', 'exit', or Ctrl+D
//
// The REPL loop:
//  1. Read: Get input line (with history/editing)
//  2. Parse: Convert text to values
//  3. Eval: Execute via evaluator
//  4. Print: Display result (suppress 'none')
//  5. Loop: Repeat until exit
package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

const (
	primaryPrompt      = ">> "
	debugPrompt        = "[debug] >> "
	continuationPrompt = "... "
	historyEnvVar      = "VIRO_HISTORY_FILE"
	historyFileName    = ".viro_history"
)

// Options configures REPL behavior and can be set via CLI flags.
type Options struct {
	Prompt      string
	NoWelcome   bool
	NoHistory   bool
	HistoryFile string
	TraceOn     bool
	Args        []string
}

// REPL implements a Read-Eval-Print-Loop for Viro.
//
// Architecture per contracts:
// - Read: Use readline for input with history
// - Eval: Parse input, then evaluate with evaluator
// - Print: Display results (suppress none per FR-044)
// - Loop: Repeat until exit command
type REPL struct {
	evaluator      core.Evaluator
	rl             *readline.Instance
	out            io.Writer
	history        []string
	historyCursor  int
	pendingLines   []string
	awaitingCont   bool
	shouldContinue bool
	historyPath    string
	customPrompt   string
	noWelcome      bool
	noHistory      bool
	debugSession   *DebugSession // Interactive debugging state
}

// DebugSession manages interactive debugging state
type DebugSession struct {
	active      bool
	evaluator   core.Evaluator
	debugger    *debug.Debugger
	lastCommand string
	pausedExpr  []core.Value // The expression that was paused
}

// NewREPL creates a new REPL instance with default options.
func NewREPL(args []string) (*REPL, error) {
	return NewREPLWithOptions(&Options{
		Args: args,
	})
}

// NewREPLWithOptions creates a new REPL instance with custom options.
func NewREPLWithOptions(opts *Options) (*REPL, error) {
	if opts == nil {
		opts = &Options{}
	}

	// Initialize trace/debug sessions (Feature 002, T154)
	// Trace is initialized with default settings (stderr, 50MB max size)
	// These will be controlled via trace --on/--off and debug --on/--off
	if err := trace.InitTrace("", 50); err != nil {
		return nil, fmt.Errorf("failed to initialize trace session: %w", err)
	}
	debug.InitDebugger()

	// Enable trace if requested
	if opts.TraceOn && trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Enable(trace.TraceFilters{})
	}

	// Determine history path
	historyPath := opts.HistoryFile
	if historyPath == "" && !opts.NoHistory {
		historyPath = resolveHistoryPath(true)
	}

	// Determine prompt
	prompt := opts.Prompt
	if prompt == "" {
		prompt = primaryPrompt
	}

	// Create readline instance with prompt
	rlConfig := &readline.Config{
		Prompt:                 prompt,
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
	}

	// Only set history file if history is enabled
	if !opts.NoHistory && historyPath != "" {
		rlConfig.HistoryFile = historyPath
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return nil, err
	}

	evaluator := eval.NewEvaluator()
	evaluator.SetOutputWriter(os.Stdout)
	evaluator.SetErrorWriter(os.Stderr)
	evaluator.SetInputReader(os.Stdin)

	rootFrame := evaluator.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, evaluator)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	initializeSystemObject(evaluator, opts.Args)

	repl := &REPL{
		evaluator:      evaluator,
		rl:             rl,
		out:            os.Stdout,
		history:        []string{},
		historyCursor:  0,
		pendingLines:   nil,
		awaitingCont:   false,
		shouldContinue: true,
		historyPath:    historyPath,
		customPrompt:   prompt,
		noWelcome:      opts.NoWelcome,
		noHistory:      opts.NoHistory,
	}

	// Load persistent history only if not disabled
	if !opts.NoHistory {
		repl.loadPersistentHistory()
	}

	return repl, nil
}

// NewREPLForTest creates a REPL with injected evaluator and writer for testing purposes.
func NewREPLForTest(e core.Evaluator, out io.Writer) *REPL {
	// Initialize trace/debug sessions for tests (same as NewREPL)
	// Use os.DevNull to avoid trace output pollution during tests
	if err := trace.InitTrace(os.DevNull, 50); err != nil {
		// Log error but continue - tests should not fail due to trace init
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize trace session: %v\n", err)
	}
	debug.InitDebugger()

	if e == nil {
		e = eval.NewEvaluator()
	}
	if out == nil {
		out = io.Discard
	}

	// Configure evaluator I/O
	e.SetOutputWriter(out)
	e.SetErrorWriter(out)                   // For tests, use same writer for both
	e.SetInputReader(strings.NewReader("")) // Empty input for tests

	rootFrame := e.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, e)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	initializeSystemObject(e, []string{})

	historyPath := resolveHistoryPath(false)
	repl := &REPL{
		evaluator:      e,
		rl:             nil,
		out:            out,
		history:        []string{},
		historyCursor:  0,
		pendingLines:   nil,
		awaitingCont:   false,
		shouldContinue: true,
		historyPath:    historyPath,
	}
	repl.loadPersistentHistory()
	return repl
}

// WelcomeMessage returns the default multi-line welcome text shown when the REPL starts.
func WelcomeMessage() string {
	return "Viro 0.1.0\nType 'exit' or 'quit' to leave\n\n"
}

// Run starts the REPL loop.
func (r *REPL) Run() error {
	if r.rl == nil {
		return fmt.Errorf("readline instance not configured")
	}
	defer r.rl.Close()

	// Print welcome message
	r.printWelcome()
	r.setPrompt(r.getCurrentPrompt())

	// Main loop
	for {
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				r.handleInterrupt(true)
				continue
			}
			if err == io.EOF {
				fmt.Fprintln(r.out, "")
				r.handleExit(true)
				return nil
			}
			return err
		}

		// Check if we're in debug mode
		if r.IsInDebugMode() {
			r.processDebugCommand(line)
		} else {
			r.processLine(line, true)
		}

		if !r.shouldContinue {
			return nil
		}
	}
}

// EvalLineForTest evaluates a single line and prints to the configured writer.
func (r *REPL) EvalLineForTest(input string) {
	if r == nil {
		return
	}
	r.processLine(strings.TrimRight(input, "\r\n"), false)
}

// AwaitingContinuation reports whether the REPL is waiting for additional lines
// to complete the current command (multi-line input state).
func (r *REPL) AwaitingContinuation() bool {
	if r == nil {
		return false
	}
	return r.awaitingCont
}

func (r *REPL) processLine(input string, interactive bool) {
	if r == nil || !r.shouldContinue {
		return
	}

	clean := strings.TrimRight(input, "\r\n")
	trimmed := strings.TrimSpace(clean)

	if !r.awaitingCont && isExitCommand(trimmed) {
		r.pendingLines = nil
		r.awaitingCont = false
		r.recordHistory(trimmed)
		r.handleExit(interactive)
		return
	}

	// Special REPL-only shortcut: bare '?' shows categories
	// In scripts, '?' requires an argument per its Arity: 1
	if !r.awaitingCont && trimmed == "?" {
		r.recordHistory(trimmed)
		r.handleHelpShortcut()
		return
	}

	if trimmed == "" && !r.awaitingCont {
		return
	}

	if trimmed != "" || r.awaitingCont {
		r.pendingLines = append(r.pendingLines, clean)
	}

	joined := strings.Join(r.pendingLines, "\n")
	values, err := parse.Parse(joined)
	if err != nil {
		if shouldAwaitContinuation(err.(*verror.Error)) {
			r.awaitingCont = true
			if interactive {
				r.setPrompt(continuationPrompt)
			}
			return
		}

		r.awaitingCont = false
		if interactive {
			r.setPrompt(r.getCurrentPrompt())
		}
		r.pendingLines = nil
		r.recordHistory(joined)
		r.printError(err)
		return
	}

	r.awaitingCont = false
	if interactive {
		r.setPrompt(r.getCurrentPrompt())
	}
	r.pendingLines = nil
	r.recordHistory(joined)
	r.evalParsedValues(values)
}

// processDebugCommand handles debug commands when in debug mode
func (r *REPL) processDebugCommand(line string) {
	if r == nil || !r.IsInDebugMode() {
		return
	}

	clean := strings.TrimRight(line, "\r\n")
	trimmed := strings.TrimSpace(clean)

	// Handle debug commands
	continueDebug, err := r.HandleDebugCommand(trimmed)
	if err != nil {
		r.printError(err)
		return
	}

	// If we should continue in debug mode, stay in debug mode
	// If not, we've exited debug mode, so resume normal execution
	if !continueDebug {
		// We've exited debug mode, resume normal execution
		// The debugger should have been resumed, so the next evaluation should continue
		return
	}

	// Still in debug mode, continue reading debug commands
}

// printWelcome displays the welcome message unless disabled.
func (r *REPL) printWelcome() {
	if !r.noWelcome {
		fmt.Fprint(r.out, WelcomeMessage())
	}
}

func (r *REPL) printError(err error) {
	if err == nil {
		return
	}
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintln(r.out, verror.FormatErrorWithContext(vErr))
	} else {
		fmt.Fprintln(r.out, err.Error())
	}
}

// HistoryEntries returns a copy of the recorded command history.
func (r *REPL) HistoryEntries() []string {
	if r == nil {
		return nil
	}
	entries := make([]string, len(r.history))
	copy(entries, r.history)
	return entries
}

// HistoryUp moves the history cursor upward (towards older commands) and returns the entry.
func (r *REPL) HistoryUp() (string, bool) {
	if r == nil || len(r.history) == 0 {
		return "", false
	}
	if r.historyCursor > 0 {
		r.historyCursor--
	} else if r.historyCursor == 0 {
		// stay at first entry
	} else {
		// cursor beyond end, step to last entry
		r.historyCursor = len(r.history) - 1
	}
	return r.history[r.historyCursor], true
}

// HistoryDown moves the history cursor downward (towards newer commands).
// When reaching the end, it returns an empty string and false to indicate fresh input.
func (r *REPL) HistoryDown() (string, bool) {
	if r == nil || len(r.history) == 0 {
		return "", false
	}
	last := len(r.history) - 1
	switch {
	case r.historyCursor < last:
		r.historyCursor++
		return r.history[r.historyCursor], true
	case r.historyCursor == last:
		r.historyCursor = len(r.history)
		return "", false
	case r.historyCursor > len(r.history):
		r.historyCursor = len(r.history)
		fallthrough
	default:
		return "", false
	}
}

func (r *REPL) recordHistory(entry string) {
	if r == nil || r.noHistory {
		return
	}
	trimmed := strings.TrimSpace(entry)
	if trimmed == "" {
		r.historyCursor = len(r.history)
		return
	}
	r.history = append(r.history, trimmed)
	r.historyCursor = len(r.history)
	r.persistHistoryLine(trimmed)
}

func (r *REPL) setPrompt(prompt string) {
	if r == nil || r.rl == nil {
		return
	}
	r.rl.SetPrompt(prompt)
}

// getCurrentPrompt returns the appropriate prompt based on debugger state (T154)
func (r *REPL) getCurrentPrompt() string {
	if r == nil {
		return primaryPrompt
	}

	// Check if debugger is in active mode (breakpoints or stepping)
	if debug.GlobalDebugger != nil && debug.GlobalDebugger.Mode() != debug.DebugModeOff {
		return debugPrompt
	}

	// Use custom prompt if set, otherwise use default
	if r.customPrompt != "" {
		return r.customPrompt
	}

	return primaryPrompt
}

// evalParsedValues evaluates parsed values and handles debug pauses
func (r *REPL) evalParsedValues(values []core.Value) {
	result, err := r.evaluator.DoBlock(values)
	if err != nil {
		// Check if this is a debug pause error
		if vErr, ok := err.(*verror.Error); ok && vErr.ID == verror.ErrIDDebugPause {
			// Enter debug mode and store the paused expression
			if !r.IsInDebugMode() {
				r.EnterDebugMode()
			}
			if r.debugSession != nil {
				r.debugSession.pausedExpr = values
			}
			// Don't print the error, just return to let the REPL handle debug mode
			return
		}
		r.printError(err)
		return
	}

	if result.GetType() != value.TypeNone {
		formResult, err := native.Form([]core.Value{result}, nil, r.evaluator)
		if err != nil {
			r.printError(err)
			return
		}
		fmt.Fprintln(r.out, formResult.Form())
	}
}

func (r *REPL) handleExit(interactive bool) {
	if r == nil {
		return
	}
	r.pendingLines = nil
	r.awaitingCont = false
	r.shouldContinue = false
	if interactive {
		r.setPrompt(r.getCurrentPrompt())
	}
	fmt.Fprintln(r.out, "Goodbye!")
}

func (r *REPL) handleInterrupt(interactive bool) {
	if r == nil {
		return
	}
	r.pendingLines = nil
	r.awaitingCont = false
	if interactive {
		r.setPrompt(r.getCurrentPrompt())
	}
	r.shouldContinue = true
	fmt.Fprintln(r.out, "^C")
}

// ShouldContinue reports whether the REPL should keep accepting input.
func (r *REPL) ShouldContinue() bool {
	if r == nil {
		return false
	}
	return r.shouldContinue
}

// ResetForTest resets the REPL continuation state for testing.
func (r *REPL) ResetForTest() {
	if r == nil {
		return
	}
	r.shouldContinue = true
	r.awaitingCont = false
	r.pendingLines = nil
	r.historyCursor = len(r.history)
}

// SimulateInterruptForTest emulates a Ctrl+C interrupt for tests.
func (r *REPL) SimulateInterruptForTest() {
	r.handleInterrupt(false)
}

func (r *REPL) loadPersistentHistory() {
	if r == nil {
		return
	}
	if r.historyPath == "" {
		r.historyCursor = len(r.history)
		return
	}
	entries, err := readHistoryFile(r.historyPath)
	if err != nil {
		return
	}
	r.history = append([]string{}, entries...)
	r.historyCursor = len(r.history)
}

func (r *REPL) persistHistoryLine(entry string) {
	if r == nil {
		return
	}
	if r.rl != nil {
		_ = r.rl.SaveHistory(entry)
		return
	}
	if r.historyPath == "" {
		return
	}
	if err := ensureHistoryDirectory(r.historyPath); err != nil {
		return
	}
	file, err := os.OpenFile(r.historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(entry + "\n")
}

func resolveHistoryPath(allowDefault bool) string {
	if override := strings.TrimSpace(os.Getenv(historyEnvVar)); override != "" {
		return filepath.Clean(override)
	}
	if !allowDefault {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, historyFileName)
}

func readHistoryFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	entries := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func ensureHistoryDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func shouldAwaitContinuation(err *verror.Error) bool {
	if err == nil {
		return false
	}

	switch err.ID {
	case verror.ErrIDUnexpectedEOF, verror.ErrIDUnclosedBlock, verror.ErrIDUnclosedParen:
		return true
	case verror.ErrIDInvalidSyntax:
		arg := strings.ToLower(err.Args[0])
		return strings.Contains(arg, "unclosed string literal")
	default:
		return false
	}
}

func isExitCommand(input string) bool {
	if input == "" {
		return false
	}
	return strings.EqualFold(input, "quit") || strings.EqualFold(input, "exit")
}

func initializeSystemObject(evaluator core.Evaluator, args []string) {
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

// handleHelpShortcut handles the special REPL-only '?' command (no arguments).
// This calls the native Help function with an empty argument list to display categories.
func (r *REPL) handleHelpShortcut() {
	if r == nil {
		return
	}

	// Temporarily redirect os.Stdout to capture Help output
	oldStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		// Fallback: call Help normally if pipe creation fails
		_, _ = native.Help([]core.Value{}, map[string]core.Value{}, r.evaluator)
		return
	}

	os.Stdout = wOut

	// Call Help function directly with no arguments to show categories
	result, helpErr := native.Help([]core.Value{}, map[string]core.Value{}, r.evaluator)

	// Restore stdout immediately
	wOut.Close()
	os.Stdout = oldStdout

	// Copy captured output to REPL's writer
	output := make([]byte, 8192)
	n, _ := rOut.Read(output)
	if n > 0 {
		r.out.Write(output[:n])
	}
	rOut.Close()

	if helpErr != nil {
		r.printError(helpErr)
		return
	}

	// Help returns none, no need to print result
	if result.GetType() != value.TypeNone {
		fmt.Fprintln(r.out, result.Form())
	}
}

// EnterDebugMode activates interactive debugging
func (r *REPL) EnterDebugMode() error {
	if r == nil {
		return fmt.Errorf("REPL is nil")
	}

	r.debugSession = &DebugSession{
		active:    true,
		evaluator: r.evaluator,
		debugger:  debug.GlobalDebugger,
	}

	// Change prompt to debug mode
	r.setPrompt(debugPrompt)
	return nil
}

// ExitDebugMode deactivates interactive debugging
func (r *REPL) ExitDebugMode() {
	if r == nil || r.debugSession == nil {
		return
	}

	r.debugSession.active = false
	r.debugSession = nil

	// Restore normal prompt
	r.setPrompt(r.getCurrentPrompt())
}

// IsInDebugMode returns true if interactive debugging is active
func (r *REPL) IsInDebugMode() bool {
	return r != nil && r.debugSession != nil && r.debugSession.active
}

// HandleDebugCommand processes debug commands during interactive debugging
func (r *REPL) HandleDebugCommand(cmd string) (continueDebug bool, err error) {
	if r == nil || r.debugSession == nil {
		return false, fmt.Errorf("not in debug mode")
	}

	cmd = strings.TrimSpace(strings.ToLower(cmd))
	r.debugSession.lastCommand = cmd

	switch cmd {
	case "n", "next", "step":
		// Step to next expression
		r.debugSession.debugger.EnableStepping()
		// Resume execution and re-evaluate the paused expression
		return r.resumeAndReevaluate()

	case "c", "continue":
		// Continue execution until next breakpoint
		r.debugSession.debugger.DisableStepping()
		// Resume execution and re-evaluate the paused expression
		return r.resumeAndReevaluate()

	case "l", "locals":
		// Show local variables
		return r.showDebugLocals()

	case "s", "stack":
		// Show call stack
		return r.showDebugStack()

	case "p", "print":
		// Print expression (requires argument)
		return false, fmt.Errorf("print command requires an expression argument")

	case "q", "quit":
		// Exit debug mode
		r.ExitDebugMode()
		return false, nil

	default:
		// Check if it's a print command with argument
		if strings.HasPrefix(cmd, "p ") || strings.HasPrefix(cmd, "print ") {
			return r.handleDebugPrint(cmd)
		}

		return false, fmt.Errorf("unknown debug command: %s", cmd)
	}
}

// showDebugLocals displays local variables in debug mode
func (r *REPL) showDebugLocals() (bool, error) {
	if r.debugSession == nil {
		return false, fmt.Errorf("not in debug mode")
	}

	// Get current frame locals
	frameIdx := r.evaluator.CurrentFrameIndex()
	locals := r.debugSession.debugger.GetFrameLocals(r.evaluator, frameIdx)

	if len(locals) == 0 {
		fmt.Fprintln(r.out, "(no local variables)")
	} else {
		fmt.Fprintln(r.out, "Local variables:")
		for name, val := range locals {
			fmt.Fprintf(r.out, "  %s: %s\n", name, val.Form())
		}
	}

	return true, nil
}

// showDebugStack displays call stack in debug mode
func (r *REPL) showDebugStack() (bool, error) {
	if r.debugSession == nil {
		return false, fmt.Errorf("not in debug mode")
	}

	// Get call stack
	stack := r.debugSession.debugger.GetCallStack(r.evaluator)

	if len(stack) == 0 {
		fmt.Fprintln(r.out, "(empty call stack)")
	} else {
		fmt.Fprintln(r.out, "Call stack:")
		for i, frame := range stack {
			fmt.Fprintf(r.out, "  [%d] %s\n", i, frame)
		}
	}

	return true, nil
}

// handleDebugPrint evaluates and prints an expression in debug context
func (r *REPL) handleDebugPrint(cmd string) (bool, error) {
	if r.debugSession == nil {
		return false, fmt.Errorf("not in debug mode")
	}

	// Extract expression from command
	var expr string
	if strings.HasPrefix(cmd, "p ") {
		expr = strings.TrimPrefix(cmd, "p ")
	} else if strings.HasPrefix(cmd, "print ") {
		expr = strings.TrimPrefix(cmd, "print ")
	} else {
		return false, fmt.Errorf("invalid print command format")
	}

	if expr == "" {
		return false, fmt.Errorf("print command requires an expression")
	}

	// Parse and evaluate the expression
	values, err := parse.Parse(expr)
	if err != nil {
		return false, fmt.Errorf("parse error: %v", err)
	}

	result, err := r.evaluator.DoBlock(values)
	if err != nil {
		return false, fmt.Errorf("evaluation error: %v", err)
	}

	fmt.Fprintf(r.out, "%s\n", result.Form())
	return true, nil
}

// resumeAndReevaluate resumes the debugger and re-evaluates the paused expression
func (r *REPL) resumeAndReevaluate() (bool, error) {
	if r.debugSession == nil || r.debugSession.pausedExpr == nil {
		return false, fmt.Errorf("no paused expression to resume")
	}

	// Resume the debugger
	r.debugSession.debugger.ResumeExecution()

	// Exit debug mode since we're resuming execution
	r.ExitDebugMode()

	// Re-evaluate the paused expression
	r.evalParsedValues(r.debugSession.pausedExpr)

	// Clear the paused expression
	r.debugSession.pausedExpr = nil

	// Return false to exit debug mode (we've resumed)
	return false, nil
}
