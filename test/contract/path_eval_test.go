package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestEvalPathSegmentsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantType core.ValueType
		want     string
	}{
		{
			name: "path word result",
			code: `field: 'profile
obj: object [profile: object [name: "Alice"]]
obj.(field).name`,
			wantType: value.TypeString,
			want:     "\"Alice\"",
		},
		{
			name: "path string result",
			code: `field: "profile"
obj: object [profile: object [name: "Eve"]]
obj.(field).name`,
			wantType: value.TypeString,
			want:     "\"Eve\"",
		},
		{
			name: "path index result",
			code: `idx: 2
data: [10 20 30]
data.(idx)`,
			wantType: value.TypeInteger,
			want:     "20",
		},
		{
			name: "get-path eval segment",
			code: `field: 'profile
obj: object [profile: object [city: "Portland"]]
:obj.(field).city`,
			wantType: value.TypeString,
			want:     "\"Portland\"",
		},
		{
			name: "set-path eval segment",
			code: `field: 'profile
obj: object [profile: object [name: "Alice"]]
obj.(field).name: "Bob"
obj.profile.name`,
			wantType: value.TypeString,
			want:     "\"Bob\"",
		},
		{
			name: "nested eval chain",
			code: `outer: 'profile
inner: 'name
obj: object [profile: object [name: "Zoe"]]
obj.(outer).(inner)`,
			wantType: value.TypeString,
			want:     "\"Zoe\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.GetType() != tt.wantType {
				t.Fatalf("got type %s, want %s", value.TypeToString(result.GetType()), value.TypeToString(tt.wantType))
			}
			if result.Mold() != tt.want {
				t.Fatalf("got %s, want %s", result.Mold(), tt.want)
			}
		})
	}
}

func TestSetPathEvalSegmentEvaluatesOnce(t *testing.T) {
	code := `state: object [counter: 0]
next-index: fn [] [
    state.counter: state.counter + 1
    state.counter
]
data: [10 20]
data.(next-index): 99
state.counter`

	result, err := Evaluate(code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GetType() != value.TypeInteger {
		t.Fatalf("got type %s, want integer", value.TypeToString(result.GetType()))
	}
	if result.Mold() != "1" {
		t.Fatalf("expected counter to equal 1, got %s", result.Mold())
	}
}

func TestEvalPathNestedEvalSegmentsEvaluateOnce(t *testing.T) {
	t.Run("get-path nested indexes", func(t *testing.T) {
		code := `state: object [outer: 0 inner: 0]
next-outer: fn [] [
    state.outer: state.outer + 1
    state.outer
]
next-inner: fn [] [
    state.inner: state.inner + 1
    state.inner
]
matrix: [[10 20] [30 40]]
matrix.(next-outer).(next-inner)
state`
		result, err := Evaluate(code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		obj, ok := value.AsObject(result)
		if !ok {
			t.Fatalf("expected object, got %s", value.TypeToString(result.GetType()))
		}
		assertObjectFieldMold(t, obj, "outer", "1")
		assertObjectFieldMold(t, obj, "inner", "1")
	})

	t.Run("set-path nested indexes", func(t *testing.T) {
		code := `state: object [row: 0 col: 0]
next-row: fn [] [
    state.row: state.row + 1
    state.row
]
next-col: fn [] [
    state.col: state.col + 1
    state.col
]
grid: [[0 0] [0 0]]
grid.(next-row).(next-col): 42
state`
		result, err := Evaluate(code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		obj, ok := value.AsObject(result)
		if !ok {
			t.Fatalf("expected object, got %s", value.TypeToString(result.GetType()))
		}
		assertObjectFieldMold(t, obj, "row", "1")
		assertObjectFieldMold(t, obj, "col", "1")
	})
}

func TestEvalPathSegmentsErrors(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		errID string
	}{
		{
			name: "path decimal eval result",
			code: `field: 1.5
obj: object [profile: object [name: "Alice"]]
obj.(field).name`,
			errID: verror.ErrIDInvalidPath,
		},
		{
			name: "get-path decimal eval result",
			code: `field: 1.5
obj: object [profile: object [name: "Alice"]]
:obj.(field).name`,
			errID: verror.ErrIDInvalidPath,
		},
		{
			name: "set-path decimal eval result",
			code: `field: 1.5
obj: object [profile: object [name: "Alice"]]
obj.(field).name: "Bob"`,
			errID: verror.ErrIDInvalidPath,
		},
		{
			name: "set-path out of bounds",
			code: `idx: 5
data: [10 20]
data.(idx): 1`,
			errID: verror.ErrIDOutOfBounds,
		},
		{
			name: "path through none",
			code: `field: 'slot
obj: object [slot: none]
obj.(field).name`,
			errID: verror.ErrIDNonePath,
		},
		{
			name: "set-path through none",
			code: `field: 'slot
obj: object [slot: none]
obj.(field).name: "Bob"`,
			errID: verror.ErrIDNonePath,
		},
		{
			name: "get-path block eval result",
			code: `field: []
obj: object [profile: object [name: "Alice"]]
:obj.(field).name`,
			errID: verror.ErrIDInvalidPath,
		},
		{
			name: "string result with dot treated literal",
			code: `field: "profile.extra"
obj: object [profile: object [extra: "Alice"]]
obj.(field)`,
			errID: verror.ErrIDNoSuchField,
		},
		{
			name: "string result with slash treated literal",
			code: `field: "profile/name"
obj: object [profile: object [name: "Alice"]]
obj.(field)`,
			errID: verror.ErrIDNoSuchField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectEvalPathError(t, tt.code, tt.errID)
		})
	}
}

func TestLeadingEvalSegmentsRejected(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		errID string
	}{
		{name: "path", code: ".(field).name", errID: verror.ErrIDPathEvalBase},
		{name: "get-path", code: ":.(field).name", errID: verror.ErrIDPathEvalBase},
		{name: "set-path", code: ".(field).name: 1", errID: verror.ErrIDPathEvalBase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if err == nil {
				t.Fatalf("expected syntax error %s", tt.errID)
			}
			verr, ok := err.(*verror.Error)
			if !ok {
				t.Fatalf("expected verror.Error, got %T", err)
			}
			if verr.ID != tt.errID {
				t.Fatalf("got error %s, want %s", verr.ID, tt.errID)
			}
			if verr.Category != verror.ErrSyntax {
				t.Fatalf("expected syntax error, got %v", verr.Category)
			}
		})
	}
}

func expectEvalPathError(t *testing.T, code, errID string) {
	t.Helper()
	_, err := Evaluate(code)
	if err == nil {
		t.Fatalf("expected error %s", errID)
	}
	verr, ok := err.(*verror.Error)
	if !ok {
		t.Fatalf("expected verror.Error, got %T", err)
	}
	if verr.ID != errID {
		t.Fatalf("got error %s, want %s", verr.ID, errID)
	}
}

func assertObjectFieldMold(t *testing.T, obj *value.ObjectInstance, field, want string) {
	t.Helper()
	got, ok := obj.GetField(field)
	if !ok {
		t.Fatalf("missing field %s", field)
	}
	if got.Mold() != want {
		t.Fatalf("field %s mold = %s, want %s", field, got.Mold(), want)
	}
}
