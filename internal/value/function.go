package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/docmodel"
)

type FunctionType uint8

const (
	FuncNative FunctionType = iota
	FuncUser
)

type ParamSpec struct {
	Name       string
	Type       core.ValueType
	Optional   bool
	Refinement bool
	TakesValue bool
	Eval       bool
}

func NewParamSpec(name string, eval bool) ParamSpec {
	return ParamSpec{
		Name:       name,
		Type:       TypeNone,
		Optional:   false,
		Refinement: false,
		TakesValue: false,
		Eval:       eval,
	}
}

func NewRefinementSpec(name string, takesValue bool) ParamSpec {
	return ParamSpec{
		Name:       name,
		Type:       TypeNone,
		Optional:   true,
		Refinement: true,
		TakesValue: takesValue,
		Eval:       true,
	}
}

type FunctionValue struct {
	Type    FunctionType
	Name    string
	Params  []ParamSpec
	Body    *BlockValue
	Native  core.NativeFunc
	Parent  int
	Infix   bool
	NoScope bool
	Doc     *docmodel.FuncDoc
}

func NewNativeFunction(name string, params []ParamSpec, impl core.NativeFunc, infix bool, doc *docmodel.FuncDoc) *FunctionValue {
	return &FunctionValue{
		Type:    FuncNative,
		Name:    name,
		Params:  params,
		Body:    nil,
		Native:  impl,
		Parent:  -1,
		Infix:   infix,
		NoScope: false,
		Doc:     doc,
	}
}

func NewUserFunction(name string, params []ParamSpec, body *BlockValue, parentFrame int, noScope bool, doc *docmodel.FuncDoc) *FunctionValue {
	return &FunctionValue{
		Type:    FuncUser,
		Name:    name,
		Params:  params,
		Body:    body,
		Native:  nil,
		Parent:  parentFrame,
		NoScope: noScope,
		Doc:     doc,
	}
}

func (f *FunctionValue) String() string {
	return f.Mold()
}

func (f *FunctionValue) Mold() string {
	if f.Type == FuncNative {
		return fmt.Sprintf("native[%s]", f.Name)
	}
	return fmt.Sprintf("function[%s]", f.Name)
}

func (f *FunctionValue) Form() string {
	return f.Mold()
}

func (f *FunctionValue) Arity() int {
	count := 0
	for _, p := range f.Params {
		if !p.Refinement && !p.Optional {
			count++
		}
	}
	return count
}

func (f *FunctionValue) HasRefinement(name string) bool {
	for _, p := range f.Params {
		if p.Refinement && p.Name == name {
			return true
		}
	}
	return false
}

func (f *FunctionValue) GetRefinement(name string) *ParamSpec {
	for i := range f.Params {
		if f.Params[i].Refinement && f.Params[i].Name == name {
			return &f.Params[i]
		}
	}
	return nil
}

func (f *FunctionValue) Equals(other core.Value) bool {
	if other.GetType() == TypeFunction {
		return other.GetPayload() == f
	}
	return false
}

func (f *FunctionValue) GetType() core.ValueType {
	return TypeFunction
}

func (f *FunctionValue) GetPayload() any {
	return f
}
