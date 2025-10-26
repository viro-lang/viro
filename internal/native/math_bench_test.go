package native_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

var benchResult core.Value

func benchmarkMathOp(b *testing.B, fn core.NativeFunc, args []core.Value) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := fn(args, nil, nil)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		benchResult = result
	}
}

func BenchmarkMathAdd(b *testing.B) {
	args := []core.Value{value.NewIntVal(123456789), value.NewIntVal(987654321)}
	benchmarkMathOp(b, native.Add, args)
}

func BenchmarkMathSubtract(b *testing.B) {
	args := []core.Value{value.NewIntVal(987654321), value.NewIntVal(123456789)}
	benchmarkMathOp(b, native.Subtract, args)
}

func BenchmarkMathMultiply(b *testing.B) {
	args := []core.Value{value.NewIntVal(12345), value.NewIntVal(6789)}
	benchmarkMathOp(b, native.Multiply, args)
}

func BenchmarkMathDivide(b *testing.B) {
	args := []core.Value{value.NewIntVal(9876543210), value.NewIntVal(12345)}
	benchmarkMathOp(b, native.Divide, args)
}

// Feature 002: Decimal arithmetic benchmarks
// Target per tasks.md: decimal operations ≤1.5× integer baseline

func BenchmarkDecimalAdd(b *testing.B) {
	d1, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("123456789.123456789")}, nil, nil)
	d2, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("987654321.987654321")}, nil, nil)
	args := []core.Value{d1, d2}
	benchmarkMathOp(b, native.Add, args)
}

func BenchmarkDecimalSubtract(b *testing.B) {
	d1, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("987654321.987654321")}, nil, nil)
	d2, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("123456789.123456789")}, nil, nil)
	args := []core.Value{d1, d2}
	benchmarkMathOp(b, native.Subtract, args)
}

func BenchmarkDecimalMultiply(b *testing.B) {
	d1, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("12345.6789")}, nil, nil)
	d2, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("6789.12345")}, nil, nil)
	args := []core.Value{d1, d2}
	benchmarkMathOp(b, native.Multiply, args)
}

func BenchmarkDecimalDivide(b *testing.B) {
	d1, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("9876543210.123")}, nil, nil)
	d2, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("12345.6789")}, nil, nil)
	args := []core.Value{d1, d2}
	benchmarkMathOp(b, native.Divide, args)
}

func BenchmarkDecimalSqrt(b *testing.B) {
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("123456.789")}, nil, nil)
	args := []core.Value{d}
	benchmarkMathOp(b, native.Sqrt, args)
}

func BenchmarkDecimalPow(b *testing.B) {
	base, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("2.5")}, nil, nil)
	exp, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("3.2")}, nil, nil)
	args := []core.Value{base, exp}
	benchmarkMathOp(b, native.Pow, args)
}

func BenchmarkDecimalExp(b *testing.B) {
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("2.5")}, nil, nil)
	args := []core.Value{d}
	benchmarkMathOp(b, native.Exp, args)
}

func BenchmarkDecimalLog(b *testing.B) {
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("12345.6789")}, nil, nil)
	args := []core.Value{d}
	benchmarkMathOp(b, native.Log, args)
}

func BenchmarkDecimalSin(b *testing.B) {
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("1.5707963267948966")}, nil, nil)
	args := []core.Value{d}
	benchmarkMathOp(b, native.Sin, args)
}

func BenchmarkDecimalRound(b *testing.B) {
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("123.456789")}, nil, nil)
	places := value.NewIntVal(2)
	args := []core.Value{d, places}
	benchmarkMathOp(b, native.Round, args)
}

// Mixed integer-decimal promotion benchmark
func BenchmarkMixedIntegerDecimalMultiply(b *testing.B) {
	i := value.NewIntVal(12345)
	d, _ := native.DecimalConstructor([]core.Value{value.NewStrVal("6789.12345")}, nil, nil)
	args := []core.Value{i, d}
	benchmarkMathOp(b, native.Multiply, args)
}
