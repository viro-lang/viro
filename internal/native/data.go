package native

import (
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Set implements the `set` native.
//
// Contract: set 'word value
// - First argument must be a lit-word (receives unevaluated word)
// - Second argument is any value (already evaluated)
// - Binds word in current frame and returns the value
func Set(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("set", 2, len(args))
	}

	if args[0].GetType() != value.TypeLitWord {
		return value.NewNoneVal(), typeError("set", "lit-word", args[0])
	}

	symbol, _ := value.AsWordValue(args[0])
	assignment := []core.Value{value.NewSetWordVal(symbol), args[1]}

	result, err := eval.DoBlock(assignment)
	if err != nil {
		return value.NewNoneVal(), err
	}

	return result, nil
}

// Get implements the `get` native.
//
// Contract: get 'word
// - Argument must be a lit-word (receives unevaluated word symbol)
// - Returns bound value from current frame chain
func Get(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("get", 1, len(args))
	}

	if args[0].GetType() != value.TypeLitWord {
		return value.NewNoneVal(), typeError("get", "lit-word", args[0])
	}

	symbol, _ := value.AsWordValue(args[0])
	_, result, err := eval.EvaluateExpression([]core.Value{value.NewGetWordVal(symbol)}, 0)
	return result, err
}

// TypeQ implements the `type?` native.
//
// Contract: type? value -> word representing type name
func TypeQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("type?", 1, len(args))
	}

	typeName := value.TypeToString(args[0].GetType())
	return value.NewWordVal(typeName), nil
}

// Form implements the `form` native.
//
// Contract: form value -> string! human-readable representation
// Returns display-friendly string format (no brackets on blocks, no quotes on strings, objects as multi-line field display)
func Form(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("form", 1, len(args))
	}

	val := args[0]
	return value.NewStrVal(val.Form()), nil
}

// Join implements the `join` native.
//
// Contract: join value1 value2 -> string!
// - Converts both values to strings using form
// - Concatenates them
// - Returns new string
func Join(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("join", 2, len(args))
	}

	str1 := args[0].Form()
	str2 := args[1].Form()
	return value.NewStrVal(str1 + str2), nil
}

// Mold implements the `mold` native.
//
// Contract: mold value -> string! code-readable representation
// Returns serialization-friendly string format (brackets on blocks, quotes on strings, objects as make object! [...])
func Mold(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("mold", 1, len(args))
	}

	val := args[0]

	return value.NewStrVal(val.Mold()), nil
}

func buildObjectSpec(nativeName string, spec *value.BlockValue) ([]string, map[string][]core.Value, error) {
	fields := []string{}
	initializers := make(map[string][]core.Value)
	seenFields := make(map[string]bool)

	for i := 0; i < len(spec.Elements); i++ {
		val := spec.Elements[i]

		switch val.GetType() {
		case value.TypeWord:
			word, _ := value.AsWordValue(val)
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
			initializers[word] = []core.Value{value.NewNoneVal()}

		case value.TypeSetWord:
			word, _ := value.AsWordValue(val)
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

			initVals := []core.Value{}
			for i < len(spec.Elements) {
				nextVal := spec.Elements[i]
				if nextVal.GetType() == value.TypeSetWord {
					i--
					break
				}
				initVals = append(initVals, nextVal)
				i++
			}

			if len(initVals) == 0 {
				initVals = []core.Value{value.NewNoneVal()}
			}
			initializers[word] = initVals

		default:
			continue
		}
	}

	return fields, initializers, nil
}

func instantiateObject(eval core.Evaluator, lexicalParent int, prototype *value.ObjectInstance, fields []string, initializers map[string][]core.Value) (core.Value, error) {
	ownedFrame := frame.NewObjectFrame(lexicalParent, fields, nil)

	// Evaluate initializers in a temporary frame context
	eval.PushFrameContext(ownedFrame)
	defer eval.PopFrameContext()

	for _, field := range fields {
		initVals := initializers[field]

		evaled, err := eval.DoBlock(initVals)
		if err != nil {
			return value.NewNoneVal(), err
		}

		ownedFrame.Bind(field, evaled)
	}

	obj := value.NewObject(ownedFrame)
	if prototype != nil {
		obj.ParentProto = prototype
	}

	trace.TraceObjectCreate(len(fields))

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
func Object(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("object", 1, len(args))
	}

	if args[0].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("object", "block", args[0])
	}

	spec, _ := value.AsBlockValue(args[0])

	fields, initializers, err := buildObjectSpec("object", spec)
	if err != nil {
		return value.NewNoneVal(), err
	}

	// Create object frame with parent set to current frame
	parentIdx := eval.CurrentFrameIndex() // Global frame as parent

	return instantiateObject(eval, parentIdx, nil, fields, initializers)
}

