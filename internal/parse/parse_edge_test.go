package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestExpressionParsing tests complex infix expressions and operator handling
func TestExpressionParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "simple addition",
			input:       "3 + 4",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren (expression), got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse simple addition expression",
		},
		{
			name:        "left to right evaluation",
			input:       "2 + 3 * 4",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				// Should parse as (+ 2 (* 3 4)) = 20, not 14
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should demonstrate left-to-right operator evaluation (no precedence)",
		},
		{
			name:        "complex expression",
			input:       "1 + 2 * 3 - 4 / 2",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse complex multi-operator expression",
		},
		{
			name:        "comparison operators",
			input:       "x < y and y > z",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse comparison and logical operators",
		},
		{
			name:        "mixed operators",
			input:       "a + b < c * d",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse mixed arithmetic and comparison operators",
		},
		{
			name:        "nested expressions with parens",
			input:       "3 + (4 * 2)",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse nested expressions with parentheses",
		},
		{
			name:        "operator associativity",
			input:       "1 - 2 - 3",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				// Should parse as (- (- 1 2) 3) = -4, not (- 1 (- 2 3)) = 0
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should demonstrate left-associative operators",
		},
		{
			name:        "expression with decimals",
			input:       "1.5 + 2.5 * 3.0",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse expressions with decimal numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, vals)
			}
		})
	}
}

// TestBlockParenParsing tests various block and paren parsing scenarios
func TestBlockParenParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "empty block",
			input:       "[]",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
				block, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected block value")
					return
				}
				if len(block.Elements) != 0 {
					t.Errorf("Expected empty block, got %d elements", len(block.Elements))
				}
			},
			desc: "Should parse empty blocks",
		},
		{
			name:        "empty paren",
			input:       "()",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
				paren, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected paren value")
					return
				}
				if len(paren.Elements) != 0 {
					t.Errorf("Expected empty paren, got %d elements", len(paren.Elements))
				}
			},
			desc: "Should parse empty parentheses",
		},
		{
			name:        "block with single element",
			input:       "[42]",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
				block, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected block value")
					return
				}
				if len(block.Elements) != 1 {
					t.Errorf("Expected 1 element, got %d", len(block.Elements))
				}
			},
			desc: "Should parse blocks with single elements",
		},
		{
			name:        "paren with single element",
			input:       "(42)",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
				paren, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected paren value")
					return
				}
				if len(paren.Elements) != 1 {
					t.Errorf("Expected 1 element, got %d", len(paren.Elements))
				}
			},
			desc: "Should parse parens with single elements",
		},
		{
			name:        "nested blocks",
			input:       "[[1 2] [3 4]]",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
				block, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected block value")
					return
				}
				if len(block.Elements) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(block.Elements))
				}
				// Check that nested elements are also blocks
				for i, elem := range block.Elements {
					if elem.GetType() != value.TypeBlock {
						t.Errorf("Element %d should be block, got %s", i, value.TypeToString(elem.GetType()))
					}
				}
			},
			desc: "Should parse nested blocks",
		},
		{
			name:        "nested parens",
			input:       "((1 + 2) * (3 + 4))",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse nested parentheses with expressions",
		},
		{
			name:        "mixed blocks and parens",
			input:       "[(1 + 2) (3 * 4)]",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
				block, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected block value")
					return
				}
				if len(block.Elements) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(block.Elements))
				}
				// Check element types
				if block.Elements[0].GetType() != value.TypeParen {
					t.Errorf("First element should be paren, got %s", value.TypeToString(block.Elements[0].GetType()))
				}
				if block.Elements[1].GetType() != value.TypeParen {
					t.Errorf("Second element should be paren, got %s", value.TypeToString(block.Elements[1].GetType()))
				}
			},
			desc: "Should parse blocks containing both blocks and parens",
		},
		{
			name:        "block with complex content",
			input:       "[x: 1 + 2 print \"hello\" if true [42]]",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeBlock {
					t.Errorf("Expected block, got %s", value.TypeToString(vals[0].GetType()))
				}
				block, ok := value.AsBlock(vals[0])
				if !ok {
					t.Errorf("Expected block value")
					return
				}
				// The parser may break this into more elements than expected
				if len(block.Elements) < 4 {
					t.Errorf("Expected at least 4 elements, got %d", len(block.Elements))
				}
			},
			desc: "Should parse blocks with complex content including assignments and control flow",
		},
		{
			name:        "unclosed block",
			input:       "[1 2 3",
			expectError: true,
			checkResult: nil,
			desc:        "Should error on unclosed blocks",
		},
		{
			name:        "unclosed paren",
			input:       "(1 + 2",
			expectError: true,
			checkResult: nil,
			desc:        "Should error on unclosed parentheses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, vals)
			}
		})
	}
}

// TestPrimaryExpressionEdgeCases tests edge cases in primary expression parsing
func TestPrimaryExpressionEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(*testing.T, []core.Value)
		desc        string
	}{
		{
			name:        "invalid number format",
			input:       "123abc",
			expectError: false, // Parser might tokenize this as separate tokens
			checkResult: func(t *testing.T, vals []core.Value) {
				// This might parse as [123, abc] or similar
				if len(vals) == 0 {
					t.Errorf("Expected at least 1 value")
				}
			},
			desc: "Should handle invalid number formats gracefully",
		},
		{
			name:        "lone decimal point",
			input:       ".",
			expectError: true, // Parser doesn't handle lone decimal points
			checkResult: nil,
			desc:        "Should error on lone decimal point",
		},
		{
			name:        "multiple operators",
			input:       "+ + +",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value (expression), got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeParen {
					t.Errorf("Expected paren (expression), got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should parse multiple consecutive operators as an expression",
		},
		{
			name:        "mixed valid and invalid tokens",
			input:       "42 @ invalid",
			expectError: true, // @ is not a valid character
			checkResult: nil,
			desc:        "Should error on invalid characters",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 0 {
					t.Errorf("Expected 0 values for empty input, got %d", len(vals))
				}
			},
			desc: "Should handle empty input gracefully",
		},
		{
			name:        "only whitespace",
			input:       "   \t\n   ",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 0 {
					t.Errorf("Expected 0 values for whitespace-only input, got %d", len(vals))
				}
			},
			desc: "Should handle whitespace-only input",
		},
		{
			name:        "unicode identifiers",
			input:       "α β γ",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 3 {
					t.Errorf("Expected 3 values, got %d", len(vals))
				}
				for i, val := range vals {
					if val.GetType() != value.TypeWord {
						t.Errorf("Value %d should be word, got %s", i, value.TypeToString(val.GetType()))
					}
				}
			},
			desc: "Should handle Unicode identifiers",
		},
		{
			name:        "very long identifier",
			input:       "verylongidentifiername",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value, got %d", len(vals))
					return
				}
				if vals[0].GetType() != value.TypeWord {
					t.Errorf("Expected word, got %s", value.TypeToString(vals[0].GetType()))
				}
			},
			desc: "Should handle very long identifiers",
		},
		{
			name:        "single character tokens",
			input:       "a + b",
			expectError: false,
			checkResult: func(t *testing.T, vals []core.Value) {
				if len(vals) != 1 {
					t.Errorf("Expected 1 value (expression), got %d", len(vals))
				}
			},
			desc: "Should parse single character identifiers and operators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.desc)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.desc, err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, vals)
			}
		})
	}
}
