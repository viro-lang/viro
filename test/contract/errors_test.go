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
	sourceName := "test-script.viro"
	_, err := parse.ParseWithSource("[1 2 3", sourceName)

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

	if vErr.File != sourceName {
		t.Fatalf("expected file %q, got %q", sourceName, vErr.File)
	}

	if vErr.Line == 0 || vErr.Column == 0 {
		t.Fatalf("expected line and column to be set, got %d:%d", vErr.Line, vErr.Column)
	}
}

func TestRuntimeErrorIncludesLocation(t *testing.T) {
	script := "print 1\nmissing\n"

	values, err := parse.ParseWithSource(script, "(test)")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	evaluator := NewTestEvaluator()
	_, runtimeErr := evaluator.DoBlock(values)
	if runtimeErr == nil {
		t.Fatal("expected runtime error but got nil")
	}

	vErr, ok := runtimeErr.(*verror.Error)
	if !ok {
		t.Fatalf("expected *verror.Error, got %T", runtimeErr)
	}

	if vErr.Line != 2 || vErr.Column != 1 {
		t.Fatalf("expected error location at line 2 column 1, got line %d column %d", vErr.Line, vErr.Column)
	}
}

func TestRuntimeErrorNestedLocation(t *testing.T) {
	script := "fn: fn [] [\n    missing\n]\nfn\n"

	values, err := parse.ParseWithSource(script, "(test)")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	evaluator := NewTestEvaluator()
	_, runtimeErr := evaluator.DoBlock(values)
	if runtimeErr == nil {
		t.Fatal("expected runtime error but got nil")
	}

	vErr, ok := runtimeErr.(*verror.Error)
	if !ok {
		t.Fatalf("expected *verror.Error, got %T", runtimeErr)
	}

	if vErr.Line != 2 || vErr.Column != 5 {
		t.Fatalf("expected error location at line 2 column 5, got line %d column %d", vErr.Line, vErr.Column)
	}
}

func TestErrors_ActionNoImpl(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
		wantMsg   string
	}{
		{
			name:      "first on integer",
			input:     "first 42",
			wantErrID: "action-no-impl",
			wantMsg:   "first",
		},
		{
			name:      "last on integer",
			input:     "last 42",
			wantErrID: "action-no-impl",
			wantMsg:   "last",
		},
		{
			name:      "append on integer",
			input:     "append 42 1",
			wantErrID: "action-no-impl",
			wantMsg:   "append",
		},
		{
			name:      "insert on logic",
			input:     "insert true false",
			wantErrID: "action-no-impl",
			wantMsg:   "insert",
		},
		{
			name:      "length? on function",
			input:     "length? fn [x] [x]",
			wantErrID: "action-no-impl",
			wantMsg:   "length?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, evalErr := Evaluate(tt.input)

			if evalErr == nil {
				t.Fatal("Expected error but got nil")
			}

			vErr, ok := evalErr.(*verror.Error)
			if !ok {
				t.Fatalf("Expected *verror.Error, got %T", evalErr)
			}

			if vErr.ID != tt.wantErrID {
				t.Errorf("Expected error ID %s, got %s", tt.wantErrID, vErr.ID)
			}

			t.Logf("Error message: %s", vErr.Error())
		})
	}
}

func TestErrors_ActionWrongArity(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "first with no args",
			input:     "first",
			wantErrID: "arg-count",
		},
		{
			name:      "last with no args",
			input:     "last",
			wantErrID: "arg-count",
		},
		{
			name:      "append with one arg",
			input:     "append [1 2]",
			wantErrID: "arg-count",
		},
		{
			name:      "insert with one arg",
			input:     "insert [1 2]",
			wantErrID: "arg-count",
		},
		{
			name:      "length? with no args",
			input:     "length?",
			wantErrID: "arg-count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, evalErr := Evaluate(tt.input)

			if evalErr == nil {
				t.Fatal("Expected error but got nil")
			}

			vErr, ok := evalErr.(*verror.Error)
			if !ok {
				t.Fatalf("Expected *verror.Error, got %T", evalErr)
			}

			if vErr.ID != tt.wantErrID {
				t.Errorf("Expected error ID %s, got %s", tt.wantErrID, vErr.ID)
			}

			t.Logf("Error message: %s", vErr.Error())
		})
	}
}

func TestErrors_ActionEmptySeries(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "first on empty block",
			input:     "first []",
			wantErrID: "empty-series",
		},
		{
			name:      "first on empty string",
			input:     `first ""`,
			wantErrID: "empty-series",
		},
		{
			name:      "last on empty block",
			input:     "last []",
			wantErrID: "empty-series",
		},
		{
			name:      "last on empty string",
			input:     `last ""`,
			wantErrID: "empty-series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, evalErr := Evaluate(tt.input)

			if evalErr == nil {
				t.Fatal("Expected error but got nil")
			}

			vErr, ok := evalErr.(*verror.Error)
			if !ok {
				t.Fatalf("Expected *verror.Error, got %T", evalErr)
			}

			if vErr.ID != tt.wantErrID {
				t.Errorf("Expected error ID %s, got %s", tt.wantErrID, vErr.ID)
			}

			t.Logf("Error message: %s", vErr.Error())
		})
	}
}

func TestErrors_ActionTypeMismatch(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErrID string
	}{
		{
			name:      "append integer to string",
			input:     `append "hello" 42`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "append block to string",
			input:     `append "hello" [1 2]`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "insert integer to string",
			input:     `insert "hello" 42`,
			wantErrID: "type-mismatch",
		},
		{
			name:      "insert block to string",
			input:     `insert "hello" [1 2]`,
			wantErrID: "type-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, evalErr := Evaluate(tt.input)

			if evalErr == nil {
				t.Fatal("Expected error but got nil")
			}

			vErr, ok := evalErr.(*verror.Error)
			if !ok {
				t.Fatalf("Expected *verror.Error, got %T", evalErr)
			}

			if vErr.ID != tt.wantErrID {
				t.Errorf("Expected error ID %s, got %s", tt.wantErrID, vErr.ID)
			}

			t.Logf("Error message: %s", vErr.Error())
		})
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
