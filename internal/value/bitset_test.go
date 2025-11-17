package value

import (
	"testing"
)

func TestBitset_NewAndSet(t *testing.T) {
	bs := NewBitsetValue()
	if !bs.IsEmpty() {
		t.Error("New bitset should be empty")
	}

	bs.Set('a')
	if !bs.Test('a') {
		t.Error("Character 'a' should be set")
	}
	if bs.Test('b') {
		t.Error("Character 'b' should not be set")
	}
	if bs.IsEmpty() {
		t.Error("Bitset should not be empty after setting a character")
	}
}

func TestBitset_FromString(t *testing.T) {
	bs := NewBitsetFromString("abc")
	if !bs.Test('a') || !bs.Test('b') || !bs.Test('c') {
		t.Error("Characters a, b, c should all be set")
	}
	if bs.Test('d') {
		t.Error("Character 'd' should not be set")
	}
	if bs.Count() != 3 {
		t.Errorf("Count should be 3, got %d", bs.Count())
	}
}

func TestBitset_FromRange(t *testing.T) {
	bs := NewBitsetFromRange('a', 'z')
	for r := 'a'; r <= 'z'; r++ {
		if !bs.Test(r) {
			t.Errorf("Character %c should be set", r)
		}
	}
	if bs.Test('A') {
		t.Error("Character 'A' should not be set")
	}
	if bs.Count() != 26 {
		t.Errorf("Count should be 26, got %d", bs.Count())
	}
}

func TestBitset_Clear(t *testing.T) {
	bs := NewBitsetFromString("abc")
	bs.Clear('b')
	if bs.Test('b') {
		t.Error("Character 'b' should be cleared")
	}
	if !bs.Test('a') || !bs.Test('c') {
		t.Error("Characters 'a' and 'c' should still be set")
	}
	if bs.Count() != 2 {
		t.Errorf("Count should be 2, got %d", bs.Count())
	}
}

func TestBitset_Unicode(t *testing.T) {
	bs := NewBitsetValue()
	bs.Set('世') // Chinese character
	bs.Set('界')
	if !bs.Test('世') || !bs.Test('界') {
		t.Error("Unicode characters should be set")
	}
	if bs.Test('a') {
		t.Error("Character 'a' should not be set")
	}
	if bs.Count() != 2 {
		t.Errorf("Count should be 2, got %d", bs.Count())
	}
}

func TestBitset_Clone(t *testing.T) {
	bs1 := NewBitsetFromString("abc")
	bs2 := bs1.Clone()

	if !bs2.Test('a') || !bs2.Test('b') || !bs2.Test('c') {
		t.Error("Cloned bitset should have same characters")
	}

	// Modify original
	bs1.Set('d')
	if bs2.Test('d') {
		t.Error("Cloned bitset should not be affected by changes to original")
	}
}

func TestBitset_Union(t *testing.T) {
	bs1 := NewBitsetFromString("abc")
	bs2 := NewBitsetFromString("bcd")
	bs3 := bs1.Union(bs2)

	expected := map[rune]bool{'a': true, 'b': true, 'c': true, 'd': true}
	for r := range expected {
		if !bs3.Test(r) {
			t.Errorf("Union should contain %c", r)
		}
	}
	if bs3.Count() != 4 {
		t.Errorf("Union count should be 4, got %d", bs3.Count())
	}
}

func TestBitset_Intersect(t *testing.T) {
	bs1 := NewBitsetFromString("abc")
	bs2 := NewBitsetFromString("bcd")
	bs3 := bs1.Intersect(bs2)

	if !bs3.Test('b') || !bs3.Test('c') {
		t.Error("Intersection should contain 'b' and 'c'")
	}
	if bs3.Test('a') || bs3.Test('d') {
		t.Error("Intersection should not contain 'a' or 'd'")
	}
	if bs3.Count() != 2 {
		t.Errorf("Intersection count should be 2, got %d", bs3.Count())
	}
}

func TestBitset_Complement(t *testing.T) {
	bs := NewBitsetFromString("abc")
	comp := bs.Complement()

	if comp.Test('a') || comp.Test('b') || comp.Test('c') {
		t.Error("Complement should not contain 'a', 'b', or 'c'")
	}
	if !comp.Test('d') || !comp.Test('z') {
		t.Error("Complement should contain other ASCII characters")
	}
}

func TestBitset_GetChars(t *testing.T) {
	bs := NewBitsetFromString("cab")
	chars := bs.GetChars()

	expected := []rune{'a', 'b', 'c'}
	if len(chars) != len(expected) {
		t.Errorf("GetChars should return %d chars, got %d", len(expected), len(chars))
	}
	for i, r := range expected {
		if chars[i] != r {
			t.Errorf("GetChars[%d] should be %c, got %c", i, r, chars[i])
		}
	}
}

func TestBitset_ValueInterface(t *testing.T) {
	bs := NewBitsetFromString("abc")
	if bs.GetType() != TypeBitset {
		t.Errorf("GetType should return TypeBitset, got %v", bs.GetType())
	}

	str := bs.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}

	mold := bs.Mold()
	if mold == "" {
		t.Error("Mold() should return non-empty string")
	}
}

func TestBitset_Equals(t *testing.T) {
	bs1 := NewBitsetFromString("abc")
	bs2 := NewBitsetFromString("abc")
	bs3 := NewBitsetFromString("abd")

	if !bs1.Equals(bs2) {
		t.Error("Bitsets with same characters should be equal")
	}
	if bs1.Equals(bs3) {
		t.Error("Bitsets with different characters should not be equal")
	}
}

func TestBitset_MoldFormat(t *testing.T) {
	// Test empty bitset
	bs := NewBitsetValue()
	if bs.Mold() != "charset []" {
		t.Errorf("Empty bitset mold should be 'charset []', got %s", bs.Mold())
	}

	// Test single character
	bs = NewBitsetFromString("a")
	mold := bs.Mold()
	if mold != "charset [#\"a\"]" {
		t.Errorf("Single char mold incorrect: %s", mold)
	}

	// Test range (should use range notation)
	bs = NewBitsetFromRange('a', 'c')
	mold = bs.Mold()
	if mold != "charset [[#\"a\" - #\"c\"]]" {
		t.Errorf("Range mold incorrect: %s", mold)
	}
}

func TestBitset_Form(t *testing.T) {
	// Small charset shows all chars
	bs := NewBitsetFromString("abc")
	form := bs.Form()
	if !contains(form, "charset") {
		t.Errorf("Form should contain 'charset', got: %s", form)
	}

	// Large charset shows count
	bs = NewBitsetFromRange('a', 'z')
	form = bs.Form()
	if !contains(form, "26 characters") {
		t.Errorf("Large charset form should show count, got: %s", form)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
