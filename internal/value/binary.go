package value

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
