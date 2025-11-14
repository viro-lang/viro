package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestBinaryLiteral_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty binary",
			input: "#{}",
			want:  "#{}",
		},
		{
			name:  "single byte",
			input: "#{FF}",
			want:  "#{FF}",
		},
		{
			name:  "multiple bytes",
			input: "#{DEADBEEF}",
			want:  "#{DEADBEEF}",
		},
		{
			name:  "lowercase hex",
			input: "#{deadbeef}",
			want:  "#{DEADBEEF}",
		},
		{
			name:  "mixed case",
			input: "#{DeAdBeEf}",
			want:  "#{DEADBEEF}",
		},
		{
			name:  "no spaces in input",
			input: "#{DEADBEEF}",
			want:  "#{DEADBEEF}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != value.TypeBinary {
				t.Fatalf("expected binary type, got %s", value.TypeToString(result.GetType()))
			}

			molded := result.Mold()
			if molded != tt.want {
				t.Errorf("Mold() = %q, want %q", molded, tt.want)
			}
		})
	}
}

func TestBinaryLiteral_TypeChecking(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  core.Value
	}{
		{
			name:  "type-of empty binary",
			input: "type-of #{}",
			want:  value.NewWordVal("binary!"),
		},
		{
			name:  "type-of non-empty binary",
			input: "type-of #{FF}",
			want:  value.NewWordVal("binary!"),
		},
		{
			name:  "type-of multi-byte binary",
			input: "type-of #{DEADBEEF}",
			want:  value.NewWordVal("binary!"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.Equals(tt.want) {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBinaryLiteral_InvalidCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "odd hex digit count",
			input:       "#{A}",
			wantErr:     true,
			expectedErr: "invalid-binary-length",
		},
		{
			name:        "odd hex digit count - three digits",
			input:       "#{ABC}",
			wantErr:     true,
			expectedErr: "invalid-binary-length",
		},
		{
			name:        "invalid hex character G",
			input:       "#{GG}",
			wantErr:     true,
			expectedErr: "invalid-binary-digit",
		},
		{
			name:        "invalid hex character X",
			input:       "#{XY}",
			wantErr:     true,
			expectedErr: "invalid-binary-digit",
		},
		{
			name:        "invalid hex character Z",
			input:       "#{ZZ}",
			wantErr:     true,
			expectedErr: "invalid-binary-digit",
		},
		{
			name:        "mixed valid and invalid",
			input:       "#{FFGG}",
			wantErr:     true,
			expectedErr: "invalid-binary-digit",
		},
		{
			name:        "special characters",
			input:       "#{@!}",
			wantErr:     true,
			expectedErr: "invalid-character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.expectedErr != "" {
				if vErr, ok := err.(*verror.Error); !ok || vErr.ID != tt.expectedErr {
					t.Errorf("expected error ID %s, got: %v (type %T)", tt.expectedErr, err, err)
				}
			}
		})
	}
}

func TestBinaryLiteral_Values(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{
			name:  "empty binary has zero length",
			input: "length? #{}",
			want:  []byte{},
		},
		{
			name:  "single byte value",
			input: "#{FF}",
			want:  []byte{0xFF},
		},
		{
			name:  "two byte value",
			input: "#{DEAD}",
			want:  []byte{0xDE, 0xAD},
		},
		{
			name:  "four byte value",
			input: "#{DEADBEEF}",
			want:  []byte{0xDE, 0xAD, 0xBE, 0xEF},
		},
		{
			name:  "zero bytes",
			input: "#{0000}",
			want:  []byte{0x00, 0x00},
		},
		{
			name:  "mixed values",
			input: "#{01FF}",
			want:  []byte{0x01, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.name == "empty binary has zero length" {
				intVal, _ := value.AsIntValue(result)
				if intVal != 0 {
					t.Errorf("length? #{} = %d, want 0", intVal)
				}
				return
			}

			if result.GetType() != value.TypeBinary {
				t.Fatalf("expected binary type, got %s", value.TypeToString(result.GetType()))
			}

			binVal, _ := value.AsBinaryValue(result)
			gotBytes := binVal.Bytes()

			if len(gotBytes) != len(tt.want) {
				t.Fatalf("length mismatch: got %d bytes, want %d bytes", len(gotBytes), len(tt.want))
			}

			for i := range tt.want {
				if gotBytes[i] != tt.want[i] {
					t.Errorf("byte[%d] = 0x%02X, want 0x%02X", i, gotBytes[i], tt.want[i])
				}
			}
		})
	}
}

func TestBinaryLiteral_InExpressions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    core.Value
		wantErr bool
	}{
		{
			name:  "append to binary",
			input: "append #{FF} 0",
			want:  value.NewBinaryValue([]byte{0xFF, 0x00}),
		},
		{
			name:  "length of binary",
			input: "length? #{DEADBEEF}",
			want:  value.NewIntVal(4),
		},
		{
			name:  "first of binary",
			input: "first #{DEADBEEF}",
			want:  value.NewIntVal(0xDE),
		},
		{
			name:  "last of binary",
			input: "last #{DEADBEEF}",
			want:  value.NewIntVal(0xEF),
		},
		{
			name:  "first of empty binary returns none",
			input: "first #{}",
			want:  value.NewNoneVal(),
		},
		{
			name:  "assignment to variable",
			input: "data: #{CAFE}  data",
			want:  value.NewBinaryValue([]byte{0xCA, 0xFE}),
		},
		{
			name:  "binary equality",
			input: "= #{FF} #{FF}",
			want:  value.NewLogicVal(true),
		},
		{
			name:  "binary inequality different values",
			input: "= #{FF} #{00}",
			want:  value.NewLogicVal(false),
		},
		{
			name:  "binary inequality different lengths",
			input: "= #{FF} #{FFFF}",
			want:  value.NewLogicVal(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if !result.Equals(tt.want) {
				t.Errorf("got %v (%s), want %v (%s)",
					result.Mold(), value.TypeToString(result.GetType()),
					tt.want.Mold(), value.TypeToString(tt.want.GetType()))
			}
		})
	}
}