// Context implements the `context` native.
//
// Contract (Feature 002, FR-009): context spec
// - Alias for object but with isolated scope (no parent frame)
// - spec: block describing fields and optional initial values
// - Returns object! instance with isolated frame
func Context(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("context", 1, len(args))
	}

	if args[0].GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("context", "block", args[0])
	}

	spec, _ := value.AsBlockValue(args[0])

	fields, initializers, err := buildObjectSpec("context", spec)
	if err != nil {
		return value.NewNoneVal(), err
	}

	return instantiateObject(eval, -1, nil, fields, initializers)
}

// Make implements the `make` native supporting object prototypes.
//
// Contract: make target spec
// - When target is word "object!" create new base object (prototype = none)
// - When target is object value (or word resolving to object), use it as prototype
// - Spec must be block describing fields/initializers (same as object)
func Make(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("make", 2, len(args))
	}

	specVal := args[1]
	if specVal.GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("make spec", "block", specVal)
	}

	specBlock, ok := value.AsBlockValue(specVal)
	if !ok {
		return value.NewNoneVal(), verror.NewInternalError("make spec missing block payload", [3]string{})
	}

	fields, initializers, err := buildObjectSpec("make", specBlock)
	if err != nil {
		return value.NewNoneVal(), err
	}

	var prototype *value.ObjectInstance
	target := args[0]

	// Handle datatype literal like object!
	if target.GetType() == value.TypeDatatype {
		dtName, _ := value.AsDatatypeValue(target)
		if strings.EqualFold(dtName, "object!") {
			prototype = nil
			goto instantiate
		}
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"make", "object!", dtName},
		)
	}

	// Evaluate target to get object prototype
	for {
		switch target.GetType() {
		case value.TypeWord:
			word, _ := value.AsWordValue(target)
			_, evaluated, evalErr := eval.EvaluateExpression([]core.Value{value.NewWordVal(word)}, 0)
			if evalErr != nil {
				return value.NewNoneVal(), evalErr
			}
			target = evaluated
			continue

		case value.TypeGetWord:
			symbol, _ := value.AsWordValue(target)
			_, evaluated, evalErr := eval.EvaluateExpression([]core.Value{value.NewGetWordVal(symbol)}, 0)
			if evalErr != nil {
				return value.NewNoneVal(), evalErr
			}
			target = evaluated
			continue

		case value.TypeObject:
			obj, _ := value.AsObject(target)
			prototype = obj
		default:
			return value.NewNoneVal(), typeError("make target", "object", target)
		}
		break
	}

instantiate:
	parentIdx := eval.CurrentFrameIndex()
	return instantiateObject(eval, parentIdx, prototype, fields, initializers)
}

// Select implements the `select` native for object field lookup with default.
//
// Contract (Feature 002, FR-014): select object field --default value
// - object: object! to query
// - field: word! or string! representing field name
// - --default: optional refinement providing fallback value when field missing
// - Returns field value or default (or none! if no default provided)
func Select(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) < 2 {
		return value.NewNoneVal(), arityError("select", 2, len(args))
	}

	targetVal := args[0]
	fieldVal := args[1]

	// Extract field name from word or string
	var fieldName string
	switch fieldVal.GetType() {
	case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
		fieldName, _ = value.AsWordValue(fieldVal)
	case value.TypeString:
		str, _ := value.AsStringValue(fieldVal)
		fieldName = str.String()
	default:
		return value.NewNoneVal(), typeError("select field", "word or string", fieldVal)
	}

	// Check for --default refinement value
	defaultVal, hasDefault := refValues["default"]
	if !hasDefault {
		defaultVal = value.NewNoneVal()
	}

	// Handle object selection
	if targetVal.GetType() == value.TypeObject {
		obj, _ := value.AsObject(targetVal)

		// Use owned frame to get field value with prototype chain traversal
		if result, found := obj.GetFieldWithProto(fieldName); found {
			trace.TraceObjectFieldRead(fieldName, true)
			return result, nil
		}

		// Field not found - return default or none (not an error)
		trace.TraceObjectFieldRead(fieldName, false) // FrameIndex removed
		if hasDefault {
			return defaultVal, nil
		}

		return value.NewNoneVal(), nil
	}

	// Handle block selection (key-value pairs)
	if targetVal.GetType() == value.TypeBlock {
		block, _ := value.AsBlockValue(targetVal)
		elements := block.Elements

		// Search for key-value pairs
		for i := 0; i+1 < len(elements); i += 2 {
			key := elements[i]
			var keyStr string

			switch key.GetType() {
			case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
				keyStr, _ = value.AsWordValue(key)
			case value.TypeString:
				str, _ := value.AsStringValue(key)
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
		return value.NewNoneVal(), nil
	}

	return value.NewNoneVal(), typeError("select target", "object or block", targetVal)
}

// Put implements the `put` native for object field mutation.
//
// Contract (Feature 002, FR-014): put object field value
// - object: object! to modify
// - field: word! or string! representing field name
// - value: any value to assign to the field
// - Updates field in object's frame after optional type validation
// - Returns the assigned value
func Put(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("put", 3, len(args))
	}

	targetVal := args[0]
	fieldVal := args[1]
	newVal := args[2]

	// Extract field name
	var fieldName string
	switch fieldVal.GetType() {
	case value.TypeWord, value.TypeGetWord, value.TypeLitWord:
		fieldName, _ = value.AsWordValue(fieldVal)
	case value.TypeString:
		str, _ := value.AsStringValue(fieldVal)
		fieldName = str.String()
	default:
		return value.NewNoneVal(), typeError("put field", "word or string", fieldVal)
	}

	// Only support objects for now
	if targetVal.GetType() != value.TypeObject {
		return value.NewNoneVal(), typeError("put target", "object", targetVal)
	}

	obj, _ := value.AsObject(targetVal)

	// Set the field using owned frame (creates field if it doesn't exist)
	obj.SetField(fieldName, newVal)

	// Emit trace event for field write (Feature 002, T097)
	trace.TraceObjectFieldWrite(fieldName, newVal.Form())

	return newVal, nil
}

