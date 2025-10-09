package native

import (
	"github.com/marcin-radoszewski/viro/internal/frame"
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
	case value.TypeDecimal:
		return "decimal!"
	case value.TypeObject:
		return "object!"
	case value.TypePort:
		return "port!"
	case value.TypePath:
		return "path!"
	default:
		return "unknown!"
	}
}

// frameManager interface for object natives to access frame operations.
type frameManager interface {
	RegisterFrame(f *frame.Frame) int
	GetFrameByIndex(idx int) *frame.Frame
	MarkFrameCaptured(idx int)
	PushFrameContext(f *frame.Frame) int
	PopFrameContext()
	Do_Blk(vals []value.Value) (value.Value, *verror.Error)
}

// Object implements the `object` native.
//
// Contract (Feature 002, FR-009): object spec
//   - spec: block describing fields and optional initial values
//   - Syntax: [word] for field declaration (initialized to none)
//     [word: value] for explicit initialization
//   - Returns object! instance with dedicated frame
//   - Evaluates initializers in object's context
func Object(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"object", "1", formatInt(len(args))},
		)
	}

	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"object", "block", args[0].Type.String()},
		)
	}

	spec, _ := args[0].AsBlock()

	// Type-assert to access frame management
	mgr, ok := eval.(frameManager)
	if !ok {
		return value.NoneVal(), verror.NewInternalError(
			"internal-error",
			[3]string{"object", "frame-manager-unavailable", ""},
		)
	}

	// Parse spec to extract field names and initializers
	fields := []string{}
	initializers := make(map[string][]value.Value)
	seenFields := make(map[string]bool)

	for i := 0; i < len(spec.Elements); i++ {
		val := spec.Elements[i]

		switch val.Type {
		case value.TypeWord:
			// Simple field declaration: field
			word, _ := val.AsWord()
			if seenFields[word] {
				return value.NoneVal(), verror.NewScriptError(
					"object-field-duplicate",
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true
			initializers[word] = []value.Value{value.NoneVal()} // Default to none

		case value.TypeSetWord:
			// Field with initializer: field: value(s)
			word, _ := val.AsWord() // SetWord AsWord returns the symbol
			if seenFields[word] {
				return value.NoneVal(), verror.NewScriptError(
					"object-field-duplicate",
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true

			// Collect initializer values until next SET-WORD
			// (Plain words/blocks/etc are all part of the initializer expression)
			i++
			if i >= len(spec.Elements) {
				return value.NoneVal(), verror.NewScriptError(
					verror.ErrIDInvalidSyntax,
					[3]string{"object", "set-word-without-value", word},
				)
			}

			// Collect all values for this initializer until next SetWord or standalone Word
			initVals := []value.Value{}
			for i < len(spec.Elements) {
				nextVal := spec.Elements[i]
				// Stop at next SET-WORD (definite field boundary)
				if nextVal.Type == value.TypeSetWord {
					i-- // Back up so outer loop processes this
					break
				}
				initVals = append(initVals, nextVal)
				i++
			}

			if len(initVals) == 0 {
				initVals = []value.Value{value.NoneVal()}
			}
			initializers[word] = initVals

		default:
			// Ignore other value types (could be refinements in future)
			continue
		}
	} // Create object frame with parent set to current frame
	parentIdx := 0 // Global frame as parent
	if provider, ok := eval.(frameProvider); ok {
		parentIdx = provider.CurrentFrameIndex()
	}

	objFrame := frame.NewObjectFrame(parentIdx, fields, nil)

	// Register frame and get its index
	frameIdx := mgr.RegisterFrame(objFrame)

	// Mark frame as captured so it persists after PopFrameContext
	// Object frames must remain in frameStore for path traversal
	mgr.MarkFrameCaptured(frameIdx)

	// Push object frame as active context for initializer evaluation
	mgr.PushFrameContext(objFrame)
	defer mgr.PopFrameContext()

	// Evaluate initializers in object context
	for _, field := range fields {
		initVals := initializers[field]

		// Evaluate the initializer value(s)
		evaled, err := mgr.Do_Blk(initVals)
		if err != nil {
			return value.NoneVal(), err
		}

		// Bind to object frame
		objFrame.Bind(field, evaled)
	}

	// Create ObjectInstance wrapping the frame
	obj := value.NewObject(frameIdx, fields, nil)

	return value.ObjectVal(obj), nil
}

// Context implements the `context` native.
//
// Contract (Feature 002, FR-009): context spec
// - Alias for object but with isolated scope (no parent frame)
// - spec: block describing fields and optional initial values
// - Returns object! instance with isolated frame
func Context(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"context", "1", formatInt(len(args))},
		)
	}

	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"context", "block", args[0].Type.String()},
		)
	}

	spec, _ := args[0].AsBlock()

	// Type-assert to access frame management
	mgr, ok := eval.(frameManager)
	if !ok {
		return value.NoneVal(), verror.NewInternalError(
			"internal-error",
			[3]string{"context", "frame-manager-unavailable", ""},
		)
	}

	// Parse spec (same as object)
	fields := []string{}
	initializers := make(map[string][]value.Value)
	seenFields := make(map[string]bool)

	for i := 0; i < len(spec.Elements); i++ {
		val := spec.Elements[i]

		switch val.Type {
		case value.TypeWord:
			word, _ := val.AsWord()
			if seenFields[word] {
				return value.NoneVal(), verror.NewScriptError(
					"object-field-duplicate",
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true
			initializers[word] = []value.Value{value.NoneVal()}

		case value.TypeSetWord:
			word, _ := val.AsWord()
			if seenFields[word] {
				return value.NoneVal(), verror.NewScriptError(
					"object-field-duplicate",
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true

			i++
			if i >= len(spec.Elements) {
				return value.NoneVal(), verror.NewScriptError(
					verror.ErrIDInvalidSyntax,
					[3]string{"context", "set-word-without-value", word},
				)
			}

			// Collect all values for this initializer until next SetWord
			initVals := []value.Value{}
			for i < len(spec.Elements) {
				nextVal := spec.Elements[i]
				if nextVal.Type == value.TypeSetWord {
					i--
					break
				}
				initVals = append(initVals, nextVal)
				i++
			}

			if len(initVals) == 0 {
				initVals = []value.Value{value.NoneVal()}
			}
			initializers[word] = initVals

		default:
			continue
		}
	} // Create object frame with NO parent (-1 = isolated)
	objFrame := frame.NewObjectFrame(-1, fields, nil)

	// Register frame
	frameIdx := mgr.RegisterFrame(objFrame)

	// Mark frame as captured so it persists after PopFrameContext
	// Object frames must remain in frameStore for path traversal
	mgr.MarkFrameCaptured(frameIdx)

	// Push object frame as active context
	mgr.PushFrameContext(objFrame)
	defer mgr.PopFrameContext()

	// Evaluate initializers in isolated context
	for _, field := range fields {
		initVals := initializers[field]

		evaled, err := mgr.Do_Blk(initVals)
		if err != nil {
			return value.NoneVal(), err
		}

		objFrame.Bind(field, evaled)
	}

	// Create ObjectInstance
	obj := value.NewObject(frameIdx, fields, nil)

	return value.ObjectVal(obj), nil
}
