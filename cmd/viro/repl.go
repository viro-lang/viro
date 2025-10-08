package main

import (
	"fmt"
	"io"
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
	}, nil
}

// Run starts the REPL loop.
func (r *REPL) Run() error {
	defer r.rl.Close()

	// Print welcome message
	r.printWelcome()

	// Main loop
	for {
		// Read
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nGoodbye!")
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
			fmt.Println("Goodbye!")
			return nil
		}

		// Eval and Print
		r.evalAndPrint(line)
	}
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
		fmt.Println(r.formatValue(result))
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
			return fmt.Sprintf(`"%s"`, s.String())
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
	fmt.Println("Viro 0.1.0")
	fmt.Println("Type 'exit' or 'quit' to leave")
	fmt.Println()
}

func (r *REPL) printError(err *verror.Error) {
	if err == nil {
		return
	}
	fmt.Println(verror.FormatErrorWithContext(err))
}
