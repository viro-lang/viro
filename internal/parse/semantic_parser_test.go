package parse

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestParser_ClassifyLiteral_Integers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
	}{
		{"positive integer", "5", value.NewIntVal(5)},
		{"negative integer", "-10", value.NewIntVal(-10)},
		{"zero", "0", value.NewIntVal(0)},
		{"large positive", "999999", value.NewIntVal(999999)},
		{"large negative", "-999999", value.NewIntVal(-999999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if !result.Equals(tt.expected) {
				t.Errorf("classifyLiteral() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_Decimals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType core.ValueType
	}{
		{"simple decimal", "3.14", value.TypeDecimal},
		{"negative decimal", "-3.14", value.TypeDecimal},
		{"scientific notation", "1e10", value.TypeDecimal},
		{"negative scientific", "-1e10", value.TypeDecimal},
		{"scientific with decimal", "3.14e2", value.TypeDecimal},
		{"negative exponent", "1e-5", value.TypeDecimal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != tt.wantType {
				t.Errorf("classifyLiteral() type = %v, want %v", result.GetType(), tt.wantType)
			}
		})
	}
}

func TestParser_ClassifyLiteral_SetWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple set-word", "abc:", "abc"},
		{"set-word with numbers", "x123:", "x123"},
		{"set-word with dash", "my-var:", "my-var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != value.TypeSetWord {
				t.Errorf("classifyLiteral() type = %v, want TypeSetWord", result.GetType())
				return
			}
			if symbol, ok := value.AsWordValue(result); !ok || symbol != tt.expected {
				t.Errorf("classifyLiteral() symbol = %v, want %v", symbol, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_GetWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple get-word", ":abc", "abc"},
		{"get-word with numbers", ":x123", "x123"},
		{"get-word with dash", ":my-var", "my-var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != value.TypeGetWord {
				t.Errorf("classifyLiteral() type = %v, want TypeGetWord", result.GetType())
				return
			}
			if symbol, ok := value.AsWordValue(result); !ok || symbol != tt.expected {
				t.Errorf("classifyLiteral() symbol = %v, want %v", symbol, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_LitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lit-word", "'abc", "abc"},
		{"lit-word with numbers", "'x123", "x123"},
		{"lit-word with dash", "'my-var", "my-var"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != value.TypeLitWord {
				t.Errorf("classifyLiteral() type = %v, want TypeLitWord", result.GetType())
				return
			}
			if symbol, ok := value.AsWordValue(result); !ok || symbol != tt.expected {
				t.Errorf("classifyLiteral() symbol = %v, want %v", symbol, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_Words(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple word", "abc", "abc"},
		{"word with numbers", "x123", "x123"},
		{"word with dash", "my-var", "my-var"},
		{"word with flag", "--flag", "--flag"},
		{"word with double dash", "--debug", "--debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != value.TypeWord {
				t.Errorf("classifyLiteral() type = %v, want TypeWord", result.GetType())
				return
			}
			if symbol, ok := value.AsWordValue(result); !ok || symbol != tt.expected {
				t.Errorf("classifyLiteral() symbol = %v, want %v", symbol, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_Datatypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"integer datatype", "integer!", "integer!"},
		{"string datatype", "string!", "string!"},
		{"object datatype", "object!", "object!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != value.TypeDatatype {
				t.Errorf("classifyLiteral() type = %v, want TypeDatatype", result.GetType())
				return
			}
			if dt, ok := value.AsDatatypeValue(result); !ok || dt != tt.expected {
				t.Errorf("classifyLiteral() datatype = %v, want %v", dt, tt.expected)
			}
		})
	}
}

func TestParser_ClassifyLiteral_Paths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType core.ValueType
	}{
		{"simple path", "obj.field", value.TypePath},
		{"multiple segments", "a.b.c", value.TypePath},
		{"index path", "data.1", value.TypePath},
		{"mixed path", "obj.field.2", value.TypePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != tt.wantType {
				t.Errorf("classifyLiteral() type = %v, want %v", result.GetType(), tt.wantType)
			}
		})
	}
}

func TestParser_ClassifyLiteral_SetPaths(t *testing.T) {
	tests := []struct {
		name string
		input    string
		wantType core.ValueType
	}{
		{"simple set-path", "obj.x:", value.TypeSetPath},
		{"multiple segments set-path", "a.b.c:", value.TypeSetPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != tt.wantType {
				t.Errorf("classifyLiteral() type = %v, want %v", result.GetType(), tt.wantType)
			}
		})
	}
}

func TestParser_ClassifyLiteral_GetPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType core.ValueType
	}{
		{"simple get-path", ":obj.x", value.TypeGetPath},
		{"multiple segments get-path", ":a.b.c", value.TypeGetPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]tokenize.Token{})
			result, err := p.classifyLiteral(tt.input)
			if err != nil {
				t.Errorf("classifyLiteral() error = %v", err)
				return
			}
			if result.GetType() != tt.wantType {
				t.Errorf("classifyLiteral() type = %v, want %v", result.GetType(), tt.wantType)
			}
		})
	}
}

