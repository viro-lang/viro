// Package eval provides the core evaluation engine for Viro.
//
// Architecture:
// - Evaluator struct holds the stack and execution state
// - Do_Next evaluates a single value
// - Do_Blk evaluates a block of values in sequence
// - Type-based dispatch routes evaluation by value type
//
// Per Constitution Principle III: Explicit type dispatch, no polymorphism.
package eval

import (
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluator is the core evaluation engine.
//
// Design per Constitution Principle IV: Index-based access to stack/frames.
// Stack holds both data and frames in a unified structure.
type Evaluator struct {
	Stack  *stack.Stack
	Frames []*frame.Frame
}

// NewEvaluator creates a new evaluation engine with an empty stack.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		Stack:  &stack.Stack{Data: make([]value.Value, 0, 1024)},
		Frames: make([]*frame.Frame, 0, 64),
	}
}

// Do_Next evaluates a single value and returns the result.
//
// Contract per data-model.md §3:
// - Integers, strings, logic, none: Return self (literals)
// - Blocks [...]: Return self (deferred evaluation)
// - Parens (...): Evaluate contents and return result
// - Words: Look up in frame and evaluate result
// - Set-words (word:): Bind next value
// - Get-words (:word): Fetch value without evaluation
// - Lit-words ('word): Return word as-is
//
// Returns the evaluated value and any error encountered.
func (e *Evaluator) Do_Next(val value.Value) (value.Value, *verror.Error) {
	// Type-based dispatch per Constitution Principle III
	switch val.Type {
	case value.TypeInteger, value.TypeString, value.TypeLogic, value.TypeNone:
		// Literals evaluate to themselves
		return val, nil

	case value.TypeBlock:
		// Blocks return self (deferred evaluation)
		return val, nil

	case value.TypeParen:
		// Parens evaluate contents
		return e.evalParen(val)

	case value.TypeWord:
		// Words look up in frame
		return e.evalWord(val)

	case value.TypeSetWord:
		// Set-words are handled in Do_Blk (need next value from sequence)
		// If we reach here, it's a set-word in isolation (error)
		wordStr, _ := val.AsWord()
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})

	case value.TypeGetWord:
		// Get-words fetch without evaluation
		return e.evalGetWord(val)

	case value.TypeLitWord:
		// Lit-words return as word
		// Extract the word symbol and return as a word
		return value.WordVal(val.Payload.(string)), nil

	case value.TypeFunction:
		// Functions return self (they're values)
		return val, nil

	default:
		return value.NoneVal(), verror.NewInternalError("unknown value type in Do_Next", [3]string{})
	}
}

// Do_Blk evaluates a block of values in sequence.
//
// Contract per data-model.md §3: Evaluates each value left-to-right.
// Returns the last result, or none if block is empty.
//
// Special handling for set-words: evaluates next value and binds result.
func (e *Evaluator) Do_Blk(vals []value.Value) (value.Value, *verror.Error) {
	if len(vals) == 0 {
		return value.NoneVal(), nil
	}

	var result value.Value
	var err *verror.Error

	// Use index-based loop to handle set-words that consume next value
	for i := 0; i < len(vals); i++ {
		val := vals[i]

		// Special case: set-word consumes next value
		if val.Type == value.TypeSetWord {
			result, err = e.evalSetWord(val, vals, &i)
			if err != nil {
				return value.NoneVal(), err
			}
			continue
		}

		result, err = e.Do_Next(val)
		if err != nil {
			return value.NoneVal(), err
		}
	}

	return result, nil
}

// evalParen evaluates the contents of a paren and returns the result.
func (e *Evaluator) evalParen(val value.Value) (value.Value, *verror.Error) {
	block, ok := val.AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("paren value does not contain BlockValue", [3]string{})
	}

	// Evaluate contents in sequence
	return e.Do_Blk(block.Elements)
}

// evalWord looks up a word in the current frame and evaluates the result.
func (e *Evaluator) evalWord(val value.Value) (value.Value, *verror.Error) {
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("word value does not contain string", [3]string{})
	}

	// Get current frame
	if len(e.Frames) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	currentFrame := e.Frames[len(e.Frames)-1]

	// Look up word in frame
	result, ok := currentFrame.Get(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Evaluate the result (words evaluate to their values, then re-evaluate)
	return e.Do_Next(result)
}

// evalSetWord handles set-word evaluation: binds next value to word.
//
// Contract per data-model.md §4:
// - Set-Word: Evaluate next expression, store in Binding frame at Symbol
//
// Parameters:
// - val: The set-word value (contains word symbol)
// - vals: The full sequence of values being evaluated
// - i: Pointer to current index (will be advanced to skip consumed value)
func (e *Evaluator) evalSetWord(val value.Value, vals []value.Value, i *int) (value.Value, *verror.Error) {
	// Extract word symbol from set-word
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("set-word value does not contain string", [3]string{})
	}

	// Check if there's a next value to evaluate
	if *i+1 >= len(vals) {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{"set-word without value: " + wordStr, "", ""})
	}

	// Get current frame (create one if needed)
	if len(e.Frames) == 0 {
		// Create a global frame if none exists
		e.Frames = append(e.Frames, frame.NewFrame(frame.FrameFunctionArgs, -1))
	}

	currentFrame := e.Frames[len(e.Frames)-1]

	// Advance to next value and evaluate it
	*i++
	nextVal := vals[*i]
	result, err := e.Do_Next(nextVal)
	if err != nil {
		return value.NoneVal(), err
	}

	// Bind the result to the word in current frame
	currentFrame.Bind(wordStr, result)

	// Return the bound value
	return result, nil
}

// evalGetWord looks up a word in the current frame WITHOUT evaluating the result.
//
// Contract per data-model.md §4:
// - Get-Word: Look up Symbol in Binding frame, return without evaluation
//
// This is the difference from regular word evaluation which re-evaluates the result.
func (e *Evaluator) evalGetWord(val value.Value) (value.Value, *verror.Error) {
	wordStr, ok := val.AsWord()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("get-word value does not contain string", [3]string{})
	}

	// Get current frame
	if len(e.Frames) == 0 {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	currentFrame := e.Frames[len(e.Frames)-1]

	// Look up word in frame
	result, ok := currentFrame.Get(wordStr)
	if !ok {
		return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
	}

	// Return without re-evaluation (key difference from evalWord)
	return result, nil
}
