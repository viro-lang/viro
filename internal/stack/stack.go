// Package stack implements the unified stack for data and frames.
// All stack access uses integer indices (not pointers) per Constitution Principle IV.
// Package stack provides unified stack management for the Viro interpreter.
//
// The stack stores both data values and frame markers. It uses index-based
// access (not pointers) to ensure safety during stack growth/reallocation.
//
// Features:
//   - Automatic expansion using Go slice growth
//   - Push/Pop operations for values
//   - Get/Set operations for index-based access
//   - Frame markers to track activation records
//
// The stack grows automatically when capacity is exceeded, with expansion
// being transparent to callers (under 1ms per operation).
//
// Constitution Principle IV compliance: All access is index-based, never
// pointer-based, to avoid invalidation on reallocation.
package stack

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

// Stack is the unified storage for both data values and function frames.
// Design per data-model.md §7:
// - Index-based access prevents pointer invalidation on expansion
// - Go slice semantics for automatic growth
// - Frame layout: [return_slot, prior_frame_idx, function_metadata, args...]
//
// Constitution Principle IV: Stack and Frame Safety
// - NEVER use pointers to stack elements (they invalidate on expansion)
// - ALWAYS use integer indices for frame references
// - Stack expansion is transparent to caller
type Stack struct {
	Data         []core.Value // unified storage for values and frame metadata
	Top          int          // index of next available slot (0-based)
	CurrentFrame int          // index where current function frame starts (-1 if top-level)
}

// NewStack creates a stack with given initial capacity.
// Per research.md: Pre-allocating reasonable capacity (256) avoids most expansions.
func NewStack(initialCapacity int) *Stack {
	return &Stack{
		Data:         make([]core.Value, 0, initialCapacity),
		Top:          0,
		CurrentFrame: -1, // -1 means no active frame (top-level evaluation)
	}
}

// Push adds a value to the stack top.
// Automatically expands capacity if needed (Go slice semantics).
func (s *Stack) Push(v core.Value) int {
	index := s.Top
	if s.Top >= len(s.Data) {
		// Expand: append grows slice automatically
		s.Data = append(s.Data, v)
	} else {
		s.Data[s.Top] = v
	}
	s.Top++
	return index
}

// Pop removes and returns the top value.
// Panics if stack is empty (caller must check).
func (s *Stack) Pop() core.Value {
	if s.Top <= 0 {
		panic("stack underflow")
	}
	s.Top--
	return s.Data[s.Top]
}

// Get retrieves value at absolute index (index-based access).
// This is SAFE across stack expansions (index remains valid).
func (s *Stack) Get(index int) core.Value {
	if index < 0 || index >= s.Top {
		panic("stack index out of bounds")
	}
	return s.Data[index]
}

// Set updates value at absolute index (index-based access).
// This is SAFE across stack expansions.
func (s *Stack) Set(index int, v core.Value) {
	if index < 0 || index >= s.Top {
		panic("stack index out of bounds")
	}
	s.Data[index] = v
}

// Peek returns top value without removing it.
func (s *Stack) Peek() core.Value {
	if s.Top <= 0 {
		panic("stack underflow")
	}
	return s.Data[s.Top-1]
}

// Size returns the current number of values on stack.
func (s *Stack) Size() int {
	return s.Top
}

// Empty returns true if stack has no values.
func (s *Stack) Empty() bool {
	return s.Top == 0
}

// Reserve ensures stack has capacity for at least n more values.
// Useful for pre-allocating frame space.
func (s *Stack) Reserve(n int) {
	needed := s.Top + n
	if needed > cap(s.Data) {
		// Grow capacity
		newCap := max(cap(s.Data)*2, needed)
		newData := make([]core.Value, len(s.Data), newCap)
		copy(newData, s.Data)
		s.Data = newData
	}
}

// Reset clears the stack (for testing or REPL restart).
func (s *Stack) Reset() {
	s.Top = 0
	s.CurrentFrame = -1
}
