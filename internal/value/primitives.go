package value

import (
	"strconv"

	"github.com/marcin-radoszewski/viro/internal/core"
)

var (
	noneValSingleton  = NoneValue{}
	trueValSingleton  = LogicValue(true)
	falseValSingleton = LogicValue(false)
)

type IntValue int64

func (i IntValue) GetType() core.ValueType {
	return TypeInteger
}

func (i IntValue) GetPayload() any {
	return int64(i)
}

func (i IntValue) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i IntValue) Mold() string {
	return i.String()
}

func (i IntValue) Form() string {
	return i.String()
}

func (i IntValue) Equals(other core.Value) bool {
	if oi, ok := other.(IntValue); ok {
		return i == oi
	}
	return false
}

type LogicValue bool

func (l LogicValue) GetType() core.ValueType {
	return TypeLogic
}

func (l LogicValue) GetPayload() any {
	return bool(l)
}

func (l LogicValue) String() string {
	if l {
		return "true"
	}
	return "false"
}

func (l LogicValue) Mold() string {
	return l.String()
}

func (l LogicValue) Form() string {
	return l.String()
}

func (l LogicValue) Equals(other core.Value) bool {
	if ol, ok := other.(LogicValue); ok {
		return l == ol
	}
	return false
}

type NoneValue struct{}

func (n NoneValue) GetType() core.ValueType {
	return TypeNone
}

func (n NoneValue) GetPayload() any {
	return nil
}

func (n NoneValue) String() string {
	return "none"
}

func (n NoneValue) Mold() string {
	return "none"
}

func (n NoneValue) Form() string {
	return "none"
}

func (n NoneValue) Equals(other core.Value) bool {
	_, ok := other.(NoneValue)
	return ok
}

type WordValue string

func (w WordValue) GetType() core.ValueType {
	return TypeWord
}

func (w WordValue) GetPayload() any {
	return string(w)
}

func (w WordValue) String() string {
	return string(w)
}

func (w WordValue) Mold() string {
	return string(w)
}

func (w WordValue) Form() string {
	return string(w)
}

func (w WordValue) Equals(other core.Value) bool {
	if ow, ok := other.(WordValue); ok {
		return w == ow
	}
	return false
}

type SetWordValue string

func (s SetWordValue) GetType() core.ValueType {
	return TypeSetWord
}

func (s SetWordValue) GetPayload() any {
	return string(s)
}

func (s SetWordValue) String() string {
	return string(s) + ":"
}

func (s SetWordValue) Mold() string {
	return string(s) + ":"
}

func (s SetWordValue) Form() string {
	return string(s) + ":"
}

func (s SetWordValue) Equals(other core.Value) bool {
	if os, ok := other.(SetWordValue); ok {
		return s == os
	}
	return false
}

type GetWordValue string

func (g GetWordValue) GetType() core.ValueType {
	return TypeGetWord
}

func (g GetWordValue) GetPayload() any {
	return string(g)
}

func (g GetWordValue) String() string {
	return ":" + string(g)
}

func (g GetWordValue) Mold() string {
	return ":" + string(g)
}

func (g GetWordValue) Form() string {
	return ":" + string(g)
}

func (g GetWordValue) Equals(other core.Value) bool {
	if og, ok := other.(GetWordValue); ok {
		return g == og
	}
	return false
}

type LitWordValue string

func (l LitWordValue) GetType() core.ValueType {
	return TypeLitWord
}

func (l LitWordValue) GetPayload() any {
	return string(l)
}

func (l LitWordValue) String() string {
	return "'" + string(l)
}

func (l LitWordValue) Mold() string {
	return "'" + string(l)
}

func (l LitWordValue) Form() string {
	return "'" + string(l)
}

func (l LitWordValue) Equals(other core.Value) bool {
	if ol, ok := other.(LitWordValue); ok {
		return l == ol
	}
	return false
}

type DatatypeValue string

func (d DatatypeValue) GetType() core.ValueType {
	return TypeDatatype
}

func (d DatatypeValue) GetPayload() any {
	return string(d)
}

func (d DatatypeValue) String() string {
	return string(d)
}

