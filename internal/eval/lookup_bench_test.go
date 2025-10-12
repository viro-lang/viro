package eval

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// BenchmarkLookupNative measures native function lookup from the top level.
// Natives are stored in the root frame, so this tests root frame lookup performance.
func BenchmarkLookupNative(b *testing.B) {
	e := NewEvaluator()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, ok := e.Lookup("+")
		if !ok {
			b.Fatal("native + not found")
		}
	}
}

// BenchmarkLookupNativeFromNested measures native function lookup from a 3-level nested scope.
// This tests frame chain traversal performance when looking up a native from deep in the call stack.
func BenchmarkLookupNativeFromNested(b *testing.B) {
	e := NewEvaluator()

	// Create 3 nested frames
	frame1 := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, 0, 5)
	frame1.Name = "level1"
	frame1.Bind("x", value.IntVal(1))
	frame1Idx := e.pushFrame(frame1)

	frame2 := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, frame1Idx, 5)
	frame2.Name = "level2"
	frame2.Bind("y", value.IntVal(2))
	frame2Idx := e.pushFrame(frame2)

	frame3 := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, frame2Idx, 5)
	frame3.Name = "level3"
	frame3.Bind("z", value.IntVal(3))
	e.pushFrame(frame3)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, ok := e.Lookup("+")
		if !ok {
			b.Fatal("native + not found from nested scope")
		}
	}
}

// BenchmarkLookupUserDefined measures local variable lookup (best case).
// This tests lookup when the word is found in the current frame.
func BenchmarkLookupUserDefined(b *testing.B) {
	e := NewEvaluator()

	// Create a frame with a local variable
	userFrame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, 0, 10)
	userFrame.Name = "userFunc"
	userFrame.Bind("myVar", value.IntVal(42))
	userFrame.Bind("x", value.StrVal("test"))
	userFrame.Bind("y", value.LogicVal(true))
	e.pushFrame(userFrame)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, ok := e.Lookup("myVar")
		if !ok {
			b.Fatal("user variable myVar not found")
		}
	}
}

// BenchmarkLookupUserDefinedFromParent measures lookup of a variable in the parent frame.
// This tests frame chain traversal for user-defined words.
func BenchmarkLookupUserDefinedFromParent(b *testing.B) {
	e := NewEvaluator()

	// Create parent frame with a variable
	parentFrame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, 0, 10)
	parentFrame.Name = "parent"
	parentFrame.Bind("parentVar", value.IntVal(100))
	parentIdx := e.pushFrame(parentFrame)

	// Create child frame without the variable
	childFrame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parentIdx, 5)
	childFrame.Name = "child"
	childFrame.Bind("childVar", value.IntVal(200))
	e.pushFrame(childFrame)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, ok := e.Lookup("parentVar")
		if !ok {
			b.Fatal("parent variable not found")
		}
	}
}
