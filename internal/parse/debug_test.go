package parse

import (
	"testing"
)

func TestDebugPathTokenization(t *testing.T) {
	input := `obj: object [name: "Alice"] obj.name`

	tokens, err := tokenize(input)
	if err != nil {
		t.Fatalf("Tokenization error: %v", err)
	}

	t.Logf("Token count: %d", len(tokens))
	for i, tok := range tokens {
		t.Logf("Token %d: type=%d, val=%q, pos=%d", i, tok.typ, tok.val, tok.pos)
	}

	// Now test parsing
	vals, parseErr := Parse(input)
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	t.Logf("Parsed %d values", len(vals))
	for i, val := range vals {
		t.Logf("Value %d: type=%v, val=%v", i, val.Type, val)
	}
}
