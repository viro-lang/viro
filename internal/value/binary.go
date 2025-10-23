package value

import "sort"

// BinaryValue represents a sequence of raw bytes.
type BinaryValue struct {
	data  []byte
	index int
}

// NewBinaryValue creates a BinaryValue from a byte slice.
func NewBinaryValue(data []byte) *BinaryValue {
	return &BinaryValue{
		data:  data,
		index: 0,
	}
}

// Bytes returns the underlying byte slice.
func (b *BinaryValue) Bytes() []byte {
	return b.data
}

// String converts BinaryValue to hex string representation.
func (b *BinaryValue) String() string {
	return b.Mold()
}

// Mold returns the mold-formatted binary representation.
func (b *BinaryValue) Mold() string {
	if len(b.data) == 0 {
		return "#{}"
	}
	return "#{" + string(b.data) + "}" // Simplified for now
}

// Equals performs deep equality comparison.
func (b *BinaryValue) Equals(other *BinaryValue) bool {
	if len(b.data) != len(other.data) {
		return false
	}
	for i := range b.data {
		if b.data[i] != other.data[i] {
			return false
		}
	}
	return true
}

// Series operations (contracts/series.md)

// First returns the first byte (error if empty handled by caller).
func (b *BinaryValue) First() byte {
	return b.data[0]
}

// Last returns the last byte (error if empty handled by caller).
func (b *BinaryValue) Last() byte {
	return b.data[len(b.data)-1]
}

// At returns byte at index (bounds checking by caller).
func (b *BinaryValue) At(index int) byte {
	return b.data[index]
}

// Length returns byte count.
func (b *BinaryValue) Length() int {
	return len(b.data)
}

// Append adds a byte or binary value to the end (in-place mutation).
func (b *BinaryValue) Append(val interface{}) {
	switch v := val.(type) {
	case byte:
		b.data = append(b.data, v)
	case []byte:
		b.data = append(b.data, v...)
	case *BinaryValue:
		b.data = append(b.data, v.data...)
	}
}

// Insert adds a byte or binary value at current position (in-place mutation).
func (b *BinaryValue) Insert(val interface{}) {
	var toInsert []byte
	switch v := val.(type) {
	case byte:
		toInsert = []byte{v}
	case []byte:
		toInsert = v
	case *BinaryValue:
		toInsert = v.data
	}

	// Insert at current index
	b.data = append(b.data[:b.index], append(toInsert, b.data[b.index:]...)...)
}

// Remove removes a specified number of bytes from the current position (in-place mutation).
func (b *BinaryValue) Remove(count int) {
	if b.index+count <= len(b.data) {
		b.data = append(b.data[:b.index], b.data[b.index+count:]...)
	}
}

// GetIndex returns current series position.
func (b *BinaryValue) GetIndex() int {
	return b.index
}

// SetIndex sets the current series position.
func (b *BinaryValue) SetIndex(index int) {
	b.index = index
}

// SortBinary sorts the bytes in the binary value in ascending order.
func SortBinary(b *BinaryValue) {
	sort.SliceStable(b.data, func(i, j int) bool {
		return b.data[i] < b.data[j]
	})
}