func TestParser_Parse_EmptyBlock(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 1},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 2},
		{Type: tokenize.TokenEOF, Line: 1, Column: 3},
	}

	p := NewParser(tokens)
	result, err := p.Parse()
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Errorf("Parse() len = %d, want 1", len(result))
		return
	}

	if result[0].GetType() != value.TypeBlock {
		t.Errorf("Parse() type = %v, want TypeBlock", result[0].GetType())
	}

	block, ok := value.AsBlockValue(result[0])
	if !ok {
		t.Errorf("Parse() could not extract block")
		return
	}

	if len(block.Elements) != 0 {
		t.Errorf("Parse() block len = %d, want 0", len(block.Elements))
	}
}

func TestParser_Parse_SimpleBlock(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 1},
		{Type: tokenize.TokenLiteral, Value: "1", Line: 1, Column: 2},
		{Type: tokenize.TokenLiteral, Value: "2", Line: 1, Column: 4},
		{Type: tokenize.TokenLiteral, Value: "3", Line: 1, Column: 6},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 7},
		{Type: tokenize.TokenEOF, Line: 1, Column: 8},
	}

	p := NewParser(tokens)
	result, err := p.Parse()
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Errorf("Parse() len = %d, want 1", len(result))
		return
	}

	block, ok := value.AsBlockValue(result[0])
	if !ok {
		t.Errorf("Parse() could not extract block")
		return
	}

	if len(block.Elements) != 3 {
		t.Errorf("Parse() block len = %d, want 3", len(block.Elements))
		return
	}

	expected := []int64{1, 2, 3}
	for i, exp := range expected {
		val, ok := value.AsIntValue(block.Elements[i])
		if !ok || val != exp {
			t.Errorf("Parse() block[%d] = %v, want %d", i, block.Elements[i], exp)
		}
	}
}

func TestParser_Parse_NestedBlocks(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 1},
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 2},
		{Type: tokenize.TokenLiteral, Value: "1", Line: 1, Column: 3},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 4},
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 6},
		{Type: tokenize.TokenLiteral, Value: "2", Line: 1, Column: 7},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 8},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 9},
		{Type: tokenize.TokenEOF, Line: 1, Column: 10},
	}

	p := NewParser(tokens)
	result, err := p.Parse()
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Errorf("Parse() len = %d, want 1", len(result))
		return
	}

	block, ok := value.AsBlockValue(result[0])
	if !ok {
		t.Errorf("Parse() could not extract block")
		return
	}

	if len(block.Elements) != 2 {
		t.Errorf("Parse() block len = %d, want 2", len(block.Elements))
		return
	}

	inner1, ok := value.AsBlockValue(block.Elements[0])
	if !ok {
		t.Errorf("Parse() block[0] is not a block")
		return
	}
	if len(inner1.Elements) != 1 {
		t.Errorf("Parse() block[0] len = %d, want 1", len(inner1.Elements))
	}

	inner2, ok := value.AsBlockValue(block.Elements[1])
	if !ok {
		t.Errorf("Parse() block[1] is not a block")
		return
	}
	if len(inner2.Elements) != 1 {
		t.Errorf("Parse() block[1] len = %d, want 1", len(inner2.Elements))
	}
}

func TestParser_Parse_Parens(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLParen, Value: "(", Line: 1, Column: 1},
		{Type: tokenize.TokenLiteral, Value: "1", Line: 1, Column: 2},
		{Type: tokenize.TokenLiteral, Value: "+", Line: 1, Column: 4},
		{Type: tokenize.TokenLiteral, Value: "2", Line: 1, Column: 6},
		{Type: tokenize.TokenRParen, Value: ")", Line: 1, Column: 7},
		{Type: tokenize.TokenEOF, Line: 1, Column: 8},
	}

	p := NewParser(tokens)
	result, err := p.Parse()
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	if len(result) != 1 {
		t.Errorf("Parse() len = %d, want 1", len(result))
		return
	}

	if result[0].GetType() != value.TypeParen {
		t.Errorf("Parse() type = %v, want TypeParen", result[0].GetType())
	}

	paren, ok := value.AsBlockValue(result[0])
	if !ok {
		t.Errorf("Parse() could not extract paren")
		return
	}

	if len(paren.Elements) != 3 {
		t.Errorf("Parse() paren len = %d, want 3", len(paren.Elements))
	}
}

