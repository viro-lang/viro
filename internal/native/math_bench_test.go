package native_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

var benchResult value.Value

func benchmarkMathOp(b *testing.B, fn func([]value.Value) (value.Value, *verror.Error), args []value.Value) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := fn(args)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		benchResult = result
	}
}

func BenchmarkMathAdd(b *testing.B) {
	args := []value.Value{value.IntVal(123456789), value.IntVal(987654321)}
	benchmarkMathOp(b, native.Add, args)
}

func BenchmarkMathSubtract(b *testing.B) {
	args := []value.Value{value.IntVal(987654321), value.IntVal(123456789)}
	benchmarkMathOp(b, native.Subtract, args)
}

func BenchmarkMathMultiply(b *testing.B) {
	args := []value.Value{value.IntVal(12345), value.IntVal(6789)}
	benchmarkMathOp(b, native.Multiply, args)
}

func BenchmarkMathDivide(b *testing.B) {
	args := []value.Value{value.IntVal(9876543210), value.IntVal(12345)}
	benchmarkMathOp(b, native.Divide, args)
}
