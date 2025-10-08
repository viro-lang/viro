package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func evaluateString(t *testing.T, e *eval.Evaluator, input string) (value.Value, *verror.Error) {
	t.Helper()

	vals, err := parse.Parse(input)
	if err != nil {
		return value.NoneVal(), err
	}

	return e.Do_Blk(vals)
}

func TestErrors_UndefinedWord(t *testing.T) {
	e := eval.NewEvaluator()
	_, err := evaluateString(t, e, "missing")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDNoValue {
		t.Fatalf("expected no-value error, got %s", err.ID)
	}

	if err.Message != "No value for word: missing" {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if !strings.Contains(err.Near, "missing") {
		t.Fatalf("near context should mention missing, got %q", err.Near)
	}

	if len(err.Where) == 0 || err.Where[len(err.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", err.Where)
	}
}

func TestErrors_DivideByZero(t *testing.T) {
	e := eval.NewEvaluator()
	_, err := evaluateString(t, e, "10 / 0")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.Category != verror.ErrMath {
		t.Fatalf("expected math error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDDivByZero {
		t.Fatalf("expected div-zero error, got %s", err.ID)
	}

	if err.Message != "Division by zero" {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if !strings.Contains(err.Near, "10") || !strings.Contains(err.Near, "0") || !strings.Contains(err.Near, "/") {
		t.Fatalf("near context should include operands and operator, got %q", err.Near)
	}

	if len(err.Where) == 0 || err.Where[len(err.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", err.Where)
	}
}

func TestErrors_TypeMismatch(t *testing.T) {
	e := eval.NewEvaluator()
	_, err := evaluateString(t, e, "first 42")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDTypeMismatch {
		t.Fatalf("expected type-mismatch error, got %s", err.ID)
	}

	expectedMessage := "Type mismatch for 'first': expected series, got integer"
	if err.Message != expectedMessage {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if !strings.Contains(err.Near, "first") || !strings.Contains(err.Near, "42") {
		t.Fatalf("near context should include expression tokens, got %q", err.Near)
	}

	if len(err.Where) == 0 || err.Where[len(err.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", err.Where)
	}
}

func TestErrors_MathTypeMismatch(t *testing.T) {
	e := eval.NewEvaluator()
	_, err := evaluateString(t, e, "10 + \"oops\"")

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDTypeMismatch {
		t.Fatalf("expected type-mismatch error, got %s", err.ID)
	}

	expectedMessage := "Type mismatch for '+': expected integer, got string"
	if err.Message != expectedMessage {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if !strings.Contains(err.Near, "10") || !strings.Contains(err.Near, "+") || !strings.Contains(err.Near, "oops") {
		t.Fatalf("near context should include expression tokens, got %q", err.Near)
	}

	if len(err.Where) == 0 || err.Where[len(err.Where)-1] != "(top level)" {
		t.Fatalf("expected call stack to include (top level), got %v", err.Where)
	}
}

func TestErrors_ArgumentCount(t *testing.T) {
	e := eval.NewEvaluator()

	if _, err := evaluateString(t, e, "square: fn [n] [n * n]"); err != nil {
		t.Fatalf("function definition failed: %v", err)
	}

	_, err := evaluateString(t, e, "square")
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.Category != verror.ErrScript {
		t.Fatalf("expected script error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDArgCount {
		t.Fatalf("expected arg-count error, got %s", err.ID)
	}

	expectedMessage := "Wrong argument count for 'square': expected 1, got 0"
	if err.Message != expectedMessage {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if !strings.Contains(err.Near, "square") {
		t.Fatalf("near context should include failing word, got %q", err.Near)
	}

	if len(err.Where) == 0 || !contains(err.Where, "square") {
		t.Fatalf("expected call stack to include square, got %v", err.Where)
	}
}

func TestErrors_CallStackPropagation(t *testing.T) {
	e := eval.NewEvaluator()

	if _, err := evaluateString(t, e, "inner: fn [y] [y + missing]"); err != nil {
		t.Fatalf("inner definition failed: %v", err)
	}

	if _, err := evaluateString(t, e, "outer: fn [n] [inner n]"); err != nil {
		t.Fatalf("outer definition failed: %v", err)
	}

	_, err := evaluateString(t, e, "outer 5")
	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if err.ID != verror.ErrIDNoValue {
		t.Fatalf("expected no-value error, got %s", err.ID)
	}

	if !strings.Contains(err.Near, "missing") {
		t.Fatalf("near context should mention missing, got %q", err.Near)
	}

	if len(err.Where) < 3 {
		t.Fatalf("expected call stack with at least three frames, got %v", err.Where)
	}

	if err.Where[0] != "inner" {
		t.Fatalf("expected most recent frame to be inner, got %s", err.Where[0])
	}

	if !contains(err.Where, "outer") {
		t.Fatalf("expected call stack to include outer, got %v", err.Where)
	}
}

func TestParse_UnclosedBlockError(t *testing.T) {
	_, err := parse.Parse("[1 2 3")

	if err == nil {
		t.Fatalf("expected parse error but got none")
	}

	if err.Category != verror.ErrSyntax {
		t.Fatalf("expected syntax error, got %v", err.Category)
	}

	if err.ID != verror.ErrIDUnclosedBlock {
		t.Fatalf("expected unclosed-block error, got %s", err.ID)
	}

	if !strings.Contains(err.Message, "Unclosed block") {
		t.Fatalf("unexpected message: %s", err.Message)
	}

	if err.Near == "" {
		t.Fatalf("expected near context to be populated")
	}

	if len(err.Where) != 0 {
		t.Fatalf("parsing errors should not have call stack, got %v", err.Where)
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
