package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestParser_Tokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(core.Value) bool
		wantErr  bool
	}{
		{
			name:  "tokenize simple literal",
			input: `tokenize "42"`,
			expected: func(v core.Value) bool {
				block, ok := value.AsBlockValue(v)
				if !ok || len(block.Elements) != 1 {
					return false
				}
				obj, ok := value.AsObject(block.Elements[0])
				if !ok {
					return false
				}
				typeVal, _ := obj.GetField("type")
				typeName, ok := value.AsWordValue(typeVal)
				return ok && typeName == "literal"
			},
			wantErr: false,
		},
		{
			name:  "tokenize block",
			input: `tokenize "[1 2 3]"`,
			expected: func(v core.Value) bool {
				block, ok := value.AsBlockValue(v)
				return ok && len(block.Elements) == 5
			},
			wantErr: false,
		},
		{
			name:  "tokenize string",
			input: `tokenize "\"hello\""`,
			expected: func(v core.Value) bool {
				block, ok := value.AsBlockValue(v)
				if !ok || len(block.Elements) != 1 {
					return false
				}
				obj, ok := value.AsObject(block.Elements[0])
				if !ok {
					return false
				}
				typeVal, _ := obj.GetField("type")
				typeName, ok := value.AsWordValue(typeVal)
				return ok && typeName == "string"
			},
			wantErr: false,
		},
		{
			name:  "tokenize word assignment",
			input: `tokenize "x: 42"`,
			expected: func(v core.Value) bool {
				block, ok := value.AsBlockValue(v)
				return ok && len(block.Elements) == 2
			},
			wantErr: false,
		},
		{
			name:    "tokenize empty string",
			input:   `tokenize ""`,
			expected: func(v core.Value) bool {
				block, ok := value.AsBlockValue(v)
				return ok && len(block.Elements) == 0
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.expected(result) {
				t.Errorf("Result validation failed for input: %s, got: %v", tt.input, result.Mold())
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "parse simple literal",
			input:    `parse tokenize "42"`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(42)}),
			wantErr:  false,
		},
		{
			name:     "parse block",
			input:    `parse tokenize "[1 2 3]"`,
			expected: value.NewBlockVal([]core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)})}),
			wantErr:  false,
		},
		{
			name:     "parse string",
			input:    `parse tokenize "\"hello\""`,
			expected: value.NewBlockVal([]core.Value{value.NewStrVal("hello")}),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParser_LoadString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "load-string simple literal",
			input:    `load-string "42"`,
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(42)}),
			wantErr:  false,
		},
		{
			name:     "load-string block",
			input:    `load-string "[1 2 3]"`,
			expected: value.NewBlockVal([]core.Value{value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)})}),
			wantErr:  false,
		},
		{
			name:     "load-string multiple values",
			input:    `load-string "x: 42"`,
			expected: value.NewBlockVal([]core.Value{value.NewSetWordVal("x"), value.NewIntVal(42)}),
			wantErr:  false,
		},
		{
			name:     "load-string string",
			input:    `load-string "\"hello\""`,
			expected: value.NewBlockVal([]core.Value{value.NewStrVal("hello")}),
			wantErr:  false,
		},
		{
			name:     "load-string empty",
			input:    `load-string ""`,
			expected: value.NewBlockVal([]core.Value{}),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParser_Classify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "classify integer",
			input:    `classify "42"`,
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "classify negative integer",
			input:    `classify "-42"`,
			expected: value.NewIntVal(-42),
			wantErr:  false,
		},
		{
			name:     "classify word",
			input:    `classify "hello"`,
			expected: value.NewWordVal("hello"),
			wantErr:  false,
		},
		{
			name:     "classify set-word",
			input:    `classify "x:"`,
			expected: value.NewSetWordVal("x"),
			wantErr:  false,
		},
		{
			name:     "classify get-word",
			input:    `classify ":x"`,
			expected: value.NewGetWordVal("x"),
			wantErr:  false,
		},
		{
			name:     "classify lit-word",
			input:    `classify "'x"`,
			expected: value.NewLitWordVal("x"),
			wantErr:  false,
		},
		{
			name:     "classify datatype",
			input:    `classify "integer!"`,
			expected: value.NewDatatypeVal("integer!"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}

func TestParser_Integration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "tokenize then parse equals load-string",
			input:    `(= (parse tokenize "42") (load-string "42"))`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name:     "classify literal",
			input:    `classify "42"`,
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected.Mold(), result.Mold())
			}
		})
	}
}
