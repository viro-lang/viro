// Package native - action creation utilities
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func CreateAction(name string, params []value.ParamSpec, doc *NativeDoc) core.Value {
	dispatcher := func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
		if len(args) == 0 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{name, "at least 1", "0"},
			)
		}

		firstArg := args[0]
		firstArgType := firstArg.GetType()

		typeFrame, found := frame.GetTypeFrame(firstArgType)
		if !found {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDActionNoImpl,
				[3]string{name, value.TypeToString(firstArgType), ""},
			)
		}

		impl, found := typeFrame.Get(name)
		if !found {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDActionNoImpl,
				[3]string{name, value.TypeToString(firstArgType), ""},
			)
		}

		fn, ok := value.AsFunctionValue(impl)
		if !ok {
			return value.NewNoneVal(), verror.NewInternalError(
				"action-frame-corrupt",
				[3]string{name, value.TypeToString(firstArgType), "type frame binding is not a function"},
			)
		}

		return fn.Native(args, refValues, eval)
	}

	return value.NewFuncVal(value.NewNativeFunction(name, params, dispatcher, false, doc))
}

func RegisterActionImpl(typ core.ValueType, actionName string, fn *value.FunctionValue) {
	typeFrame, found := frame.GetTypeFrame(typ)
	if !found {
		panic("RegisterActionImpl: type frame not found for " + value.TypeToString(typ))
	}

	typeFrame.Bind(actionName, value.NewFuncVal(fn))
}