func TestParser_Parse_Mixed(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLiteral, Value: "abc:", Line: 1, Column: 1},
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 6},
		{Type: tokenize.TokenLiteral, Value: "1", Line: 1, Column: 7},
		{Type: tokenize.TokenLiteral, Value: "2", Line: 1, Column: 9},
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 10},
		{Type: tokenize.TokenEOF, Line: 1, Column: 11},
	}

	p := NewParser(tokens)
	result, err := p.Parse()
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	if len(result) != 2 {
		t.Errorf("Parse() len = %d, want 2", len(result))
		return
	}

	if result[0].GetType() != value.TypeSetWord {
		t.Errorf("Parse() result[0] type = %v, want TypeSetWord", result[0].GetType())
	}

	if result[1].GetType() != value.TypeBlock {
		t.Errorf("Parse() result[1] type = %v, want TypeBlock", result[1].GetType())
	}
}

func TestParser_Parse_UnclosedBlock(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenLBracket, Value: "[", Line: 1, Column: 1},
		{Type: tokenize.TokenLiteral, Value: "1", Line: 1, Column: 2},
		{Type: tokenize.TokenLiteral, Value: "2", Line: 1, Column: 4},
		{Type: tokenize.TokenEOF, Line: 1, Column: 5},
	}

	p := NewParser(tokens)
	_, err := p.Parse()
	if err == nil {
		t.Errorf("Parse() expected error for unclosed block, got nil")
	}
}

func TestParser_Parse_UnexpectedClosingBracket(t *testing.T) {
	tokens := []tokenize.Token{
		{Type: tokenize.TokenRBracket, Value: "]", Line: 1, Column: 1},
		{Type: tokenize.TokenEOF, Line: 1, Column: 2},
	}

	p := NewParser(tokens)
	_, err := p.Parse()
	if err == nil {
		t.Errorf("Parse() expected error for unexpected closing bracket, got nil")
	}
}

func TestParser_ParsePath_SimpleWord(t *testing.T) {
	p := NewParser([]tokenize.Token{})
	segments, err := p.parsePath("obj.field")
	if err != nil {
		t.Errorf("parsePath() error = %v", err)
		return
	}

	if len(segments) != 2 {
		t.Errorf("parsePath() len = %d, want 2", len(segments))
		return
	}

	if segments[0].Type != value.PathSegmentWord || segments[0].Value != "obj" {
		t.Errorf("parsePath() segment[0] = %v, want Word(obj)", segments[0])
	}

	if segments[1].Type != value.PathSegmentWord || segments[1].Value != "field" {
		t.Errorf("parsePath() segment[1] = %v, want Word(field)", segments[1])
	}
}

func TestParser_ParsePath_IndexPath(t *testing.T) {
	p := NewParser([]tokenize.Token{})
	segments, err := p.parsePath("data.1")
	if err != nil {
		t.Errorf("parsePath() error = %v", err)
		return
	}

	if len(segments) != 2 {
		t.Errorf("parsePath() len = %d, want 2", len(segments))
		return
	}

	if segments[0].Type != value.PathSegmentWord || segments[0].Value != "data" {
		t.Errorf("parsePath() segment[0] = %v, want Word(data)", segments[0])
	}

	if segments[1].Type != value.PathSegmentIndex || segments[1].Value != int64(1) {
		t.Errorf("parsePath() segment[1] = %v, want Index(1)", segments[1])
	}
}

func TestParser_ParsePath_EmptySegment(t *testing.T) {
	p := NewParser([]tokenize.Token{})
	_, err := p.parsePath("obj.")
	if err == nil {
		t.Errorf("parsePath() expected error for empty segment, got nil")
	}
}

func TestParser_ParsePath_EmptyPath(t *testing.T) {
	p := NewParser([]tokenize.Token{})
	_, err := p.parsePath("")
	if err == nil {
		t.Errorf("parsePath() expected error for empty path, got nil")
	}
}
