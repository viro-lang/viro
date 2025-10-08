package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

const (
	primaryPrompt      = ">> "
	continuationPrompt = "... "
)

// REPL implements a Read-Eval-Print-Loop for Viro.
//
// Architecture per contracts:
// - Read: Use readline for input with history
// - Eval: Parse input, then evaluate with evaluator
// - Print: Display results (suppress none per FR-044)
// - Loop: Repeat until exit command
type REPL struct {
	evaluator     *eval.Evaluator
	rl            *readline.Instance
	out           io.Writer
	history       []string
	historyCursor int
	pendingLines  []string
	awaitingCont  bool
}

// NewREPL creates a new REPL instance.
func NewREPL() (*REPL, error) {
	// Create readline instance with prompt
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          primaryPrompt,
		HistoryFile:     "", // TODO: Add history file in home directory
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}

	return &REPL{
		evaluator:     eval.NewEvaluator(),
		rl:            rl,
		out:           os.Stdout,
		history:       []string{},
		historyCursor: 0,
		pendingLines:  nil,
		awaitingCont:  false,
	}, nil
}

// NewREPLForTest creates a REPL with injected evaluator and writer for testing purposes.
func NewREPLForTest(e *eval.Evaluator, out io.Writer) *REPL {
	if e == nil {
		e = eval.NewEvaluator()
	}
	if out == nil {
		out = io.Discard
	}
	return &REPL{
		evaluator:     e,
		rl:            nil,
		out:           out,
		history:       []string{},
		historyCursor: 0,
		pendingLines:  nil,
		awaitingCont:  false,
	}
}

// Run starts the REPL loop.
func (r *REPL) Run() error {
	if r.rl == nil {
		return fmt.Errorf("readline instance not configured")
	}
	defer r.rl.Close()

	// Print welcome message
	r.printWelcome()
	r.setPrompt(primaryPrompt)

	// Main loop
	for {
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Fprintln(r.out, "")
				fmt.Fprintln(r.out, "Goodbye!")
				return nil
			}
			return err
		}

		clean := strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimSpace(clean)
		if trimmed == "" && !r.awaitingCont {
			continue
		}

		if !r.awaitingCont && (trimmed == "exit" || trimmed == "quit") {
			fmt.Fprintln(r.out, "Goodbye!")
			return nil
		}

		r.processLine(clean, true)
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
	clean := strings.TrimRight(input, "\r\n")
	trimmed := strings.TrimSpace(clean)

	if trimmed == "" && !r.awaitingCont {
		return
	}

	if trimmed != "" || r.awaitingCont {
		r.pendingLines = append(r.pendingLines, clean)
	}

	joined := strings.Join(r.pendingLines, "\n")
	values, err := parse.Parse(joined)
	if err != nil {
		if shouldAwaitContinuation(err) {
			r.awaitingCont = true
			if interactive {
				r.setPrompt(continuationPrompt)
			}
			return
		}

		r.awaitingCont = false
		if interactive {
			r.setPrompt(primaryPrompt)
		}
		r.pendingLines = nil
		r.recordHistory(joined)
		r.printError(err)
		return
	}

	r.awaitingCont = false
	if interactive {
		r.setPrompt(primaryPrompt)
	}
	r.pendingLines = nil
	r.recordHistory(joined)
	r.evalParsedValues(values)
}

// evalAndPrint parses, evaluates, and displays the result of an input line.
func (r *REPL) evalAndPrint(input string) {
	// Parse
	values, err := parse.Parse(input)
	if err != nil {
		r.printError(err)
		return
	}

	r.evalParsedValues(values)
}

// formatValue formats a value for display in the REPL.
func (r *REPL) formatValue(v value.Value) string {
	switch v.Type {
	case value.TypeInteger:
		if i, ok := v.AsInteger(); ok {
			return fmt.Sprintf("%d", i)
		}
	case value.TypeString:
		if s, ok := v.AsString(); ok {
			return fmt.Sprintf("\"%s\"", s.String())
		}
	case value.TypeLogic:
		if b, ok := v.AsLogic(); ok {
			if b {
				return "true"
			}
			return "false"
		}
	case value.TypeNone:
		return ""
	case value.TypeWord:
		if w, ok := v.AsWord(); ok {
			return w
		}
	case value.TypeSetWord:
		if w, ok := v.AsWord(); ok {
			return w + ":"
		}
	case value.TypeGetWord:
		if w, ok := v.AsWord(); ok {
			return ":" + w
		}
	case value.TypeLitWord:
		if w, ok := v.AsWord(); ok {
			return "'" + w
		}
	case value.TypeBlock:
		if b, ok := v.AsBlock(); ok {
			var parts []string
			for _, elem := range b.Elements {
				parts = append(parts, r.formatValue(elem))
			}
			return "[" + strings.Join(parts, " ") + "]"
		}
	case value.TypeParen:
		if p, ok := v.AsBlock(); ok {
			var parts []string
			for _, elem := range p.Elements {
				parts = append(parts, r.formatValue(elem))
			}
			return "(" + strings.Join(parts, " ") + ")"
		}
	}
	return fmt.Sprintf("%v", v)
}

// printWelcome displays the welcome message.
func (r *REPL) printWelcome() {
	fmt.Fprintln(r.out, "Viro 0.1.0")
	fmt.Fprintln(r.out, "Type 'exit' or 'quit' to leave")
	fmt.Fprintln(r.out, "")
}

func (r *REPL) printError(err *verror.Error) {
	if err == nil {
		return
	}
	fmt.Fprintln(r.out, verror.FormatErrorWithContext(err))
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
	if r == nil {
		return
	}
	trimmed := strings.TrimSpace(entry)
	if trimmed == "" {
		r.historyCursor = len(r.history)
		return
	}
	r.history = append(r.history, trimmed)
	r.historyCursor = len(r.history)
	if r.rl != nil {
		if err := r.rl.SaveHistory(trimmed); err != nil {
			// Saving history is best-effort; ignore errors for now.
		}
	}
}

func (r *REPL) setPrompt(prompt string) {
	if r == nil || r.rl == nil {
		return
	}
	r.rl.SetPrompt(prompt)
}

func (r *REPL) evalParsedValues(values []value.Value) {
	result, err := r.evaluator.Do_Blk(values)
	if err != nil {
		r.printError(err)
		return
	}

	if result.Type != value.TypeNone {
		fmt.Fprintln(r.out, r.formatValue(result))
	}
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
