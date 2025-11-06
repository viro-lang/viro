package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
)

// Test that object field definitions don't leak to outer scope
func TestObjectFieldIsolation(t *testing.T) {
	code := `o: object [a: 10]`

	values, err := parse.ParseWithSource(code, "(test)")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	e := NewTestEvaluator()

	_, evalErr := e.DoBlock(values)
	if evalErr != nil {
		t.Fatalf("eval error: %v", evalErr)
	}

	t.Logf("Active frames count after object creation: %d", len(e.Frames))
	for i, f := range e.Frames {
		words := make([]string, 0)
		for _, binding := range f.GetAll() {
			words = append(words, binding.Symbol)
		}
		t.Logf("Frame %d: type=%v, parent=%d, words=%v", i, f.GetType(), f.GetParent(), words)
	}

	// Now try to lookup 'a' - it should NOT be found
	code2 := `a`
	values2, _ := parse.ParseWithSource(code2, "(test)")
	_, err2 := e.DoBlock(values2)

	if err2 == nil {
		t.Fatal("BUG: Variable 'a' from object scope leaked to outer scope! Should have gotten 'no-value' error.")
	}

	t.Logf("Correctly got error when accessing object field from outer scope: %v", err2)
}
