package native

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestArityError(t *testing.T) {
	err := arityError("+", 2, 3)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Category != verror.ErrScript {
		t.Fatalf("unexpected category: %v", err.Category)
	}
	if err.ID != verror.ErrIDArgCount {
		t.Fatalf("unexpected id: %s", err.ID)
	}
	if err.Args != [3]string{"+", "2", "3"} {
		t.Fatalf("unexpected args: %#v", err.Args)
	}
}

func TestTypeError(t *testing.T) {
	val := value.IntVal(10)
	err := typeError("add", "integer", val)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Category != verror.ErrScript {
		t.Fatalf("unexpected category: %v", err.Category)
	}
	if err.ID != verror.ErrIDTypeMismatch {
		t.Fatalf("unexpected id: %s", err.ID)
	}
	if err.Args != [3]string{"add", "integer", val.Type.String()} {
		t.Fatalf("unexpected args: %#v", err.Args)
	}
}

func TestMathTypeError(t *testing.T) {
	val := value.StrVal("oops")
	err := mathTypeError("/", val)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Category != verror.ErrScript {
		t.Fatalf("unexpected category: %v", err.Category)
	}
	if err.ID != verror.ErrIDTypeMismatch {
		t.Fatalf("unexpected id: %s", err.ID)
	}
	if err.Args != [3]string{"/", "integer", val.Type.String()} {
		t.Fatalf("unexpected args: %#v", err.Args)
	}
}

func TestOverflowError(t *testing.T) {
	err := overflowError("*")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Category != verror.ErrMath {
		t.Fatalf("unexpected category: %v", err.Category)
	}
	if err.ID != verror.ErrIDOverflow {
		t.Fatalf("unexpected id: %s", err.ID)
	}
	if err.Args != [3]string{"*", "", ""} {
		t.Fatalf("unexpected args: %#v", err.Args)
	}
}

func TestUnderflowError(t *testing.T) {
	err := underflowError("-")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Category != verror.ErrMath {
		t.Fatalf("unexpected category: %v", err.Category)
	}
	if err.ID != verror.ErrIDUnderflow {
		t.Fatalf("unexpected id: %s", err.ID)
	}
	if err.Args != [3]string{"-", "", ""} {
		t.Fatalf("unexpected args: %#v", err.Args)
	}
}
