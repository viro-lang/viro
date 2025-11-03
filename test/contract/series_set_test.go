package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestBlockIntersect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple intersection",
			input:    "intersect [1 2 3] [2 3 4]",
			expected: "[2 3]",
		},
		{
			name:     "no common elements",
			input:    "intersect [1 2 3] [4 5 6]",
			expected: "[]",
		},
		{
			name:     "all common elements",
			input:    "intersect [1 2 3] [1 2 3]",
			expected: "[1 2 3]",
		},
		{
			name:     "duplicates in first series",
			input:    "intersect [1 2 2 3] [2 3 4]",
			expected: "[2 3]",
		},
		{
			name:     "duplicates in second series",
			input:    "intersect [1 2 3] [2 2 3 4]",
			expected: "[2 3]",
		},
		{
			name:     "empty first series",
			input:    "intersect [] [1 2 3]",
			expected: "[]",
		},
		{
			name:     "empty second series",
			input:    "intersect [1 2 3] []",
			expected: "[]",
		},
		{
			name:     "both empty",
			input:    "intersect [] []",
			expected: "[]",
		},
		{
			name:     "mixed types",
			input:    `intersect [1 "a" 2] [2 "a" 3]`,
			expected: `["a" 2]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestBlockDifference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple difference",
			input:    "difference [1 2 3] [2 3 4]",
			expected: "[1]",
		},
		{
			name:     "no common elements",
			input:    "difference [1 2 3] [4 5 6]",
			expected: "[1 2 3]",
		},
		{
			name:     "all common elements",
			input:    "difference [1 2 3] [1 2 3]",
			expected: "[]",
		},
		{
			name:     "duplicates in first series",
			input:    "difference [1 1 2 3] [2 3 4]",
			expected: "[1]",
		},
		{
			name:     "empty first series",
			input:    "difference [] [1 2 3]",
			expected: "[]",
		},
		{
			name:     "empty second series",
			input:    "difference [1 2 3] []",
			expected: "[1 2 3]",
		},
		{
			name:     "both empty",
			input:    "difference [] []",
			expected: "[]",
		},
		{
			name:     "mixed types",
			input:    `difference [1 "a" 2] [2 "a" 3]`,
			expected: "[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestBlockUnion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple union",
			input:    "union [1 2 3] [2 3 4]",
			expected: "[1 2 3 4]",
		},
		{
			name:     "no common elements",
			input:    "union [1 2 3] [4 5 6]",
			expected: "[1 2 3 4 5 6]",
		},
		{
			name:     "all common elements",
			input:    "union [1 2 3] [1 2 3]",
			expected: "[1 2 3]",
		},
		{
			name:     "duplicates in first series",
			input:    "union [1 1 2 3] [4 5]",
			expected: "[1 2 3 4 5]",
		},
		{
			name:     "duplicates in second series",
			input:    "union [1 2] [3 3 4 5]",
			expected: "[1 2 3 4 5]",
		},
		{
			name:     "duplicates in both",
			input:    "union [1 1 2 3] [2 3 3 4]",
			expected: "[1 2 3 4]",
		},
		{
			name:     "empty first series",
			input:    "union [] [1 2 3]",
			expected: "[1 2 3]",
		},
		{
			name:     "empty second series",
			input:    "union [1 2 3] []",
			expected: "[1 2 3]",
		},
		{
			name:     "both empty",
			input:    "union [] []",
			expected: "[]",
		},
		{
			name:     "mixed types",
			input:    `union [1 "a" 2] [2 "b" 3]`,
			expected: `[1 "a" 2 "b" 3]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestStringIntersect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple intersection",
			input:    `intersect "hello" "world"`,
			expected: `"lo"`,
		},
		{
			name:     "no common characters",
			input:    `intersect "abc" "def"`,
			expected: `""`,
		},
		{
			name:     "all common characters",
			input:    `intersect "abc" "abc"`,
			expected: `"abc"`,
		},
		{
			name:     "duplicates removed",
			input:    `intersect "aabbcc" "abc"`,
			expected: `"abc"`,
		},
		{
			name:     "empty first string",
			input:    `intersect "" "abc"`,
			expected: `""`,
		},
		{
			name:     "empty second string",
			input:    `intersect "abc" ""`,
			expected: `""`,
		},
		{
			name:     "both empty",
			input:    `intersect "" ""`,
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestStringDifference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple difference",
			input:    `difference "hello" "world"`,
			expected: `"he"`,
		},
		{
			name:     "no common characters",
			input:    `difference "abc" "def"`,
			expected: `"abc"`,
		},
		{
			name:     "all common characters",
			input:    `difference "abc" "abc"`,
			expected: `""`,
		},
		{
			name:     "duplicates removed",
			input:    `difference "aabbcc" "abc"`,
			expected: `""`,
		},
		{
			name:     "empty first string",
			input:    `difference "" "abc"`,
			expected: `""`,
		},
		{
			name:     "empty second string",
			input:    `difference "abc" ""`,
			expected: `"abc"`,
		},
		{
			name:     "both empty",
			input:    `difference "" ""`,
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestStringUnion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple union",
			input:    `union "hello" "world"`,
			expected: `"helowrd"`,
		},
		{
			name:     "no common characters",
			input:    `union "abc" "def"`,
			expected: `"abcdef"`,
		},
		{
			name:     "all common characters",
			input:    `union "abc" "abc"`,
			expected: `"abc"`,
		},
		{
			name:     "duplicates removed",
			input:    `union "aabbcc" "abc"`,
			expected: `"abc"`,
		},
		{
			name:     "empty first string",
			input:    `union "" "abc"`,
			expected: `"abc"`,
		},
		{
			name:     "empty second string",
			input:    `union "abc" ""`,
			expected: `"abc"`,
		},
		{
			name:     "both empty",
			input:    `union "" ""`,
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Mold() != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result.Mold())
			}
		})
	}
}

