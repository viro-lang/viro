package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// NewNativeFunction creates a native (built-in) function.
func NewNativeFunction(name string, params []value.ParamSpec, impl func([]core.Value, map[string]core.Value, core.Evaluator) (core.Value, error)) *value.FunctionValue {
	return &value.FunctionValue{
		Type:   value.FuncNative,
		Name:   name,
		Params: params,
		Body:   nil,
		Native: impl,
		Parent: -1,
	}
}
