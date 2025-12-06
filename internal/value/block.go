package value

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

type BlockValue struct {
	Elements  []core.Value
	Index     int
	typ       core.ValueType
	locations []core.SourceLocation
}

func NewBlockValue(elements []core.Value) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements:  elements,
		Index:     0,
		typ:       TypeBlock,
		locations: make([]core.SourceLocation, len(elements)),
	}
}

func NewBlockValueWithType(elements []core.Value, typ core.ValueType) *BlockValue {
	if elements == nil {
		elements = []core.Value{}
	}
	return &BlockValue{
		Elements:  elements,
		Index:     0,
		typ:       typ,
		locations: make([]core.SourceLocation, len(elements)),
	}
}

func (b *BlockValue) ensureLocationCapacity() {
	if len(b.locations) != len(b.Elements) {
		newLocations := make([]core.SourceLocation, len(b.Elements))
		copy(newLocations, b.locations)
		b.locations = newLocations
	}
}

func (b *BlockValue) SetLocations(locations []core.SourceLocation) {
	b.locations = make([]core.SourceLocation, len(b.Elements))
	copy(b.locations, locations)
}

func (b *BlockValue) SetLocationAt(index int, location core.SourceLocation) {
	if index < 0 || index >= len(b.Elements) {
		return
	}
	b.ensureLocationCapacity()
	b.locations[index] = location
}

func (b *BlockValue) LocationAt(index int) core.SourceLocation {
	if index < 0 || index >= len(b.locations) {
		return core.SourceLocation{}
	}
	return b.locations[index]
}

func (b *BlockValue) Locations() []core.SourceLocation {
	if len(b.locations) == 0 {
		return nil
	}
	return b.locations
}

func (b *BlockValue) String() string {
	return "[" + b.StringElements() + "]"
}