func (d DatatypeValue) Mold() string {
	return string(d)
}

func (d DatatypeValue) Form() string {
	return string(d)
}

func (d DatatypeValue) Equals(other core.Value) bool {
	if od, ok := other.(DatatypeValue); ok {
		return d == od
	}
	return false
}

func NewIntVal(i int64) core.Value {
	return IntValue(i)
}

func NewLogicVal(b bool) core.Value {
	if b {
		return trueValSingleton
	}
	return falseValSingleton
}

func NewNoneVal() core.Value {
	return noneValSingleton
}

func NewWordVal(symbol string) core.Value {
	return WordValue(symbol)
}

func NewSetWordVal(symbol string) core.Value {
	return SetWordValue(symbol)
}

func NewGetWordVal(symbol string) core.Value {
	return GetWordValue(symbol)
}

func NewLitWordVal(symbol string) core.Value {
	return LitWordValue(symbol)
}

func NewDatatypeVal(name string) core.Value {
	return DatatypeValue(name)
}

func AsIntValue(v core.Value) (int64, bool) {
	if iv, ok := v.(IntValue); ok {
		return int64(iv), true
	}
	return 0, false
}

func AsLogicValue(v core.Value) (bool, bool) {
	if lv, ok := v.(LogicValue); ok {
		return bool(lv), true
	}
	return false, false
}

func AsWordValue(v core.Value) (string, bool) {
	switch wv := v.(type) {
	case WordValue:
		return string(wv), true
	case SetWordValue:
		return string(wv), true
	case GetWordValue:
		return string(wv), true
	case LitWordValue:
		return string(wv), true
	default:
		return "", false
	}
}

func AsDatatypeValue(v core.Value) (string, bool) {
	if dv, ok := v.(DatatypeValue); ok {
		return string(dv), true
	}
	return "", false
}

func GetNoneVal() core.Value {
	return noneValSingleton
}

func NewStrVal(s string) core.Value {
	return NewStringValue(s)
}

func NewBlockVal(elements []core.Value) core.Value {
	return NewBlockValue(elements)
}

func NewParenVal(elements []core.Value) core.Value {
	return NewBlockValueWithType(elements, TypeParen)
}

func NewFuncVal(fn *FunctionValue) core.Value {
	return fn
}

func NewBinaryVal(data []byte) core.Value {
	return NewBinaryValue(data)
}

func AsStringValue(v core.Value) (*StringValue, bool) {
	if v.GetType() != TypeString {
		return nil, false
	}
	if sv, ok := v.(*StringValue); ok {
		return sv, true
	}
	return nil, false
}

func AsBlockValue(v core.Value) (*BlockValue, bool) {
	if v.GetType() != TypeBlock && v.GetType() != TypeParen {
		return nil, false
	}
	if bv, ok := v.(*BlockValue); ok {
		return bv, true
	}
	return nil, false
}

func AsFunctionValue(v core.Value) (*FunctionValue, bool) {
	if v.GetType() != TypeFunction {
		return nil, false
	}
	if fv, ok := v.(*FunctionValue); ok {
		return fv, true
	}
	return nil, false
}

func AsBinaryValue(v core.Value) (*BinaryValue, bool) {
	if v.GetType() != TypeBinary {
		return nil, false
	}
	if bv, ok := v.(*BinaryValue); ok {
		return bv, true
	}
	return nil, false
}

// DeepCloneValue recursively clones values that contain mutable state.
// Currently handles blocks and parens to prevent state sharing between
// function invocations. Other value types are returned as-is since they
// are either immutable or properly isolated.
func DeepCloneValue(val core.Value) core.Value {
	switch val.GetType() {
	case TypeBlock, TypeParen:
		block, _ := AsBlockValue(val)
		clonedElements := make([]core.Value, len(block.Elements))
		for i, elem := range block.Elements {
			clonedElements[i] = DeepCloneValue(elem)
		}
		if val.GetType() == TypeBlock {
			return NewBlockVal(clonedElements)
		} else {
			return NewParenVal(clonedElements)
		}
	default:
		return val
	}
}
