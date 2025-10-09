package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
)

// Test that object field definitions don't leak to outer scope
func TestObjectFieldIsolation(t *testing.T) {
	code := `o: object [a: 10]`

	values, err := parse.Parse(code)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	e := eval.NewEvaluator()

	_, evalErr := e.Do_Blk(values)
	if evalErr != nil {
		t.Fatalf("eval error: %v", evalErr)
	}

	t.Logf("Active frames count after object creation: %d", len(e.Frames))
	for i, f := range e.Frames {
		t.Logf("Frame %d: type=%v, parent=%d, words=%v", i, f.Type, f.Parent, f.Words)
	}

	// Now try to lookup 'a' - it should NOT be found
	code2 := `a`
	values2, _ := parse.Parse(code2)
	_, err2 := e.Do_Blk(values2)

	if err2 == nil {
		t.Fatal("BUG: Variable 'a' from object scope leaked to outer scope! Should have gotten 'no-value' error.")
	}

	t.Logf("Correctly got error when accessing object field from outer scope: %v", err2)
}
