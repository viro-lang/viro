package native

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Set implements the `set` native.
//
// Contract: set 'word value
// - First argument must be a lit-word (receives unevaluated word)
// - Second argument is any value (already evaluated)
// - Binds word in current frame and returns the value
func Set(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("set", 2, len(args))
	}

	if args[0].Type != value.TypeLitWord {
		return value.NoneVal(), typeError("set", "lit-word", args[0])
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
// Contract: get 'word
// - Argument must be a lit-word (receives unevaluated word symbol)
// - Returns bound value from current frame chain
func Get(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("get", 1, len(args))
	}

	if args[0].Type != value.TypeLitWord {
		return value.NoneVal(), typeError("get", "lit-word", args[0])
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

	// Emit trace event for object creation (Feature 002, T097)
	TraceObjectCreate(frameIdx, len(fields))

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
func Object(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
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
func Context(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
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
func Make(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
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

// Select implements the `select` native for object field lookup with default.
//
// Contract (Feature 002, FR-014): select object field --default value
// - object: object! to query
// - field: word! or string! representing field name
// - --default: optional refinement providing fallback value when field missing
// - Returns field value or default (or none! if no default provided)
func Select(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) < 2 {
		return value.NoneVal(), arityError("select", 2, len(args))
	}

	targetVal := args[0]
	fieldVal := args[1]

	// Extract field name from word or string
	var fieldName string
	switch fieldVal.Type {
	case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
		fieldName, _ = fieldVal.AsWord()
	case value.TypeString:
		str, _ := fieldVal.AsString()
		fieldName = str.String()
	default:
		return value.NoneVal(), typeError("select field", "word or string", fieldVal)
	}

	// Check for --default refinement value
	defaultVal, hasDefault := refValues["default"]
	if !hasDefault {
		defaultVal = value.NoneVal()
	}

	// Handle object selection
	if targetVal.Type == value.TypeObject {
		obj, _ := targetVal.AsObject()

		// Type-assert to access frame operations
		mgr, ok := eval.(frameManager)
		if !ok {
			return value.NoneVal(), verror.NewInternalError(
				"internal-error",
				[3]string{"select", "frame-manager-unavailable", ""},
			)
		}

		// Look up field in object's frame
		objFrame := mgr.GetFrameByIndex(obj.FrameIndex)
		if objFrame == nil {
			return value.NoneVal(), verror.NewInternalError(
				"internal-error",
				[3]string{"select", "invalid-frame-index", ""},
			)
		}

		// Try to get the field value
		if result, found := objFrame.Get(fieldName); found {
			TraceObjectFieldRead(obj.FrameIndex, fieldName, true)
			return result, nil
		}

		// Check parent prototype chain
		current := obj.ParentProto
		for current != nil {
			parentFrame := mgr.GetFrameByIndex(current.FrameIndex)
			if parentFrame != nil {
				if result, found := parentFrame.Get(fieldName); found {
					TraceObjectFieldRead(current.FrameIndex, fieldName, true)
					return result, nil
				}
			}
			current = current.ParentProto
		}

		// Field not found - return default or none (not an error)
		TraceObjectFieldRead(obj.FrameIndex, fieldName, false)
		if hasDefault {
			return defaultVal, nil
		}

		return value.NoneVal(), nil
	}

	// Handle block selection (key-value pairs)
	if targetVal.Type == value.TypeBlock {
		block, _ := targetVal.AsBlock()
		elements := block.Elements

		// Search for key-value pairs
		for i := 0; i+1 < len(elements); i += 2 {
			key := elements[i]
			var keyStr string

			switch key.Type {
			case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
				keyStr, _ = key.AsWord()
			case value.TypeString:
				str, _ := key.AsString()
				keyStr = str.String()
			default:
				continue
			}

			if keyStr == fieldName {
				return elements[i+1], nil
			}
		}

		// Not found - return default or none
		if hasDefault {
			return defaultVal, nil
		}
		return value.NoneVal(), nil
	}

	return value.NoneVal(), typeError("select target", "object or block", targetVal)
}

// Put implements the `put` native for object field mutation.
//
// Contract (Feature 002, FR-014): put object field value
// - object: object! to modify
// - field: word! or string! representing field name
// - value: any value to assign to the field
// - Updates field in object's frame after optional type validation
// - Returns the assigned value
func Put(args []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 3 {
		return value.NoneVal(), arityError("put", 3, len(args))
	}

	targetVal := args[0]
	fieldVal := args[1]
	newVal := args[2]

	// Extract field name
	var fieldName string
	switch fieldVal.Type {
	case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
		fieldName, _ = fieldVal.AsWord()
	case value.TypeString:
		str, _ := fieldVal.AsString()
		fieldName = str.String()
	default:
		return value.NoneVal(), typeError("put field", "word or string", fieldVal)
	}

	// Only support objects for now
	if targetVal.Type != value.TypeObject {
		return value.NoneVal(), typeError("put target", "object", targetVal)
	}

	obj, _ := targetVal.AsObject()

	// Type-assert to access frame operations
	mgr, ok := eval.(frameManager)
	if !ok {
		return value.NoneVal(), verror.NewInternalError(
			"internal-error",
			[3]string{"put", "frame-manager-unavailable", ""},
		)
	}

	// Get object's frame
	objFrame := mgr.GetFrameByIndex(obj.FrameIndex)
	if objFrame == nil {
		return value.NoneVal(), verror.NewInternalError(
			"internal-error",
			[3]string{"put", "invalid-frame-index", ""},
		)
	}

	// Check if field exists in manifest
	fieldIndex := -1
	for i, word := range obj.Manifest.Words {
		if word == fieldName {
			fieldIndex = i
			break
		}
	}

	if fieldIndex == -1 {
		// Field doesn't exist - error per contract (no dynamic field addition)
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDNoSuchField,
			[3]string{fieldName, "", ""},
		)
	}

	// Optional: Type validation if type hint is present
	expectedType := obj.Manifest.Types[fieldIndex]
	if expectedType != value.TypeNone && expectedType != newVal.Type {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{typeNameFor(expectedType), typeNameFor(newVal.Type), fieldName},
		)
	}

	// Update the field in the frame
	objFrame.Set(fieldName, newVal)

	// Emit trace event for field write (Feature 002, T097)
	TraceObjectFieldWrite(obj.FrameIndex, fieldName, newVal.String())

	return newVal, nil
}
