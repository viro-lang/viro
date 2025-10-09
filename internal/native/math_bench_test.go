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

// Feature 002: Decimal arithmetic benchmarks
// Target per tasks.md: decimal operations ≤1.5× integer baseline

func BenchmarkDecimalAdd(b *testing.B) {
	d1, _ := native.DecimalConstructor([]value.Value{value.StrVal("123456789.123456789")})
	d2, _ := native.DecimalConstructor([]value.Value{value.StrVal("987654321.987654321")})
	args := []value.Value{d1, d2}
	benchmarkMathOp(b, native.Add, args)
}

func BenchmarkDecimalSubtract(b *testing.B) {
	d1, _ := native.DecimalConstructor([]value.Value{value.StrVal("987654321.987654321")})
	d2, _ := native.DecimalConstructor([]value.Value{value.StrVal("123456789.123456789")})
	args := []value.Value{d1, d2}
	benchmarkMathOp(b, native.Subtract, args)
}

func BenchmarkDecimalMultiply(b *testing.B) {
	d1, _ := native.DecimalConstructor([]value.Value{value.StrVal("12345.6789")})
	d2, _ := native.DecimalConstructor([]value.Value{value.StrVal("6789.12345")})
	args := []value.Value{d1, d2}
	benchmarkMathOp(b, native.Multiply, args)
}

func BenchmarkDecimalDivide(b *testing.B) {
	d1, _ := native.DecimalConstructor([]value.Value{value.StrVal("9876543210.123")})
	d2, _ := native.DecimalConstructor([]value.Value{value.StrVal("12345.6789")})
	args := []value.Value{d1, d2}
	benchmarkMathOp(b, native.Divide, args)
}

func BenchmarkDecimalSqrt(b *testing.B) {
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("123456.789")})
	args := []value.Value{d}
	benchmarkMathOp(b, native.Sqrt, args)
}

func BenchmarkDecimalPow(b *testing.B) {
	base, _ := native.DecimalConstructor([]value.Value{value.StrVal("2.5")})
	exp, _ := native.DecimalConstructor([]value.Value{value.StrVal("3.2")})
	args := []value.Value{base, exp}
	benchmarkMathOp(b, native.Pow, args)
}

func BenchmarkDecimalExp(b *testing.B) {
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("2.5")})
	args := []value.Value{d}
	benchmarkMathOp(b, native.Exp, args)
}

func BenchmarkDecimalLog(b *testing.B) {
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("12345.6789")})
	args := []value.Value{d}
	benchmarkMathOp(b, native.Log, args)
}

func BenchmarkDecimalSin(b *testing.B) {
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("1.5707963267948966")}) // π/2
	args := []value.Value{d}
	benchmarkMathOp(b, native.Sin, args)
}

func BenchmarkDecimalRound(b *testing.B) {
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("123.456789")})
	places := value.IntVal(2)
	args := []value.Value{d, places}
	benchmarkMathOp(b, native.Round, args)
}

// Mixed integer-decimal promotion benchmark
func BenchmarkMixedIntegerDecimalMultiply(b *testing.B) {
	i := value.IntVal(12345)
	d, _ := native.DecimalConstructor([]value.Value{value.StrVal("6789.12345")})
	args := []value.Value{i, d}
	benchmarkMathOp(b, native.Multiply, args)
}

