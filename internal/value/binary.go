package value

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type BinaryValue struct {
	data  []byte
	index int
}

func NewBinaryValue(data []byte) *BinaryValue {
	return &BinaryValue{
		data:  data,
		index: 0,
	}
}

func (b *BinaryValue) Bytes() []byte {
	return b.data
}

func (b *BinaryValue) String() string {
	return b.Mold()
}

func (b *BinaryValue) formatHex(maxBytes int, showEllipsis bool) string {
	if len(b.data) == 0 || b.index >= len(b.data) {
		return "#{}"
	}

	visibleData := b.data[b.index:]
	bytesToFormat := len(visibleData)
	if maxBytes > 0 && bytesToFormat > maxBytes {
		bytesToFormat = maxBytes
	}

	var builder strings.Builder
	builder.WriteString("#{")

	for i := 0; i < bytesToFormat; i++ {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprintf("%02X", visibleData[i]))
	}

	if showEllipsis && bytesToFormat < len(visibleData) {
		builder.WriteString(fmt.Sprintf(" ... (%d bytes)", len(visibleData)))
	}

	builder.WriteString("}")
	return builder.String()
}

func (b *BinaryValue) Mold() string {
	return b.formatHex(0, false)
}

func (b *BinaryValue) Form() string {
	if b.index >= len(b.data) {
		return "#{}"
	}
	visibleLength := len(b.data) - b.index
	if visibleLength <= 64 {
		return b.Mold()
	}
	return b.formatHex(8, true)
}

func (b *BinaryValue) EqualsBinary(other *BinaryValue) bool {
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

func (b *BinaryValue) Equals(other core.Value) bool {
	if other.GetType() != TypeBinary {
		return false
	}
	otherBin, ok := other.(*BinaryValue)
	if !ok {
		return false
	}
	return b.EqualsBinary(otherBin)
}

func (b *BinaryValue) GetType() core.ValueType {
	return TypeBinary
}

func (b *BinaryValue) GetPayload() any {
	return b
}

func (b *BinaryValue) First() byte {
	return b.data[0]
}

func (b *BinaryValue) Last() byte {
	return b.data[len(b.data)-1]
}

func (b *BinaryValue) At(index int) byte {
	return b.data[index]
}

func (b *BinaryValue) ElementAt(index int) core.Value {
	return NewIntVal(int64(b.At(index)))
}

func (b *BinaryValue) Length() int {
	return len(b.data)
}

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

	b.data = append(toInsert, b.data...)
}

func (b *BinaryValue) Remove(count int) {
	b.data = append(b.data[:b.index], b.data[b.index+count:]...)
}

func (b *BinaryValue) Clone() Series {
	dataCopy := make([]byte, len(b.data))
	copy(dataCopy, b.data)
	return &BinaryValue{
		data:  dataCopy,
		index: b.index,
	}
}

func (b *BinaryValue) GetIndex() int {
	return b.index
}

func (b *BinaryValue) SetIndex(index int) {
	b.index = index
}

func (b *BinaryValue) FirstValue() (core.Value, error) {
	if len(b.data) == 0 {
		return NewNoneVal(), errors.New("empty series: first element")
	}
	if b.index >= len(b.data) {
		return NewNoneVal(), fmt.Errorf("out of bounds: %d >= %d", b.index, len(b.data))
	}
	return NewIntVal(int64(b.data[b.index])), nil
}

func (b *BinaryValue) LastValue() (core.Value, error) {
	if len(b.data) == 0 {
		return NewNoneVal(), errors.New("empty series: last element")
	}
	return NewIntVal(int64(b.Last())), nil
}

func (b *BinaryValue) AppendValue(val core.Value) error {
	switch val.GetType() {
	case TypeInteger:
		intVal, _ := AsIntValue(val)
		if intVal < 0 || intVal > 255 {
			return fmt.Errorf("out of bounds: %d not in range 0-255", intVal)
		}
		b.Append(byte(intVal))
	case TypeBinary:
		appendBin, _ := AsBinaryValue(val)
		b.Append(appendBin)
	default:
		return fmt.Errorf("type mismatch: expected integer or binary, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (b *BinaryValue) InsertValue(val core.Value) error {
	switch val.GetType() {
	case TypeInteger:
		intVal, _ := AsIntValue(val)
		if intVal < 0 || intVal > 255 {
			return fmt.Errorf("out of bounds: %d not in range 0-255", intVal)
		}
		b.SetIndex(0)
		b.Insert(byte(intVal))
	case TypeBinary:
		insertBin, _ := AsBinaryValue(val)
		b.SetIndex(0)
		b.Insert(insertBin)
	default:
		return fmt.Errorf("type mismatch: expected integer or binary, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (b *BinaryValue) CopyPart(count int) (Series, error) {
	clampedCount := ClampToRemaining(b.index, len(b.data), count)
	dataCopy := make([]byte, clampedCount)
	copy(dataCopy, b.data[b.index:b.index+clampedCount])
	return NewBinaryValue(dataCopy), nil
}

func (b *BinaryValue) RemoveCount(count int) error {
	if count < 0 {
		return fmt.Errorf("out of bounds: %d must be non-negative", count)
	}
	if b.index+count > len(b.data) {
		return fmt.Errorf("out of bounds: index %d + count %d > length %d", b.index, count, len(b.data))
	}
	b.Remove(count)
	return nil
}

func (b *BinaryValue) SkipBy(count int) {
	newIndex := b.index + count
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(b.data) {
		newIndex = len(b.data)
	}
	b.SetIndex(newIndex)
}

func (b *BinaryValue) TakeCount(count int) Series {
	if count > b.Length()-b.index {
		count = b.Length() - b.index
	}
	end := b.index + count
	if end > len(b.data) {
		end = len(b.data)
	}
	dataCopy := make([]byte, count)
	copy(dataCopy, b.data[b.index:end])
	return NewBinaryValue(dataCopy)
}

func (b *BinaryValue) ChangeValue(val core.Value) error {
	switch val.GetType() {
	case TypeInteger:
		intVal, _ := AsIntValue(val)
		if intVal < 0 || intVal > 255 {
			return fmt.Errorf("out of bounds: %d not in range 0-255", intVal)
		}
		if b.index >= len(b.data) {
			return fmt.Errorf("out of bounds: index %d >= length %d", b.index, len(b.data))
		}
		b.data[b.index] = byte(intVal)
	default:
		return fmt.Errorf("type mismatch: expected integer, got %s", TypeToString(val.GetType()))
	}
	return nil
}

func (b *BinaryValue) ClearSeries() {
	b.data = []byte{}
	b.index = 0
}

func SortBinary(b *BinaryValue) {
	sort.SliceStable(b.data, func(i, j int) bool {
		return b.data[i] < b.data[j]
	})
}
