package stack_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func BenchmarkStackPushPop(b *testing.B) {
	st := stack.NewStack(1024)
	v := value.NewIntVal(42)
	b.ReportAllocs()

	for b.Loop() {
		st.Push(v)
		_ = st.Pop()
	}
}

func BenchmarkStackGetSet(b *testing.B) {
	st := stack.NewStack(1024)
	idx := st.Push(value.NewIntVal(1))
	v := value.NewIntVal(99)
	b.ReportAllocs()

	for b.Loop() {
		st.Set(idx, v)
		_ = st.Get(idx)
	}
}

func BenchmarkStackReserve(b *testing.B) {
	st := stack.NewStack(16)
	b.ReportAllocs()

	for b.Loop() {
		st.Reserve(64)
	}
}
