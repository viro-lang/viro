package stack

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// Frame layout on stack (per data-model.md §7):
// [frameBase]     Return value slot (initially none)
// [frameBase+1]   Prior frame index (int, -1 if none)
// [frameBase+2]   Function metadata (FunctionValue)
// [frameBase+3]   Argument 1
// [frameBase+4]   Argument 2
// ...             Additional arguments
// [frameBase+3+N] Local variables (grows as needed)

// FrameLayout constants for accessing frame components.
const (
	FrameOffsetReturn   = 0 // Return value slot
	FrameOffsetPrior    = 1 // Prior frame index
	FrameOffsetFunction = 2 // Function metadata
	FrameOffsetArgs     = 3 // Start of arguments
)

// NewFrame allocates a new function frame on the stack.
// Returns the frame base index.
//
// Layout created:
// - Return slot: initialized to none
// - Prior frame: current frame index (or -1)
// - Function: the function being called
// - Args: space for argCount arguments (caller will fill)
func (s *Stack) NewFrame(fn *value.FunctionValue, argCount int) int {
	frameBase := s.Top

	// Allocate frame slots: return + prior + function + args
	frameSize := FrameOffsetArgs + argCount
	s.Reserve(frameSize)

	// Initialize frame layout
	s.Push(value.NoneVal())                     // Return slot
	s.Push(value.IntVal(int64(s.CurrentFrame))) // Prior frame index
	s.Push(value.FuncVal(fn))                   // Function metadata

	// Push placeholder none values for arguments (caller will set)
	for range argCount {
		s.Push(value.NoneVal())
	}

	// Update current frame pointer
	s.CurrentFrame = frameBase

	return frameBase
}

// DestroyFrame unwinds the current frame and restores prior frame.
// Returns the return value from the frame.
func (s *Stack) DestroyFrame() core.Value {
	if s.CurrentFrame == -1 {
		panic("no active frame to destroy")
	}

	// Get return value before unwinding
	returnVal := s.Get(s.CurrentFrame + FrameOffsetReturn)

	// Get prior frame index
	priorFrameVal := s.Get(s.CurrentFrame + FrameOffsetPrior)
	priorFrame, ok := value.AsInteger(priorFrameVal)
	if !ok {
		panic("corrupted frame: invalid prior frame index")
	}

	// Unwind stack to frame base
	s.Top = s.CurrentFrame

	// Restore prior frame as current
	s.CurrentFrame = int(priorFrame)

	return returnVal
}

// GetFrameArg retrieves an argument from the current frame.
// argIndex is 0-based (0 = first argument).
func (s *Stack) GetFrameArg(argIndex int) core.Value {
	if s.CurrentFrame == -1 {
		panic("no active frame")
	}
	return s.Get(s.CurrentFrame + FrameOffsetArgs + argIndex)
}

// SetFrameArg sets an argument in the current frame.
// argIndex is 0-based.
func (s *Stack) SetFrameArg(argIndex int, v core.Value) {
	if s.CurrentFrame == -1 {
		panic("no active frame")
	}
	s.Set(s.CurrentFrame+FrameOffsetArgs+argIndex, v)
}

// GetFrameReturn retrieves the return value slot.
func (s *Stack) GetFrameReturn() core.Value {
	if s.CurrentFrame == -1 {
		panic("no active frame")
	}
	return s.Get(s.CurrentFrame + FrameOffsetReturn)
}

// SetFrameReturn sets the return value for the current frame.
func (s *Stack) SetFrameReturn(v core.Value) {
	if s.CurrentFrame == -1 {
		panic("no active frame")
	}
	s.Set(s.CurrentFrame+FrameOffsetReturn, v)
}

// GetFrameFunction retrieves the function metadata from current frame.
func (s *Stack) GetFrameFunction() *value.FunctionValue {
	if s.CurrentFrame == -1 {
		panic("no active frame")
	}
	fnVal := s.Get(s.CurrentFrame + FrameOffsetFunction)
	fn, ok := value.AsFunction(fnVal)
	if !ok {
		panic("corrupted frame: invalid function metadata")
	}
	return fn
}

// GetFrameBase returns the base index of the current frame.
func (s *Stack) GetFrameBase() int {
	return s.CurrentFrame
}

// HasFrame returns true if there's an active frame.
func (s *Stack) HasFrame() bool {
	return s.CurrentFrame != -1
}

// CaptureCallStack walks the frame chain and returns function names.
// Used for error reporting (Where context).
func (s *Stack) CaptureCallStack() []string {
	var calls []string
	frameIdx := s.CurrentFrame

	for frameIdx != -1 {
		// Get function from frame
		fnVal := s.Get(frameIdx + FrameOffsetFunction)
		if fn, ok := value.AsFunction(fnVal); ok {
			calls = append(calls, fn.Name)
		}

		// Get prior frame
		priorVal := s.Get(frameIdx + FrameOffsetPrior)
		if priorInt, ok := value.AsInteger(priorVal); ok {
			frameIdx = int(priorInt)
		} else {
			break
		}
	}

	return calls
}
