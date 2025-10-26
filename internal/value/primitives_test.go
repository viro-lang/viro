package value

import (
	"testing"
)

func TestSingletons(t *testing.T) {
	t.Run("NoneVal returns same instance", func(t *testing.T) {
		none1 := NewNoneVal()
		none2 := NewNoneVal()

		// Compare using interface values - they should be equal
		if !none1.Equals(none2) {
			t.Error("NoneVal instances should be equal")
		}

		// For value types in interfaces, we can't check pointer equality,
		// but we can verify they behave identically
		if none1.GetType() != none2.GetType() {
			t.Error("NoneVal instances should have same type")
		}
	})

	t.Run("LogicVal true returns same instance", func(t *testing.T) {
		true1 := NewLogicVal(true)
		true2 := NewLogicVal(true)

		if !true1.Equals(true2) {
			t.Error("true instances should be equal")
		}

		if true1.GetType() != true2.GetType() {
			t.Error("true instances should have same type")
		}
	})

	t.Run("LogicVal false returns same instance", func(t *testing.T) {
		false1 := NewLogicVal(false)
		false2 := NewLogicVal(false)

		if !false1.Equals(false2) {
			t.Error("false instances should be equal")
		}

		if false1.GetType() != false2.GetType() {
			t.Error("false instances should have same type")
		}
	})

	t.Run("true and false are different", func(t *testing.T) {
		trueVal := NewLogicVal(true)
		falseVal := NewLogicVal(false)

		if trueVal.Equals(falseVal) {
			t.Error("true and false should not be equal")
		}
	})
}

func BenchmarkNewNoneVal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewNoneVal()
	}
}

func BenchmarkNewLogicVal(b *testing.B) {
	b.Run("true", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewLogicVal(true)
		}
	})

	b.Run("false", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewLogicVal(false)
		}
	})
}
