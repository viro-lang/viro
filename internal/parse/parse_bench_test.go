package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/tokenize"
)

func BenchmarkTokenizeSimple(b *testing.B) {
	input := "x: 42"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		t := tokenize.NewTokenizer(input)
		_, err := t.Tokenize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenizeMedium(b *testing.B) {
	input := "sum: fn [a b] [a + b]\nresult: sum 10 20"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		t := tokenize.NewTokenizer(input)
		_, err := t.Tokenize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenizeComplex(b *testing.B) {
	input := `
fib: fn [n] [
	if n <= 1 [
		n
	] [
		fib (n - 1) + fib (n - 2)
	]
]
result: fib 10
`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		t := tokenize.NewTokenizer(input)
		_, err := t.Tokenize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenizeLongString(b *testing.B) {
	input := `msg: "This is a reasonably long string that might appear in real code, with punctuation, numbers like 123, and special chars!"`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		t := tokenize.NewTokenizer(input)
		_, err := t.Tokenize()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseSimple(b *testing.B) {
	input := "x: 42"
	t := tokenize.NewTokenizer(input)
	tokens, err := t.Tokenize()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := NewParser(tokens)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseMedium(b *testing.B) {
	input := "sum: fn [a b] [a + b]\nresult: sum 10 20"
	t := tokenize.NewTokenizer(input)
	tokens, err := t.Tokenize()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := NewParser(tokens)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseComplex(b *testing.B) {
	input := `
fib: fn [n] [
	if n <= 1 [
		n
	] [
		fib (n - 1) + fib (n - 2)
	]
]
result: fib 10
`
	t := tokenize.NewTokenizer(input)
	tokens, err := t.Tokenize()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := NewParser(tokens)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseBlock(b *testing.B) {
	input := "[1 2 3 4 5 6 7 8 9 10]"
	t := tokenize.NewTokenizer(input)
	tokens, err := t.Tokenize()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := NewParser(tokens)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParsePath(b *testing.B) {
	input := "user.profile.address.city"
	t := tokenize.NewTokenizer(input)
	tokens, err := t.Tokenize()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := NewParser(tokens)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullSimple(b *testing.B) {
	input := "x: 42"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullMedium(b *testing.B) {
	input := "sum: fn [a b] [a + b]\nresult: sum 10 20"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullComplex(b *testing.B) {
	input := `
fib: fn [n] [
	if n <= 1 [
		n
	] [
		fib (n - 1) + fib (n - 2)
	]
]
result: fib 10
`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullMathExpression(b *testing.B) {
	input := "1 + 2 * 3 - 4 / 2 + 5 * (6 - 7)"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullDataTypes(b *testing.B) {
	input := `x: 42  y: 3.14  z: "hello"  flag: true  data: [1 2 3]  path: obj.field`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullNestedBlocks(b *testing.B) {
	input := `[[1 2] [3 4] [5 [6 7 [8 9]]]]`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFullRealWorldScript(b *testing.B) {
	input := `
; Calculate factorial
factorial: fn [n] [
	if n <= 1 [
		1
	] [
		n * factorial (n - 1)
	]
]

; Calculate sum of factorials
sum: 0
loop 5 [i] [
	sum: sum + factorial i
]

; Print result
print ["Total:" sum]
`

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := Parse(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}
