package value

import (
	"strconv"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type IntValue struct {
	baseValue
	value int64
}

func (i *IntValue) GetType() core.ValueType {
	return TypeInteger
}

func (i *IntValue) GetPayload() any {
	return i.value
}

func (i *IntValue) String() string {
	return strconv.FormatInt(i.value, 10)
}

func (i *IntValue) Mold() string {
	return i.String()
}

func (i *IntValue) Form() string {
	return i.String()
}

func (i *IntValue) Equals(other core.Value) bool {
	if oi, ok := other.(*IntValue); ok {
		return i.value == oi.value
	}
	return false
}

type LogicValue struct {
	baseValue
	value bool
}

func (l *LogicValue) GetType() core.ValueType {
	return TypeLogic
}

func (l *LogicValue) GetPayload() any {
	return l.value
}

func (l *LogicValue) String() string {
	if l.value {
		return "true"
	}
	return "false"
}

func (l *LogicValue) Mold() string {
	return l.String()
}

func (l *LogicValue) Form() string {
	return l.String()
}

func (l *LogicValue) Equals(other core.Value) bool {
	if ol, ok := other.(*LogicValue); ok {
		return l.value == ol.value
	}
	return false
}

type NoneValue struct {
	baseValue
}

func (n *NoneValue) GetType() core.ValueType {
	return TypeNone
}

func (n *NoneValue) GetPayload() any {
	return nil
}

func (n *NoneValue) String() string {
	return "none"
}

func (n *NoneValue) Mold() string {
	return "none"
}

func (n *NoneValue) Form() string {
	return "none"
}

func (n *NoneValue) Equals(other core.Value) bool {
	_, ok := other.(*NoneValue)
	return ok
}

type WordValue struct {
	baseValue
	symbol string
}

func (w *WordValue) GetType() core.ValueType {
	return TypeWord
}

func (w *WordValue) GetPayload() any {
	return w.symbol
}

func (w *WordValue) String() string {
	return w.symbol
}

func (w *WordValue) Mold() string {
	return w.symbol
}

func (w *WordValue) Form() string {
	return w.symbol
}

func (w *WordValue) Equals(other core.Value) bool {
	if ow, ok := other.(*WordValue); ok {
		return w.symbol == ow.symbol
	}
	return false
}

type SetWordValue struct {
	baseValue
	symbol string
}

func (s *SetWordValue) GetType() core.ValueType {
	return TypeSetWord
}

func (s *SetWordValue) GetPayload() any {
	return s.symbol
}

func (s *SetWordValue) String() string {
	return s.symbol + ":"
}

func (s *SetWordValue) Mold() string {
	return s.String()
}

func (s *SetWordValue) Form() string {
	return s.String()
}

func (s *SetWordValue) Equals(other core.Value) bool {
	if os, ok := other.(*SetWordValue); ok {
		return s.symbol == os.symbol
	}
	return false
}

type GetWordValue struct {
	baseValue
	symbol string
}

func (g *GetWordValue) GetType() core.ValueType {
	return TypeGetWord
}

func (g *GetWordValue) GetPayload() any {
	return g.symbol
}

func (g *GetWordValue) String() string {
	return ":" + g.symbol
}

func (g *GetWordValue) Mold() string {
	return g.String()
}

func (g *GetWordValue) Form() string {
	return g.String()
}

func (g *GetWordValue) Equals(other core.Value) bool {
	if og, ok := other.(*GetWordValue); ok {
		return g.symbol == og.symbol
	}
	return false
}

type LitWordValue struct {
	baseValue
	symbol string
}

func (l *LitWordValue) GetType() core.ValueType {
	return TypeLitWord
}

func (l *LitWordValue) GetPayload() any {
	return l.symbol
}

func (l *LitWordValue) String() string {
	return "'" + l.symbol
}

func (l *LitWordValue) Mold() string {
	return l.String()
}

func (l *LitWordValue) Form() string {
	return l.String()
}

func (l *LitWordValue) Equals(other core.Value) bool {
	if ol, ok := other.(*LitWordValue); ok {
		return l.symbol == ol.symbol
	}
	return false
}

type DatatypeValue struct {
	baseValue
	name string
}

func (d *DatatypeValue) GetType() core.ValueType {
	return TypeDatatype
}

func (d *DatatypeValue) GetPayload() any {
	return d.name
}

func (d *DatatypeValue) String() string {
	return d.name
}

func (d *DatatypeValue) Mold() string {
	return d.String()
}

func (d *DatatypeValue) Form() string {
	return d.String()
}

func (d *DatatypeValue) Equals(other core.Value) bool {
	if od, ok := other.(*DatatypeValue); ok {
		return d.name == od.name
	}
	return false
}

func NewIntVal(i int64) core.Value {
	return &IntValue{value: i}
}

func NewLogicVal(b bool) core.Value {
	return &LogicValue{value: b}
}

func NewNoneVal() core.Value {
	return &NoneValue{}
}

func NewWordVal(symbol string) core.Value {
	return &WordValue{symbol: symbol}
}

func NewSetWordVal(symbol string) core.Value {
	return &SetWordValue{symbol: symbol}
}

func NewGetWordVal(symbol string) core.Value {
	return &GetWordValue{symbol: symbol}
}

func NewLitWordVal(symbol string) core.Value {
	return &LitWordValue{symbol: symbol}
}

func NewDatatypeVal(name string) core.Value {
	return &DatatypeValue{name: name}
}

func AsIntValue(v core.Value) (int64, bool) {
	if iv, ok := v.(*IntValue); ok {
		return iv.value, true
	}
	return 0, false
}

func AsLogicValue(v core.Value) (bool, bool) {
	if lv, ok := v.(*LogicValue); ok {
		return lv.value, true
	}
	return false, false
}

func AsWordValue(v core.Value) (string, bool) {
	switch wv := v.(type) {
	case *WordValue:
		return wv.symbol, true
	case *SetWordValue:
		return wv.symbol, true
	case *GetWordValue:
		return wv.symbol, true
	case *LitWordValue:
		return wv.symbol, true
	default:
		return "", false
	}
}

func AsDatatypeValue(v core.Value) (string, bool) {
	if dv, ok := v.(*DatatypeValue); ok {
		return dv.name, true
	}
	return "", false
}

func GetNoneVal() core.Value {
	return NewNoneVal()
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
