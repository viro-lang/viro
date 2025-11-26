package native

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TypeOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("type-of", 1, len(args))
	}

	val := args[0]

	if val.GetType() == value.TypeFunction {
		if fn, ok := value.AsFunctionValue(val); ok {
			if fn.Type == value.FuncNative {
				return value.NewWordVal("native!"), nil
			}
			return value.NewWordVal("function!"), nil
		}
	}

	typeName := value.TypeToString(val.GetType())
	return value.NewWordVal(typeName), nil
}

func SpecOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("spec-of", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunctionValue(val)
		specElements := []core.Value{}
		for _, param := range fn.Params {
			specElements = append(specElements, value.NewWordVal(param.Name))
		}
		return value.NewBlockVal(specElements), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		specElements := []core.Value{}
		bindings := obj.GetAllFieldsWithProto()
		for _, binding := range bindings {
			specElements = append(specElements, value.NewWordVal(binding.Symbol))
		}
		return value.NewBlockVal(specElements), nil

	default:
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDSpecUnsupported,
			[3]string{value.TypeToString(val.GetType()), "", ""},
		)
	}
}

func BodyOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("body-of", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunctionValue(val)
		if fn.Type == value.FuncNative {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDNoBody,
				[3]string{"native functions have no accessible body", "", ""},
			)
		}
		if fn.Body == nil {
			return value.NewBlockVal([]core.Value{}), nil
		}
		bodyElements := make([]core.Value, len(fn.Body.Elements))
		copy(bodyElements, fn.Body.Elements)
		return value.NewBlockVal(bodyElements), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		bodyElements := []core.Value{}
		bindings := obj.GetAllFieldsWithProto()
		for _, binding := range bindings {
			bodyElements = append(bodyElements, value.NewSetWordVal(binding.Symbol))
			bodyElements = append(bodyElements, value.NewNoneVal()) // Placeholder
		}
		return value.NewBlockVal(bodyElements), nil

	default:
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDNoBody,
			[3]string{value.TypeToString(val.GetType()), "has no accessible body", ""},
		)
	}
}

func WordsOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("words-of", 1, len(args))
	}

	val := args[0]

	if val.GetType() != value.TypeObject {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"object!", value.TypeToString(val.GetType()), ""},
		)
	}

	obj, _ := value.AsObject(val)

	bindings := obj.GetAllFieldsWithProto()
	wordElements := make([]core.Value, len(bindings))
	for i, binding := range bindings {
		wordElements[i] = value.NewWordVal(binding.Symbol)
	}

	return value.NewBlockVal(wordElements), nil
}

func ValuesOf(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("values-of", 1, len(args))
	}

	val := args[0]

	if val.GetType() != value.TypeObject {
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDTypeMismatch,
			[3]string{"object!", value.TypeToString(val.GetType()), ""},
		)
	}

	obj, _ := value.AsObject(val)

	bindings := obj.GetAllFieldsWithProto()
	valueElements := make([]core.Value, len(bindings))
	for i, binding := range bindings {
		valueElements[i] = binding.Value
	}

	return value.NewBlockVal(valueElements), nil
}

func Source(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("source", 1, len(args))
	}

	val := args[0]

	switch val.GetType() {
	case value.TypeFunction:
		fn, _ := value.AsFunctionValue(val)
		specElements := []core.Value{}
		for _, param := range fn.Params {
			specElements = append(specElements, value.NewWordVal(param.Name))
		}
		specStr := formatBlock(specElements)

		bodyStr := "[]"
		if fn.Body != nil {
			bodyStr = formatBlock(fn.Body.Elements)
		}
		source := fmt.Sprintf("fn %s %s", specStr, bodyStr)
		return value.NewStrVal(source), nil

	case value.TypeObject:
		obj, _ := value.AsObject(val)
		bindings := obj.GetAllFieldsWithProto()
		fields := []string{}
		for _, binding := range bindings {
			fields = append(fields, binding.Symbol)
		}
		fieldsStr := strings.Join(fields, " ")
		source := fmt.Sprintf("object [%s]", fieldsStr)
		return value.NewStrVal(source), nil

	default:
		return value.NewNoneVal(), verror.NewScriptError(
			verror.ErrIDSourceUnsupported,
			[3]string{value.TypeToString(val.GetType()), "", ""},
		)
	}
}

func Has(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("has?", 2, len(args))
	}

	targetVal := args[0]
	soughtVal := args[1]

	if targetVal.GetType() == value.TypeObject {
		var fieldName string
		if soughtVal.GetType() == value.TypeWord {
			fieldName, _ = value.AsWordValue(soughtVal)
		} else if soughtVal.GetType() == value.TypeString {
			str, _ := value.AsStringValue(soughtVal)
			fieldName = str.String()
		} else {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"word! or string!", value.TypeToString(soughtVal.GetType()), ""},
			)
		}

		obj, _ := value.AsObject(targetVal)
		_, exists := obj.GetFieldWithProto(fieldName)
		return value.NewLogicVal(exists), nil
	} else {
		seriesVal, err := assertSeries(targetVal)
		if err != nil {
			return value.NewNoneVal(), err
		}
		return value.NewLogicVal(seriesHasValue(seriesVal, soughtVal)), nil
	}
}

func formatBlock(elements []core.Value) string {
	parts := make([]string, len(elements))
	for i, elem := range elements {
		parts[i] = elem.Mold()
	}
	return "[" + strings.Join(parts, " ") + "]"
}
