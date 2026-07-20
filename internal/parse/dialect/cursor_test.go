package dialect

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestStringCursor(t *testing.T) {
	sc := NewStringCursor("hello")

	if sc.Length() != 5 {
		t.Errorf("Length() = %d, want 5", sc.Length())
	}

	val, ok := sc.At(0)
	if !ok || val.String() != "h" {
		t.Errorf("At(0) = %v, want 'h'", val.String())
	}

	val, ok = sc.At(10)
	if ok {
		t.Error("At(10) should return false")
	}

	slice := sc.Slice(1, 4)
	if strVal, ok := value.AsStringValue(slice); !ok || strVal.String() != "ell" {
		t.Errorf("Slice(1, 4) = %v, want 'ell'", slice.String())
	}
}

func TestBlockCursor(t *testing.T) {
	elements := []core.Value{
		value.NewIntVal(1),
		value.NewIntVal(2),
		value.NewIntVal(3),
	}
	bc := NewBlockCursor(elements)

	if bc.Length() != 3 {
		t.Errorf("Length() = %d, want 3", bc.Length())
	}

	val, ok := bc.At(0)
	if !ok {
		t.Fatal("At(0) should return true")
	}
	if intVal, ok := value.AsIntValue(val); !ok || intVal != 1 {
		t.Errorf("At(0) = %v, want 1", val.String())
	}

	val, ok = bc.At(10)
	if ok {
		t.Error("At(10) should return false")
	}

	slice := bc.Slice(1, 3)
	if blockVal, ok := value.AsBlockValue(slice); !ok || len(blockVal.Elements) != 2 {
		t.Errorf("Slice(1, 3) length = %d, want 2", len(blockVal.Elements))
	}
}

func TestMatchString(t *testing.T) {
	tests := []struct {
		s1            string
		s2            string
		caseSensitive bool
		want          bool
	}{
		{"hello", "hello", true, true},
		{"hello", "HELLO", true, false},
		{"hello", "HELLO", false, true},
		{"Hello", "hello", false, true},
		{"abc", "def", false, false},
	}

	for _, tt := range tests {
		got := MatchString(tt.s1, tt.s2, tt.caseSensitive)
		if got != tt.want {
			t.Errorf("MatchString(%q, %q, %v) = %v, want %v", tt.s1, tt.s2, tt.caseSensitive, got, tt.want)
		}
	}
}
