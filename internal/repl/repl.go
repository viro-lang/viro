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

// REPL implements a Read-Eval-Print-Loop for Viro.
//
// Architecture per contracts:
// - Read: Use readline for input with history
// - Eval: Parse input, then evaluate with evaluator
// - Print: Display results (suppress none per FR-044)
// - Loop: Repeat until exit command
type REPL struct {
	evaluator *eval.Evaluator
	rl        *readline.Instance
	out       io.Writer
}

// NewREPL creates a new REPL instance.
func NewREPL() (*REPL, error) {
	// Create readline instance with prompt
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          ">> ",
		HistoryFile:     "", // TODO: Add history file in home directory
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}

	return &REPL{
		evaluator: eval.NewEvaluator(),
		rl:        rl,
		out:       os.Stdout,
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
		evaluator: e,
		rl:        nil,
		out:       out,
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

	// Main loop
	for {
		// Read
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Fprintln(r.out, "")
				fmt.Fprintln(r.out, "Goodbye!")
				return nil
			}
			return err
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for exit commands
		if line == "exit" || line == "quit" {
			fmt.Fprintln(r.out, "Goodbye!")
			return nil
		}

		// Eval and Print
		r.evalAndPrint(line)
	}
}

// EvalLineForTest evaluates a single line and prints to the configured writer.
func (r *REPL) EvalLineForTest(input string) {
	if r == nil {
		return
	}
	r.evalAndPrint(strings.TrimSpace(input))
}

// evalAndPrint parses, evaluates, and displays the result of an input line.
func (r *REPL) evalAndPrint(input string) {
	// Parse
	values, err := parse.Parse(input)
	if err != nil {
		r.printError(err)
		return
	}

	// Evaluate
	result, err := r.evaluator.Do_Blk(values)
	if err != nil {
		r.printError(err)
		return
	}

	// Print (suppress none per FR-044)
	if result.Type != value.TypeNone {
		fmt.Fprintln(r.out, r.formatValue(result))
	}
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
