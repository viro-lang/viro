package native

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

type frameProvider interface {
	CurrentFrameIndex() int
	MarkFrameCaptured(index int)
}

// Fn implements the function definition native.
//
// Contract per contracts/function.md:
//
//	fn [params] [body] -> function value
//
// - Parameters block defines positional parameters and refinements
// - Body block captures function code (stored as block value)
// - Returns a user-defined function with captured lexical parent
func Fn(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("fn", 2, len(args))
	}

	paramsVal := args[0]
	if paramsVal.Type != value.TypeBlock {
		return value.NoneVal(), typeError("fn parameters", "block", paramsVal)
	}

	paramsBlock, ok := paramsVal.AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("fn parameters missing block payload", [3]string{})
	}

	specs, err := parseParamSpecs(paramsBlock)
	if err != nil {
		return value.NoneVal(), err
	}

	bodyVal := args[1]
	if bodyVal.Type != value.TypeBlock {
		return value.NoneVal(), typeError("fn body", "block", bodyVal)
	}

	bodyBlock, ok := bodyVal.AsBlock()
	if !ok {
		return value.NoneVal(), verror.NewInternalError("fn body missing block payload", [3]string{})
	}

	bodyClone := bodyBlock.Clone()
	bodyClone.SetIndex(0)

	parentIndex := -1
	if provider, ok := eval.(frameProvider); ok {
		parentIndex = provider.CurrentFrameIndex()
		if parentIndex >= 0 {
			provider.MarkFrameCaptured(parentIndex)
		}
	}

	fnValue := value.NewUserFunction("", specs, bodyClone, parentIndex)
	return value.FuncVal(fnValue), nil
}

func parseParamSpecs(block *value.BlockValue) ([]value.ParamSpec, *verror.Error) {
	specs := make([]value.ParamSpec, 0, len(block.Elements))
	seen := make(map[string]struct{})

	for i := 0; i < len(block.Elements); i++ {
		elem := block.Elements[i]
		if elem.Type != value.TypeWord {
			return nil, invalidParamSpecError(elem.String())
		}

		symbol, _ := elem.AsWord()
		if strings.HasPrefix(symbol, "--") {
			name := strings.TrimPrefix(symbol, "--")
			if name == "" {
				return nil, verror.NewScriptError(
					verror.ErrIDInvalidOperation,
					[3]string{"Invalid refinement name", "", ""},
				)
			}

			if _, exists := seen[name]; exists {
				return nil, duplicateParamError(name)
			}
			seen[name] = struct{}{}

			takesValue := false
			if i+1 < len(block.Elements) && block.Elements[i+1].Type == value.TypeBlock {
				takesValue = true
				i++ // Skip metadata block (type/docstring)
			}

			specs = append(specs, value.ParamSpec{
				Name:       name,
				Type:       value.TypeNone,
				Optional:   true,
				Refinement: true,
				TakesValue: takesValue,
			})
			continue
		}

		name := symbol
		if _, exists := seen[name]; exists {
			return nil, duplicateParamError(name)
		}
		seen[name] = struct{}{}

		specs = append(specs, value.ParamSpec{
			Name:       name,
			Type:       value.TypeNone,
			Optional:   false,
			Refinement: false,
			TakesValue: false,
		})
	}

	return specs, nil
}
