package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Set implements the `set` native.
//
// Contract: set word value
// - First argument must be a word (symbol to bind)
// - Second argument is any value (already evaluated)
// - Binds word in current frame and returns the value
func Set(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"set", "2", formatInt(len(args))},
		)
	}

	if args[0].Type != value.TypeWord {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"set", "word", args[0].Type.String()},
		)
	}

	symbol, _ := args[0].AsWord()
	assignment := []value.Value{value.SetWordVal(symbol), args[1]}

	result, err := eval.Do_Blk(assignment)
	if err != nil {
		return value.NoneVal(), err
	}

	return result, nil
}

// Get implements the `get` native.
//
// Contract: get word
// - Argument must be a word symbol
// - Returns bound value from current frame chain
func Get(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"get", "1", formatInt(len(args))},
		)
	}

	if args[0].Type != value.TypeWord {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"get", "word", args[0].Type.String()},
		)
	}

	symbol, _ := args[0].AsWord()
	return eval.Do_Next(value.GetWordVal(symbol))
}

// TypeQ implements the `type?` native.
//
// Contract: type? value -> word representing type name
func TypeQ(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"type?", "1", formatInt(len(args))},
		)
	}

	typeName := typeNameFor(args[0].Type)
	return value.WordVal(typeName), nil
}

func typeNameFor(t value.ValueType) string {
	switch t {
	case value.TypeInteger:
		return "integer!"
	case value.TypeString:
		return "string!"
	case value.TypeLogic:
		return "logic!"
	case value.TypeNone:
		return "none!"
	case value.TypeBlock:
		return "block!"
	case value.TypeWord:
		return "word!"
	case value.TypeSetWord:
		return "set-word!"
	case value.TypeGetWord:
		return "get-word!"
	case value.TypeLitWord:
		return "lit-word!"
	case value.TypeFunction:
		return "function!"
	case value.TypeParen:
		return "paren!"
	default:
		return "unknown!"
	}
}