func (b *BlockValue) Mold() string {
	if len(b.Elements) == 0 {
		return "[]"
	}
	if b.Index >= len(b.Elements) {
		return "[]"
	}
	visibleElements := b.Elements[b.Index:]
	parts := make([]string, len(visibleElements))
	for i, elem := range visibleElements {
		parts[i] = elem.Mold()
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func (b *BlockValue) Form() string {
	if len(b.Elements) == 0 {
		return ""
	}
	if b.Index >= len(b.Elements) {
		return ""
	}
	visibleElements := b.Elements[b.Index:]
	parts := make([]string, len(visibleElements))
	for i, elem := range visibleElements {
		parts[i] = elem.Form()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) MoldElements() string {
	if len(b.Elements) == 0 {
		return ""
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
		parts[i] = elem.Mold()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) StringElements() string {
	if len(b.Elements) == 0 {
		return ""
	}
	parts := make([]string, len(b.Elements))
	for i, elem := range b.Elements {
		parts[i] = elem.Form()
	}
	return strings.Join(parts, " ")
}

func (b *BlockValue) EqualsBlock(other *BlockValue) bool {
	if len(b.Elements) != len(other.Elements) {
		return false
	}
	for i := range b.Elements {
		if !b.Elements[i].Equals(other.Elements[i]) {
			return false
		}
	}
	return true
}

func (b *BlockValue) Equals(other core.Value) bool {
	if other.GetType() != TypeBlock && other.GetType() != TypeParen {
		return false
	}
	otherBlock, ok := other.(*BlockValue)
	if !ok {
		return false
	}
	return b.EqualsBlock(otherBlock)
}

func (b *BlockValue) GetType() core.ValueType {
	return b.typ
}

func (b *BlockValue) GetPayload() any {
	return b
}

func (b *BlockValue) First() core.Value {
	return b.Elements[0]
}

func (b *BlockValue) Last() core.Value {
	return b.Elements[len(b.Elements)-1]
}

func (b *BlockValue) At(index int) core.Value {
	return b.Elements[index]
}

func (b *BlockValue) ElementAt(index int) core.Value {
	return b.At(index)
}

func (b *BlockValue) Length() int {
	return len(b.Elements)
}

func (b *BlockValue) Append(val core.Value) {
	b.ensureLocationCapacity()
	b.Elements = append(b.Elements, val)
	b.locations = append(b.locations, core.SourceLocation{})
}

func (b *BlockValue) Insert(val core.Value) {
	b.ensureLocationCapacity()
	b.Elements = append([]core.Value{val}, b.Elements...)
	b.locations = append([]core.SourceLocation{{}}, b.locations...)
}

func (b *BlockValue) Remove(count int) {
	b.ensureLocationCapacity()
	b.Elements = append(b.Elements[:b.Index], b.Elements[b.Index+count:]...)
	if len(b.locations) >= b.Index+count {
		b.locations = append(b.locations[:b.Index], b.locations[b.Index+count:]...)
	} else if len(b.locations) > b.Index {
		b.locations = b.locations[:b.Index]
	}
}

func (b *BlockValue) GetIndex() int {
	return b.Index
}

func (b *BlockValue) SetIndex(idx int) {
	b.Index = idx
}

func (b *BlockValue) Clone() Series {
	elemsCopy := make([]core.Value, len(b.Elements))
	copy(elemsCopy, b.Elements)
	locCopy := make([]core.SourceLocation, len(b.Elements))
	copy(locCopy, b.locations)
	return &BlockValue{
		Elements:  elemsCopy,
		Index:     b.Index,
		typ:       b.typ,
		locations: locCopy,
	}
}

func (b *BlockValue) GoString() string {
	return fmt.Sprintf("Block{Elements: %d, Index: %d}", len(b.Elements), b.Index)
}

func (b *BlockValue) FirstValue() (core.Value, error) {
	if len(b.Elements) == 0 {
		return NewNoneVal(), errors.New("empty series: first element")
	}
	if b.Index >= len(b.Elements) {
		return NewNoneVal(), fmt.Errorf("out of bounds: %d >= %d", b.Index, len(b.Elements))
	}
	return b.Elements[b.Index], nil
}

func (b *BlockValue) LastValue() (core.Value, error) {
	if len(b.Elements) == 0 {
		return NewNoneVal(), errors.New("empty series: last element")
	}
	return b.Last(), nil
}

func (b *BlockValue) AppendValue(val core.Value) error {
	b.Append(val)
	return nil
}

func (b *BlockValue) InsertValue(val core.Value) error {
	b.SetIndex(0)
	b.Insert(val)
	return nil
}

func (b *BlockValue) CopyPart(count int) (Series, error) {
	b.ensureLocationCapacity()
	clampedCount := ClampToRemaining(b.Index, len(b.Elements), count)
	elemsCopy := make([]core.Value, clampedCount)
	copy(elemsCopy, b.Elements[b.Index:b.Index+clampedCount])
	copyBlock := NewBlockValue(elemsCopy)
	if clampedCount > 0 {
		locCopy := make([]core.SourceLocation, clampedCount)
		copy(locCopy, b.locations[b.Index:b.Index+clampedCount])
		copyBlock.SetLocations(locCopy)
	}
	return copyBlock, nil
}

func (b *BlockValue) RemoveCount(count int) error {
	if count < 0 {
		return fmt.Errorf("out of bounds: %d must be non-negative", count)
	}
	if b.Index+count > len(b.Elements) {
		return fmt.Errorf("out of bounds: index %d + count %d > length %d", b.Index, count, len(b.Elements))
	}
	b.Remove(count)
	return nil
}

func (b *BlockValue) SkipBy(count int) {
	newIndex := b.Index + count
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(b.Elements) {
		newIndex = len(b.Elements)
	}
	b.SetIndex(newIndex)
}

func (b *BlockValue) TakeCount(count int) Series {
	b.ensureLocationCapacity()
	if count > b.Length()-b.Index {
		count = b.Length() - b.Index
	}
	end := b.Index + count
	if end > len(b.Elements) {
		end = len(b.Elements)
	}
	elemsCopy := make([]core.Value, count)
	copy(elemsCopy, b.Elements[b.Index:end])
	taken := NewBlockValue(elemsCopy)
	if count > 0 {
		locCopy := make([]core.SourceLocation, count)
		copy(locCopy, b.locations[b.Index:end])
		taken.SetLocations(locCopy)
	}
	return taken
}

func (b *BlockValue) ChangeValue(val core.Value) error {
	if b.Index >= len(b.Elements) {
		return fmt.Errorf("out of bounds: index %d >= length %d", b.Index, len(b.Elements))
	}
	b.Elements[b.Index] = val
	return nil
}

func (b *BlockValue) ClearSeries() {
	b.Elements = []core.Value{}
	b.Index = 0
	b.locations = []core.SourceLocation{}
}

// compareBlockValuesLex performs lexicographic comparison of two block values
func compareBlockValuesLex(a, b *BlockValue) int {
	minLen := len(a.Elements)
	if len(b.Elements) < minLen {
		minLen = len(b.Elements)
	}

	for i := 0; i < minLen; i++ {
		cmp := compareValuesForSort(a.Elements[i], b.Elements[i])
		if cmp != 0 {
			return cmp
		}
	}

	// All compared elements are equal, shorter block comes first
	if len(a.Elements) < len(b.Elements) {
		return -1
	} else if len(a.Elements) > len(b.Elements) {
		return 1
	}
	return 0
}

// getTypePrecedence returns a precedence value for sorting different types
// Lower values sort before higher values
func getTypePrecedence(t core.ValueType) int {
	switch t {
	case TypeInteger:
		return 1
	case TypeDecimal:
		return 2
	case TypeString:
		return 3
	case TypeBinary:
		return 4
	case TypeLogic:
		return 5
	case TypeWord, TypeSetWord, TypeGetWord, TypeLitWord:
		return 6
	case TypePath, TypeGetPath, TypeSetPath:
		return 7
	case TypeDatatype:
		return 8
	case TypeNone:
		return 9
	case TypeBlock:
		return 10
	case TypeParen:
		return 11
	default:
		return 99 // unsupported types should be caught by schema validation
	}
}

// compareValuesForSort compares two values for sorting purposes
// Returns -1 if a < b, 0 if a == b, 1 if a > b
// For different types, uses a defined type ordering
func compareValuesForSort(a, b core.Value) int {
	aType := a.GetType()
	bType := b.GetType()

	// If types differ, compare by type precedence
	if aType != bType {
		aPrecedence := getTypePrecedence(aType)
		bPrecedence := getTypePrecedence(bType)
		if aPrecedence < bPrecedence {
			return -1
		} else if aPrecedence > bPrecedence {
			return 1
		}
		return 0 // same precedence (shouldn't happen for different types)
	}

	switch aType {
	case TypeInteger:
		aVal, _ := AsIntValue(a)
		bVal, _ := AsIntValue(b)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	case TypeDecimal:
		aVal, _ := AsDecimal(a)
		bVal, _ := AsDecimal(b)
		if aVal.Magnitude == nil && bVal.Magnitude == nil {
			return 0
		}
		if aVal.Magnitude == nil {
			return -1
		}
		if bVal.Magnitude == nil {
			return 1
		}
		return aVal.Magnitude.Cmp(bVal.Magnitude)
	case TypeString:
		aVal, _ := AsStringValue(a)
		bVal, _ := AsStringValue(b)
		aStr := aVal.Form()
		bStr := bVal.Form()
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	case TypeBinary:
		aVal, _ := AsBinaryValue(a)
		bVal, _ := AsBinaryValue(b)
		aData := aVal.Bytes()
		bData := bVal.Bytes()
		minLen := len(aData)
		if len(bData) < minLen {
			minLen = len(bData)
		}
		for i := 0; i < minLen; i++ {
			if aData[i] < bData[i] {
				return -1
			} else if aData[i] > bData[i] {
				return 1
			}
		}
		if len(aData) < len(bData) {
			return -1
		} else if len(aData) > len(bData) {
			return 1
		}
		return 0
	case TypeLogic:
		aVal, _ := AsLogicValue(a)
		bVal, _ := AsLogicValue(b)
		// false < true
		if !aVal && bVal {
			return -1
		} else if aVal && !bVal {
			return 1
		}
		return 0
	case TypeWord, TypeSetWord, TypeGetWord, TypeLitWord:
		aVal, _ := AsWordValue(a)
		bVal, _ := AsWordValue(b)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	case TypePath, TypeGetPath, TypeSetPath:
		return comparePathsForSort(a, b)
	case TypeDatatype:
		aVal, _ := AsDatatypeValue(a)
		bVal, _ := AsDatatypeValue(b)
		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	case TypeNone:
		return 0 // all none values are equal
	case TypeBlock, TypeParen:
		aVal, _ := AsBlockValue(a)
		bVal, _ := AsBlockValue(b)
		return compareBlockValuesLex(aVal, bVal)
	default:
		// Unsupported type - should not reach here if schema validation worked
		return 0
	}
}

// comparePathsForSort compares two path values lexicographically
func comparePathsForSort(a, b core.Value) int {
	var aSegments, bSegments []PathSegment

	// Extract segments based on path type
	switch a.GetType() {
	case TypePath:
		if path, ok := AsPath(a); ok {
			aSegments = path.Segments
		}
	case TypeGetPath:
		if path, ok := AsGetPath(a); ok {
			aSegments = path.Segments
		}
	case TypeSetPath:
		if path, ok := AsSetPath(a); ok {
			aSegments = path.Segments
		}
	}

	switch b.GetType() {
	case TypePath:
		if path, ok := AsPath(b); ok {
			bSegments = path.Segments
		}
	case TypeGetPath:
		if path, ok := AsGetPath(b); ok {
			bSegments = path.Segments
		}
	case TypeSetPath:
		if path, ok := AsSetPath(b); ok {
			bSegments = path.Segments
		}
	}

	minLen := len(aSegments)
	if len(bSegments) < minLen {
		minLen = len(bSegments)
	}

	for i := 0; i < minLen; i++ {
		aSeg := aSegments[i]
		bSeg := bSegments[i]

		// Compare segment types first
		if aSeg.Type < bSeg.Type {
			return -1
		} else if aSeg.Type > bSeg.Type {
			return 1
		}

		// Same type, compare values
		switch aSeg.Type {
		case PathSegmentWord:
			aWord, _ := aSeg.Value.(string)
			bWord, _ := bSeg.Value.(string)
			if aWord < bWord {
				return -1
			} else if aWord > bWord {
				return 1
			}
		case PathSegmentIndex:
			aIdx, _ := aSeg.Value.(int64)
			bIdx, _ := bSeg.Value.(int64)
			if aIdx < bIdx {
				return -1
			} else if aIdx > bIdx {
				return 1
			}
		case PathSegmentEval:
			// For eval segments, compare the string representation
			aStr := fmt.Sprintf("%v", aSeg.Value)
			bStr := fmt.Sprintf("%v", bSeg.Value)
			if aStr < bStr {
				return -1
			} else if aStr > bStr {
				return 1
			}
		}
	}

	// All compared segments equal, shorter path comes first
	if len(aSegments) < len(bSegments) {
		return -1
	} else if len(aSegments) > len(bSegments) {
		return 1
	}
	return 0
}

func SortBlock(b *BlockValue) {
	sort.SliceStable(b.Elements, func(i, j int) bool {
		elemI := b.Elements[i]
		elemJ := b.Elements[j]
		switch elemI.GetType() {
		case TypeInteger:
			iVal, _ := AsIntValue(elemI)
			jVal, _ := AsIntValue(elemJ)
			return iVal < jVal
		case TypeString:
			iVal, _ := AsStringValue(elemI)
			jVal, _ := AsStringValue(elemJ)
			return iVal.Form() < jVal.Form()
		case TypeBlock, TypeParen:
			iVal, _ := AsBlockValue(elemI)
			jVal, _ := AsBlockValue(elemJ)
			return compareBlockValuesLex(iVal, jVal) < 0
		default:
			return false
		}
	})
}