// ToInteger implements the `to-integer` native for converting values to integers.
//
// Contract: to-integer value -> integer!
// - Converts integer (pass-through), decimal (truncate), string (parse) to integer
// - Returns error for invalid conversions
func ToInteger(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("to-integer", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeInteger:
		return val, nil

	case value.TypeDecimal:
		if dec, ok := value.AsDecimal(val); ok && dec != nil && dec.Magnitude != nil {
			i, ok := dec.Magnitude.Int64()
			if !ok {
				return value.NewNoneVal(), verror.NewMathError("to-integer-overflow", [3]string{dec.String(), "", ""})
			}
			return value.NewIntVal(i), nil
		}
		return value.NewNoneVal(), verror.NewScriptError("to-integer-invalid-decimal", [3]string{"", "", ""})

	case value.TypeString:
		if str, ok := value.AsStringValue(val); ok {
			goStr := str.String()
			i, err := strconv.ParseInt(goStr, 10, 64)
			if err != nil {
				return value.NewNoneVal(), verror.NewScriptError("to-integer-invalid-string", [3]string{goStr, "", ""})
			}
			return value.NewIntVal(i), nil
		}
		return value.NewNoneVal(), verror.NewScriptError("to-integer-invalid-string", [3]string{"", "", ""})

	default:
		return value.NewNoneVal(), typeError("to-integer", "integer, decimal, or string", val)
	}
}

// ToDecimal implements the `to-decimal` native for converting values to decimals.
//
// Contract: to-decimal value -> decimal!
// - Converts integer (exact), decimal (pass-through), string (parse) to decimal
// - Returns error for invalid conversions
func ToDecimal(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("to-decimal", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeInteger:
		if i, ok := value.AsIntValue(val); ok {
			d := decimal.New(i, 0)
			return value.DecimalVal(d, 0), nil
		}
		return value.NewNoneVal(), verror.NewScriptError("to-decimal-invalid-integer", [3]string{"", "", ""})

	case value.TypeDecimal:
		return val, nil

	case value.TypeString:
		if str, ok := value.AsStringValue(val); ok {
			goStr := str.String()
			d := new(decimal.Big)
			_, ok := d.SetString(goStr)
			if !ok || d.IsNaN(0) {
				return value.NewNoneVal(), verror.NewScriptError("to-decimal-invalid-string", [3]string{goStr, "", ""})
			}
			scale := int16(0)
			if idx := findDecimalPoint(goStr); idx >= 0 {
				scale = int16(len(goStr) - idx - 1)
			}
			return value.DecimalVal(d, scale), nil
		}
		return value.NewNoneVal(), verror.NewScriptError("to-decimal-invalid-string", [3]string{"", "", ""})

	default:
		return value.NewNoneVal(), typeError("to-decimal", "integer, decimal, or string", val)
	}
}

// ToString implements the `to-string` native for converting values to strings.
//
// Contract: to-string value -> string!
// - Converts any value to string using its Form representation
func ToString(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("to-string", 1, len(args))
	}

	val := args[0]
	return value.NewStrVal(val.Form()), nil
}
