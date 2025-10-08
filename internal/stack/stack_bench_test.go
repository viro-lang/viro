package stack_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/stack"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func BenchmarkStackPushPop(b *testing.B) {
	st := stack.NewStack(1024)
	v := value.IntVal(42)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		st.Push(v)
		_ = st.Pop()
	}
}

func BenchmarkStackGetSet(b *testing.B) {
	st := stack.NewStack(1024)
	idx := st.Push(value.IntVal(1))
	v := value.IntVal(99)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		st.Set(idx, v)
		_ = st.Get(idx)
	}
}

func BenchmarkStackReserve(b *testing.B) {
	st := stack.NewStack(16)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		st.Reserve(64)
	}
}
