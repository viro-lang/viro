package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestBitwiseInteger_AND(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "and with all bits set",
			input:    "bit.and 255 15",
			expected: value.NewIntVal(15),
		},
		{
			name:     "and function call",
			input:    "bit.and 6 3",
			expected: value.NewIntVal(2),
		},
		{
			name:     "and with zero",
			input:    "bit.and 42 0",
			expected: value.NewIntVal(0),
		},
		{
			name:     "and with negative",
			input:    "bit.and -1 255",
			expected: value.NewIntVal(255),
		},
		{
			name:     "and negative numbers",
			input:    "bit.and -5 -3",
			expected: value.NewIntVal(-7),
		},
		{
			name:     "and large numbers",
			input:    "bit.and 9223372036854775807 1",
			expected: value.NewIntVal(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_OR(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "or basic",
			input:    "bit.or 6 3",
			expected: value.NewIntVal(7),
		},
		{
			name:     "or function call",
			input:    "bit.or 2 4",
			expected: value.NewIntVal(6),
		},
		{
			name:     "or with zero",
			input:    "bit.or 42 0",
			expected: value.NewIntVal(42),
		},
		{
			name:     "or with all bits set",
			input:    "bit.or 15 240",
			expected: value.NewIntVal(255),
		},
		{
			name:     "or negative numbers",
			input:    "bit.or -5 -3",
			expected: value.NewIntVal(-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_XOR(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "xor basic",
			input:    "bit.xor 6 3",
			expected: value.NewIntVal(5),
		},
		{
			name:     "xor function call",
			input:    "bit.xor 15 10",
			expected: value.NewIntVal(5),
		},
		{
			name:     "xor with zero",
			input:    "bit.xor 42 0",
			expected: value.NewIntVal(42),
		},
		{
			name:     "xor same values",
			input:    "bit.xor 255 255",
			expected: value.NewIntVal(0),
		},
		{
			name:     "xor negative numbers",
			input:    "bit.xor -5 -3",
			expected: value.NewIntVal(6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_NOT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "not zero",
			input:    "bit.not 0",
			expected: value.NewIntVal(-1),
		},
		{
			name:     "not negative one",
			input:    "bit.not -1",
			expected: value.NewIntVal(0),
		},
		{
			name:     "not positive",
			input:    "bit.not 5",
			expected: value.NewIntVal(-6),
		},
		{
			name:     "not all bits set",
			input:    "bit.not 255",
			expected: value.NewIntVal(-256),
		},
		{
			name:     "not large number",
			input:    "bit.not 9223372036854775807",
			expected: value.NewIntVal(-9223372036854775808),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_Shifts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "left shift simple",
			input:    "1 << 2",
			expected: value.NewIntVal(4),
		},
		{
			name:     "left shift using bit.shl",
			input:    "bit.shl 3 4",
			expected: value.NewIntVal(48),
		},
		{
			name:     "right shift simple",
			input:    "8 >> 2",
			expected: value.NewIntVal(2),
		},
		{
			name:     "right shift using bit.shr",
			input:    "bit.shr 32 3",
			expected: value.NewIntVal(4),
		},
		{
			name:     "right shift negative (arithmetic)",
			input:    "-16 >> 2",
			expected: value.NewIntVal(-4),
		},
		{
			name:     "right shift negative by one",
			input:    "-1 >> 1",
			expected: value.NewIntVal(-1),
		},
		{
			name:     "left shift large",
			input:    "1 << 63",
			expected: value.NewIntVal(-9223372036854775808),
		},
		{
			name:     "right shift large",
			input:    "-9223372036854775808 >> 63",
			expected: value.NewIntVal(-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_AND(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "and same length",
			input:    "bit.and #{FF00} #{0FF0}",
			expected: value.NewBinaryValue([]byte{0x0F, 0x00}),
		},
		{
			name:     "and different length - left longer",
			input:    "bit.and #{FF00} #{0F}",
			expected: value.NewBinaryValue([]byte{0x00, 0x00}),
		},
		{
			name:     "and different length - right longer",
			input:    "bit.and #{0F} #{FF00}",
			expected: value.NewBinaryValue([]byte{0x00, 0x00}),
		},
		{
			name:     "and function call",
			input:    "bit.and #{AA} #{55}",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
		{
			name:     "and right-aligned comparison",
			input:    "bit.and #{0102} #{03}",
			expected: value.NewBinaryValue([]byte{0x00, 0x02}),
		},
		{
			name:     "and empty with non-empty",
			input:    "bit.and #{} #{FF}",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
		{
			name:     "and both empty",
			input:    "bit.and #{} #{}",
			expected: value.NewBinaryValue([]byte{}),
		},
		{
			name:     "and from plan example",
			input:    "bit.and #{FFFF} #{FF}",
			expected: value.NewBinaryValue([]byte{0x00, 0xFF}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_OR(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "or same length",
			input:    "bit.or #{0F00} #{F00F}",
			expected: value.NewBinaryValue([]byte{0xFF, 0x0F}),
		},
		{
			name:     "or different length - left longer",
			input:    "bit.or #{FF00} #{0F}",
			expected: value.NewBinaryValue([]byte{0xFF, 0x0F}),
		},
		{
			name:     "or different length - right longer",
			input:    "bit.or #{0F} #{FF00}",
			expected: value.NewBinaryValue([]byte{0xFF, 0x0F}),
		},
		{
			name:     "or function call",
			input:    "bit.or #{AA} #{55}",
			expected: value.NewBinaryValue([]byte{0xFF}),
		},
		{
			name:     "or right-aligned comparison",
			input:    "bit.or #{0102} #{03}",
			expected: value.NewBinaryValue([]byte{0x01, 0x03}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_XOR(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "xor same length",
			input:    "bit.xor #{FF00} #{0FF0}",
			expected: value.NewBinaryValue([]byte{0xF0, 0xF0}),
		},
		{
			name:     "xor different length - left longer",
			input:    "bit.xor #{FF00} #{0F}",
			expected: value.NewBinaryValue([]byte{0xFF, 0x0F}),
		},
		{
			name:     "xor different length - right longer",
			input:    "bit.xor #{0F} #{FF00}",
			expected: value.NewBinaryValue([]byte{0xFF, 0x0F}),
		},
		{
			name:     "xor function call",
			input:    "bit.xor #{AA} #{55}",
			expected: value.NewBinaryValue([]byte{0xFF}),
		},
		{
			name:     "xor right-aligned comparison",
			input:    "bit.xor #{0102} #{03}",
			expected: value.NewBinaryValue([]byte{0x01, 0x01}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_NOT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "not all ones",
			input:    "bit.not #{FF}",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
		{
			name:     "not all zeros",
			input:    "bit.not #{00}",
			expected: value.NewBinaryValue([]byte{0xFF}),
		},
		{
			name:     "not multiple bytes",
			input:    "bit.not #{FF00AA}",
			expected: value.NewBinaryValue([]byte{0x00, 0xFF, 0x55}),
		},
		{
			name:     "not empty",
			input:    "bit.not #{}",
			expected: value.NewBinaryValue([]byte{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_Shifts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "left shift within byte",
			input:    "#{01} << 2",
			expected: value.NewBinaryValue([]byte{0x04}),
		},
		{
			name:     "left shift overflow lost",
			input:    "#{80} << 1",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
		{
			name:     "right shift within byte",
			input:    "#{08} >> 2",
			expected: value.NewBinaryValue([]byte{0x02}),
		},
		{
			name:     "right shift underflow lost",
			input:    "#{01} >> 1",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
		{
			name:     "left shift multi-byte",
			input:    "#{0100} << 8",
			expected: value.NewBinaryValue([]byte{0x00, 0x01}),
		},
		{
			name:     "right shift multi-byte",
			input:    "#{0080} >> 8",
			expected: value.NewBinaryValue([]byte{0x80, 0x00}),
		},
		{
			name:     "left shift using bit.shl",
			input:    "bit.shl #{0F} 4",
			expected: value.NewBinaryValue([]byte{0xF0}),
		},
		{
			name:     "right shift using bit.shr",
			input:    "bit.shr #{F0} 4",
			expected: value.NewBinaryValue([]byte{0x0F}),
		},
		{
			name:     "large shift beyond length",
			input:    "#{FF} << 16",
			expected: value.NewBinaryValue([]byte{0x00}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_BitOn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "set bit 0",
			input:    "bit.on 0 0",
			expected: value.NewIntVal(1),
		},
		{
			name:     "set bit 3",
			input:    "bit.on 0 3",
			expected: value.NewIntVal(8),
		},
		{
			name:     "set already-set bit",
			input:    "bit.on 5 0",
			expected: value.NewIntVal(5),
		},
		{
			name:     "set high bit",
			input:    "bit.on 0 7",
			expected: value.NewIntVal(128),
		},
		{
			name:     "set bit 63",
			input:    "bit.on 0 63",
			expected: value.NewIntVal(-9223372036854775808),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_BitOff(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "clear bit 0",
			input:    "bit.off 1 0",
			expected: value.NewIntVal(0),
		},
		{
			name:     "clear bit 3",
			input:    "bit.off 15 3",
			expected: value.NewIntVal(7),
		},
		{
			name:     "clear already-clear bit",
			input:    "bit.off 4 0",
			expected: value.NewIntVal(4),
		},
		{
			name:     "clear high bit",
			input:    "bit.off 255 7",
			expected: value.NewIntVal(127),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseInteger_Count(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "count zero",
			input:    "bit.count 0",
			expected: value.NewIntVal(0),
		},
		{
			name:     "count single bit",
			input:    "bit.count 8",
			expected: value.NewIntVal(1),
		},
		{
			name:     "count multiple bits",
			input:    "bit.count 15",
			expected: value.NewIntVal(4),
		},
		{
			name:     "count negative",
			input:    "bit.count -1",
			expected: value.NewIntVal(64),
		},
		{
			name:     "count alternating bits",
			input:    "bit.count 170",
			expected: value.NewIntVal(4),
		},
		{
			name:     "count negative two",
			input:    "bit.count -2",
			expected: value.NewIntVal(63),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseBinary_Count(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "count all zeros",
			input:    "bit.count #{00}",
			expected: value.NewIntVal(0),
		},
		{
			name:     "count all ones",
			input:    "bit.count #{FF}",
			expected: value.NewIntVal(8),
		},
		{
			name:     "count multiple bytes",
			input:    "bit.count #{FF000F}",
			expected: value.NewIntVal(12),
		},
		{
			name:     "count alternating pattern",
			input:    "bit.count #{AA55}",
			expected: value.NewIntVal(8),
		},
		{
			name:     "count empty",
			input:    "bit.count #{}",
			expected: value.NewIntVal(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBitwiseErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "mixed types integer and binary",
			input:   "bit.and 3 #{FF}",
			wantErr: true,
			errMsg:  "operands must be same type",
		},
		{
			name:    "wrong type for bit.and",
			input:   `bit.and "hello" "world"`,
			wantErr: true,
			errMsg:  "Type mismatch",
		},
		{
			name:    "bit.on with binary",
			input:   "bit.on #{FF} 2",
			wantErr: true,
			errMsg:  "Type mismatch",
		},
		{
			name:    "bit.off with binary",
			input:   "bit.off #{FF} 2",
			wantErr: true,
			errMsg:  "Type mismatch",
		},
		{
			name:    "shift with wrong type",
			input:   `"hello" << 2`,
			wantErr: true,
			errMsg:  "Type mismatch",
		},
		{
			name:    "negative shift count",
			input:   "1 << -1",
			wantErr: true,
			errMsg:  "shift count must be non-negative",
		},
		{
			name:    "mixed types for bit.or",
			input:   "bit.or 3 #{FF}",
			wantErr: true,
			errMsg:  "operands must be same type",
		},
		{
			name:    "mixed types for bit.xor",
			input:   "bit.xor 3 #{FF}",
			wantErr: true,
			errMsg:  "operands must be same type",
		},
		{
			name:    "negative shift count for bit.shr",
			input:   "bit.shr 8 -1",
			wantErr: true,
			errMsg:  "shift count must be non-negative",
		},
		{
			name:    "non-integer shift count",
			input:   "1 << \"a\"",
			wantErr: true,
			errMsg:  "Type mismatch",
		},
		{
			name:    "bit.and arity error - too few args",
			input:   "bit.and 1",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.or arity error - too few args",
			input:   "bit.or 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.xor arity error - too few args",
			input:   "bit.xor 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.not arity error - too few args",
			input:   "bit.not",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.on arity error - too few args",
			input:   "bit.on 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.off arity error - too few args",
			input:   "bit.off 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.count arity error - too few args",
			input:   "bit.count",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.shl arity error - too few args",
			input:   "bit.shl 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.shr arity error - too few args",
			input:   "bit.shr 5",
			wantErr: true,
			errMsg:  "Wrong argument count",
		},
		{
			name:    "bit.on negative position",
			input:   "bit.on 0 -1",
			wantErr: true,
			errMsg:  "out of range",
		},
		{
			name:    "bit.on position too large",
			input:   "bit.on 0 64",
			wantErr: true,
			errMsg:  "out of range",
		},
		{
			name:    "bit.off negative position",
			input:   "bit.off 15 -1",
			wantErr: true,
			errMsg:  "out of range",
		},
		{
			name:    "bit.off position too large",
			input:   "bit.off 15 64",
			wantErr: true,
			errMsg:  "out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
					return
				}
				if err.Error() == "" || (tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg)) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Equals(value.NewNoneVal()) {
				t.Errorf("Expected non-none result, got none")
			}
		})
	}
}

func TestBitwiseComposition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "left-to-right evaluation",
			input:    "2 << 3 + 1",
			expected: value.NewIntVal(17),
		},
		{
			name:     "parentheses force order",
			input:    "2 << (3 + 1)",
			expected: value.NewIntVal(32),
		},
		{
			name:     "multiple bitwise ops",
			input:    "bit.or (bit.and 15 7) 8",
			expected: value.NewIntVal(15),
		},
		{
			name:     "bit manipulation chain",
			input:    "x: 0\nx: bit.on x 0\nx: bit.on x 2\nx",
			expected: value.NewIntVal(5),
		},
		{
			name:     "mixed operations",
			input:    "bit.on (bit.shl 1 3) 0",
			expected: value.NewIntVal(9),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equals(tt.expected) {
				t.Errorf("Evaluate(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
