package native

import (
	"strings"

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
		return value.NoneVal(), arityError("set", 2, len(args))
	}

	if args[0].Type != value.TypeWord {
		return value.NoneVal(), typeError("set", "word", args[0])
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
		return value.NoneVal(), arityError("get", 1, len(args))
	}

	if args[0].Type != value.TypeWord {
		return value.NoneVal(), typeError("get", "word", args[0])
	}

	symbol, _ := args[0].AsWord()
	return eval.Do_Next(value.GetWordVal(symbol))
}

// TypeQ implements the `type?` native.
//
// Contract: type? value -> word representing type name
func TypeQ(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("type?", 1, len(args))
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
	case value.TypeDatatype:
		return "datatype!"
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

func buildObjectSpec(nativeName string, spec *value.BlockValue) ([]string, map[string][]value.Value, *verror.Error) {
	fields := []string{}
	initializers := make(map[string][]value.Value)
	seenFields := make(map[string]bool)

	for i := 0; i < len(spec.Elements); i++ {
		val := spec.Elements[i]

		switch val.Type {
		case value.TypeWord:
			word, _ := val.AsWord()
			if isReservedField(word) {
				return nil, nil, verror.NewScriptError(
					verror.ErrIDReservedField,
					[3]string{word, "", ""},
				)
			}
			if seenFields[word] {
				return nil, nil, verror.NewScriptError(
					verror.ErrIDObjectFieldDup,
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true
			initializers[word] = []value.Value{value.NoneVal()}

		case value.TypeSetWord:
			word, _ := val.AsWord()
			if isReservedField(word) {
				return nil, nil, verror.NewScriptError(
					verror.ErrIDReservedField,
					[3]string{word, "", ""},
				)
			}
			if seenFields[word] {
				return nil, nil, verror.NewScriptError(
					verror.ErrIDObjectFieldDup,
					[3]string{word, "", ""},
				)
			}
			fields = append(fields, word)
			seenFields[word] = true

			i++
			if i >= len(spec.Elements) {
				return nil, nil, verror.NewScriptError(
					verror.ErrIDInvalidSyntax,
					[3]string{nativeName, "set-word-without-value", word},
				)
			}

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
	}

	return fields, initializers, nil
}

func instantiateObject(mgr frameManager, eval Evaluator, lexicalParent int, prototype *value.ObjectInstance, fields []string, initializers map[string][]value.Value) (value.Value, *verror.Error) {
	objFrame := frame.NewObjectFrame(lexicalParent, fields, nil)

	frameIdx := mgr.RegisterFrame(objFrame)
	mgr.MarkFrameCaptured(frameIdx)

	mgr.PushFrameContext(objFrame)
	defer mgr.PopFrameContext()

	for _, field := range fields {
		initVals := initializers[field]

		evaled, err := mgr.Do_Blk(initVals)
		if err != nil {
			return value.NoneVal(), err
		}

		objFrame.Bind(field, evaled)
	}

	obj := value.NewObject(frameIdx, fields, nil)
	if prototype != nil {
		obj.ParentProto = prototype
		obj.Parent = prototype.FrameIndex // For backward compatibility
		mgr.MarkFrameCaptured(prototype.FrameIndex)
	}

	return value.ObjectVal(obj), nil
}

func isReservedField(name string) bool {
	switch strings.ToLower(name) {
	case "parent", "spec":
		return true
	default:
		return false
	}
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
		return value.NoneVal(), arityError("object", 1, len(args))
	}

	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), typeError("object", "block", args[0])
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

	fields, initializers, err := buildObjectSpec("object", spec)
	if err != nil {
		return value.NoneVal(), err
	}

	// Create object frame with parent set to current frame
	parentIdx := 0 // Global frame as parent
	if provider, ok := eval.(frameProvider); ok {
		parentIdx = provider.CurrentFrameIndex()
	}

	return instantiateObject(mgr, eval, parentIdx, nil, fields, initializers)
}

// Context implements the `context` native.
//
// Contract (Feature 002, FR-009): context spec
// - Alias for object but with isolated scope (no parent frame)
// - spec: block describing fields and optional initial values
// - Returns object! instance with isolated frame
func Context(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("context", 1, len(args))
	}

	if args[0].Type != value.TypeBlock {
		return value.NoneVal(), typeError("context", "block", args[0])
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

	fields, initializers, err := buildObjectSpec("context", spec)
	if err != nil {
		return value.NoneVal(), err
	}

	return instantiateObject(mgr, eval, -1, nil, fields, initializers)
}

// Make implements the `make` native supporting object prototypes.
//
// Contract: make target spec
// - When target is word "object!" create new base object (prototype = none)
// - When target is object value (or word resolving to object), use it as prototype
// - Spec must be block describing fields/initializers (same as object)
func Make(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("make", 2, len(args))
	}

	specVal := args[1]
	if specVal.Type != value.TypeBlock {
		return value.NoneVal(), typeError("make spec", "block", specVal)
	}

	specBlock, ok := specVal.AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("make spec missing block payload", [3]string{})
	}

	mgr, ok := eval.(frameManager)
	if !ok {
		return value.NoneVal(), verror.NewInternalError(
			"internal-error",
			[3]string{"make", "frame-manager-unavailable", ""},
		)
	}

	fields, initializers, err := buildObjectSpec("make", specBlock)
	if err != nil {
		return value.NoneVal(), err
	}

	var prototype *value.ObjectInstance
	target := args[0]

	// Handle datatype literal like object!
	if target.Type == value.TypeDatatype {
		dtName, _ := target.AsDatatype()
		if strings.EqualFold(dtName, "object!") {
			prototype = nil
			goto instantiate
		}
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"make", "object!", dtName},
		)
	}

	// Evaluate target to get object prototype
	for {
		switch target.Type {
		case value.TypeWord:
			word, _ := target.AsWord()
			evaluated, evalErr := eval.Do_Next(value.WordVal(word))
			if evalErr != nil {
				return value.NoneVal(), evalErr
			}
			target = evaluated
			continue

		case value.TypeGetWord:
			symbol, _ := target.AsWord()
			evaluated, evalErr := eval.Do_Next(value.GetWordVal(symbol))
			if evalErr != nil {
				return value.NoneVal(), evalErr
			}
			target = evaluated
			continue

		case value.TypeObject:
			obj, _ := target.AsObject()
			prototype = obj
		default:
			return value.NoneVal(), typeError("make target", "object", target)
		}
		break
	}

instantiate:
	parentIdx := 0
	if provider, ok := eval.(frameProvider); ok {
		parentIdx = provider.CurrentFrameIndex()
	}

	return instantiateObject(mgr, eval, parentIdx, prototype, fields, initializers)
}
