package core

import "io"

type ValueType uint8

type NativeFunc func(args []Value, refValues map[string]Value, eval Evaluator) (Value, error)

type Value interface {
	GetType() ValueType
	GetPayload() any
	String() string
	Equals(other Value) bool
}

type Binding struct {
	Symbol string
	Value  Value
}

type FrameType uint8

type Frame interface {
	GetType() FrameType
	ChangeType(newType FrameType)
	Bind(symbol string, value Value)
	Get(symbol string) (Value, bool)
	Set(symbol string, value Value) bool
	HasWord(symbol string) bool
	GetParent() int
	GetIndex() int
	SetIndex(int)
	Count() int
	GetAll() []Binding
	Clone() Frame
	ValidateFieldType(symbol string, value Value) bool
	HasManifestField(symbol string) bool
	GetName() string
	SetName(name string)
}

type Evaluator interface {
	CurrentFrameIndex() int
	RegisterFrame(frame Frame) int
	MarkFrameCaptured(idx int)
	GetFrameByIndex(idx int) Frame
	PushFrameContext(frame Frame) int
	PopFrameContext()
	Lookup(symbol string) (Value, bool)
	DoNext(value Value) (Value, error)
	DoBlock(vals []Value) (Value, error)
	Callstack() []string
	SetOutputWriter(writer io.Writer)
	GetOutputWriter() io.Writer
	SetErrorWriter(writer io.Writer)
	GetErrorWriter() io.Writer
	SetInputReader(reader io.Reader)
	GetInputReader() io.Reader
}
