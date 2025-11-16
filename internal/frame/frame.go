package frame

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"slices"
)

const (
	FrameFunctionArgs core.FrameType = iota
	FrameClosure
	FrameObject
	FrameTypeFrame
)

type Frame struct {
	Type   core.FrameType
	Words  []string
	Values []core.Value
	Parent int
	Index  int
	Name   string
}

func NewFrame(frameType core.FrameType, parent int) core.Frame {
	return &Frame{
		Type:   frameType,
		Words:  []string{},
		Values: []core.Value{},
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

func NewFrameWithCapacity(frameType core.FrameType, parent int, capacity int) *Frame {
	return &Frame{
		Type:   frameType,
		Words:  make([]string, 0, capacity),
		Values: make([]core.Value, 0, capacity),
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

func NewObjectFrame(parent int, words []string, types []core.ValueType) *Frame {
	return &Frame{
		Type:   FrameObject,
		Words:  make([]string, 0, len(words)),
		Values: make([]core.Value, 0, len(words)),
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

func (f *Frame) GetType() core.FrameType {
	return f.Type
}

func (f *Frame) ChangeType(newType core.FrameType) {
	f.Type = newType
}

func (f *Frame) GetParent() int {
	return f.Parent
}

func (f *Frame) GetIndex() int {
	return f.Index
}

func (f *Frame) SetIndex(idx int) {
	f.Index = idx
}

func (f *Frame) GetName() string {
	return f.Name
}

func (f *Frame) SetName(name string) {
	f.Name = name
}

func (f *Frame) Bind(symbol string, val core.Value) {
	for i, w := range f.Words {
		if w == symbol {
			f.Values[i] = val
			return
		}
	}

	f.Words = append(f.Words, symbol)
	f.Values = append(f.Values, val)
}

func (f *Frame) Get(symbol string) (core.Value, bool) {
	for i, w := range f.Words {
		if w == symbol {
			return f.Values[i], true
		}
	}
	return value.NewNoneVal(), false
}

func (f *Frame) Set(symbol string, val core.Value) bool {
	for i, w := range f.Words {
		if w == symbol {
			f.Values[i] = val
			return true
		}
	}
	return false
}

func (f *Frame) HasWord(symbol string) bool {
	return slices.Contains(f.Words, symbol)
}

func (f *Frame) Unbind(symbol string) bool {
	for i, w := range f.Words {
		if w == symbol {
			f.Words = append(f.Words[:i], f.Words[i+1:]...)
			f.Values = append(f.Values[:i], f.Values[i+1:]...)
			return true
		}
	}
	return false
}

func (f *Frame) Count() int {
	return len(f.Words)
}

func (f *Frame) GetAll() []core.Binding {
	bindings := make([]core.Binding, len(f.Words))
	for i := range f.Words {
		bindings[i] = core.Binding{
			Symbol: f.Words[i],
			Value:  f.Values[i],
		}
	}
	return bindings
}

func (f *Frame) Clone() core.Frame {
	wordsCopy := make([]string, len(f.Words))
	valuesCopy := make([]core.Value, len(f.Values))
	copy(wordsCopy, f.Words)
	copy(valuesCopy, f.Values)

	return &Frame{
		Type:   f.Type,
		Words:  wordsCopy,
		Values: valuesCopy,
		Parent: f.Parent,
		Index:  -1,
	}
}

type SharedFrame struct {
	callerFrame   core.Frame
	lexicalParent int
}

func NewSharedFrame(callerFrame core.Frame, lexicalParent int) *SharedFrame {
	return &SharedFrame{
		callerFrame:   callerFrame,
		lexicalParent: lexicalParent,
	}
}

func (s *SharedFrame) GetType() core.FrameType {
	return s.callerFrame.GetType()
}

func (s *SharedFrame) ChangeType(newType core.FrameType) {
	s.callerFrame.ChangeType(newType)
}

func (s *SharedFrame) Bind(symbol string, value core.Value) {
	s.callerFrame.Bind(symbol, value)
}

func (s *SharedFrame) Get(symbol string) (core.Value, bool) {
	return s.callerFrame.Get(symbol)
}

func (s *SharedFrame) Set(symbol string, value core.Value) bool {
	return s.callerFrame.Set(symbol, value)
}

func (s *SharedFrame) HasWord(symbol string) bool {
	return s.callerFrame.HasWord(symbol)
}

func (s *SharedFrame) Unbind(symbol string) bool {
	return s.callerFrame.Unbind(symbol)
}

func (s *SharedFrame) GetParent() int {
	return s.lexicalParent
}

func (s *SharedFrame) GetIndex() int {
	return s.callerFrame.GetIndex()
}

func (s *SharedFrame) SetIndex(index int) {
	s.callerFrame.SetIndex(index)
}

func (s *SharedFrame) Count() int {
	return s.callerFrame.Count()
}

func (s *SharedFrame) GetAll() []core.Binding {
	return s.callerFrame.GetAll()
}

func (s *SharedFrame) Clone() core.Frame {
	return s.callerFrame.Clone()
}

func (s *SharedFrame) GetName() string {
	return s.callerFrame.GetName()
}

func (s *SharedFrame) SetName(name string) {
	s.callerFrame.SetName(name)
}
