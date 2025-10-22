package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestErrors_UndefinedWord(t *testing.T) {
	_, err := Evaluate("missing")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDNoValue {
		t.Fatalf("expected no-value error, got %s", vErr.ID)
	}

	if vErr.Message != "No value for word: missing" {
		t.Fatalf("unexpected message: %s", vErr.Message)
	}

	if !strings.Contains(vErr.Near, "missing") {
		t.Fatalf("near context should mention missing, got %q", vErr.Near)
	}

	if len(vErr.Where) == 0 || vErr.Where[len(vErr.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", vErr.Where)
	}
}

func TestErrors_DivideByZero(t *testing.T) {
	_, err := Evaluate("10 / 0")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrMath {
		t.Fatalf("expected math error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDDivByZero {
		t.Fatalf("expected div-zero error, got %s", vErr.ID)
	}

	if vErr.Message != "Division by zero" {
		t.Fatalf("unexpected message: %s", vErr.Message)
	}

	if !strings.Contains(vErr.Near, "10") || !strings.Contains(vErr.Near, "0") || !strings.Contains(vErr.Near, "/") {
		t.Fatalf("near context should include operands and operator, got %q", vErr.Near)
	}

	if len(vErr.Where) == 0 || vErr.Where[len(vErr.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", vErr.Where)
	}
}

func TestErrors_TypeMismatch(t *testing.T) {
	_, err := Evaluate("first 42")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDActionNoImpl {
		t.Fatalf("expected action-no-impl error, got %s", vErr.ID)
	}

	if !strings.Contains(vErr.Near, "first") || !strings.Contains(vErr.Near, "42") {
		t.Fatalf("near context should include expression tokens, got %q", vErr.Near)
	}

	if len(vErr.Where) == 0 || vErr.Where[len(vErr.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", vErr.Where)
	}
}

func TestErrors_MathTypeMismatch(t *testing.T) {
	_, err := Evaluate("10 + \"oops\"")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDTypeMismatch {
		t.Fatalf("expected type-mismatch error, got %s", vErr.ID)
	}

	expectedMessage := "Type mismatch for '+': expected integer, got string!"
	if vErr.Message != expectedMessage {
		t.Fatalf("unexpected message: %s", vErr.Message)
	}

	if !strings.Contains(vErr.Near, "10") || !strings.Contains(vErr.Near, "+") || !strings.Contains(vErr.Near, "oops") {
		t.Fatalf("near context should include expression tokens, got %q", vErr.Near)
	}

	if len(vErr.Where) == 0 || vErr.Where[len(vErr.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", vErr.Where)
	}
}

func TestErrors_ArgumentCount(t *testing.T) {
	_, err := Evaluate("square: fn [n] [n * n]\nsquare")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDArgCount {
		t.Fatalf("expected arg-count error, got %s", vErr.ID)
	}

	expectedMessage := "Wrong argument count for 'square': expected 1, got 0"
	if vErr.Message != expectedMessage {
		t.Fatalf("unexpected message: %s", vErr.Message)
	}

	if !strings.Contains(vErr.Near, "square") {
		t.Fatalf("near context should include failing word, got %q", vErr.Near)
	}

	if len(vErr.Where) == 0 || !contains(vErr.Where, "square") {
		t.Fatalf("expected call stack to include square, got %v", vErr.Where)
	}
}

func TestErrors_CallStackPropagation(t *testing.T) {
	script := `inner: fn [y] [y + missing]
outer: fn [n] [inner n]
outer 5`

	_, err := Evaluate(script)
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.ID != verror.ErrIDNoValue {
		t.Fatalf("expected no-value error, got %s", vErr.ID)
	}

	if !strings.Contains(vErr.Near, "missing") {
		t.Fatalf("near context should mention missing, got %q", vErr.Near)
	}

	if len(vErr.Where) < 3 {
		t.Fatalf("expected call stack with at least three frames, got %v", vErr.Where)
	}

	if !contains(vErr.Where, "inner") {
		t.Fatalf("expected call stack to include inner, got %v", vErr.Where)
	}

	if !contains(vErr.Where, "outer") {
		t.Fatalf("expected call stack to include outer, got %v", vErr.Where)
	}
}

func TestParse_UnclosedBlockError(t *testing.T) {
	_, err := parse.Parse("[1 2 3")

	if err == nil {
		t.Fatalf("expected parse error but got none")
	}

	vErr := err.(*verror.Error)

	if vErr.Category != verror.ErrSyntax {
		t.Fatalf("expected syntax error, got %v", vErr.Category)
	}

	if vErr.ID != verror.ErrIDUnclosedBlock {
		t.Fatalf("expected unclosed-block error, got %s", vErr.ID)
	}

	if !strings.Contains(vErr.Message, "Unclosed block") {
		t.Fatalf("unexpected message: %s", vErr.Message)
	}

	if vErr.Near == "" {
		t.Fatalf("expected near context to be populated")
	}

	if len(vErr.Where) != 0 {
		t.Fatalf("parsing errors should not have call stack, got %v", vErr.Where)
	}
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