func TestBinaryIntersect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "simple intersection",
			input:    "intersect #{010203} #{020304}",
			expected: []byte{0x02, 0x03},
		},
		{
			name:     "no common bytes",
			input:    "intersect #{010203} #{040506}",
			expected: []byte{},
		},
		{
			name:     "all common bytes",
			input:    "intersect #{010203} #{010203}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "duplicates removed",
			input:    "intersect #{010101} #{0102}",
			expected: []byte{0x01},
		},
		{
			name:     "empty first binary",
			input:    "intersect #{} #{010203}",
			expected: []byte{},
		},
		{
			name:     "empty second binary",
			input:    "intersect #{010203} #{}",
			expected: []byte{},
		},
		{
			name:     "both empty",
			input:    "intersect #{} #{}",
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			binVal, ok := value.AsBinaryValue(result)
			if !ok {
				t.Fatalf("expected binary value, got %v", result.GetType())
			}
			actual := binVal.Bytes()
			if len(actual) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(actual))
			}
			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Fatalf("at index %d: expected %02x, got %02x", i, tt.expected[i], actual[i])
				}
			}
		})
	}
}

func TestBinaryDifference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "simple difference",
			input:    "difference #{010203} #{020304}",
			expected: []byte{0x01},
		},
		{
			name:     "no common bytes",
			input:    "difference #{010203} #{040506}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "all common bytes",
			input:    "difference #{010203} #{010203}",
			expected: []byte{},
		},
		{
			name:     "duplicates removed",
			input:    "difference #{010101} #{01}",
			expected: []byte{},
		},
		{
			name:     "empty first binary",
			input:    "difference #{} #{010203}",
			expected: []byte{},
		},
		{
			name:     "empty second binary",
			input:    "difference #{010203} #{}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "both empty",
			input:    "difference #{} #{}",
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			binVal, ok := value.AsBinaryValue(result)
			if !ok {
				t.Fatalf("expected binary value, got %v", result.GetType())
			}
			actual := binVal.Bytes()
			if len(actual) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(actual))
			}
			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Fatalf("at index %d: expected %02x, got %02x", i, tt.expected[i], actual[i])
				}
			}
		})
	}
}

func TestBinaryUnion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "simple union",
			input:    "union #{010203} #{020304}",
			expected: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:     "no common bytes",
			input:    "union #{010203} #{040506}",
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		},
		{
			name:     "all common bytes",
			input:    "union #{010203} #{010203}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "duplicates removed",
			input:    "union #{010101} #{0102}",
			expected: []byte{0x01, 0x02},
		},
		{
			name:     "empty first binary",
			input:    "union #{} #{010203}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "empty second binary",
			input:    "union #{010203} #{}",
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "both empty",
			input:    "union #{} #{}",
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			binVal, ok := value.AsBinaryValue(result)
			if !ok {
				t.Fatalf("expected binary value, got %v", result.GetType())
			}
			actual := binVal.Bytes()
			if len(actual) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(actual))
			}
			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Fatalf("at index %d: expected %02x, got %02x", i, tt.expected[i], actual[i])
				}
			}
		})
	}
}
