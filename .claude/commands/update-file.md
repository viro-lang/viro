Modify only current file. If a compilation error originates from a different file than ignore it and continue with other changes.
change each direct usage of `value.Value` to the interface `core.Value`.
change each direct usage of `value.ValueType` to the interface `core.ValueType`
cahnge each direct usage of `*frame.Frame` to the interface `core.Frame`
change each direct usage of `Evaluator` to `core.Evaluator`.
change each `*verror.Error` to `error`. Do not remove the `verror` package import and do NOT replace error creating code (like `verror.NewScriptErrror(...)`) since it implements the `error` interface and will work.
update casts like `val.AsBlock()` to updated form `value.AsBlock(val)`
when there is something like `val.GetType().toString()` you can use `value.TypeToString(val.GetType())`
when you encounter function `typeNameFor(val.GetType())` rewrite it to use the `value.TypeToString(val.GetType())`
`Equals` is always a method on a value so it should be like `result.Equals(something)`, and NOT like `value.Equals(result, something)`
`String` is always a method so use as `result.String()` and NOT `value.ToString(result)`
