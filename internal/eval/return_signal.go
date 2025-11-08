package eval

import "github.com/marcin-radoszewski/viro/internal/core"

type ReturnSignal struct {
	value core.Value
}

func NewReturnSignal(val core.Value) *ReturnSignal {
	return &ReturnSignal{value: val}
}

func (r *ReturnSignal) Error() string {
	return "return signal"
}

func (r *ReturnSignal) Value() core.Value {
	return r.value
}
