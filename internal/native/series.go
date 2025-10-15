package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Copy implements the `copy` native for series values.
//
// Contract: copy series -> new copy
// Contract: copy --part series count -> new copy of first count elements/chars
func Copy(args []core.Value, refinements map[string]core.Value) (core.Value, error) {
	// --part refinement: copy only first N elements/chars
	partVal, hasPart := refinements["part"]
	// Check if refinement was actually provided (not just default none)
	hasPart = hasPart && partVal.GetType() != value.TypeNone
	if len(args) < 1 {
		return value.NoneVal(), arityError("copy", 1, len(args))
	}
	series := args[0]
	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		if hasPart {
			if partVal.GetType() != value.TypeInteger {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count64, ok := value.AsInteger(partVal)
			if !ok {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count := int(count64)
			if count < 0 || count > blk.Length() {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "block", "out of range"})
			}
			elems := make([]core.Value, count)
			copy(elems, blk.Elements[:count])
			return value.BlockVal(elems), nil
		}
		// Full copy
		return value.BlockVal(append([]core.Value{}, blk.Elements...)), nil
	case value.TypeString:
		str, _ := value.AsString(series)
		if hasPart {
			if partVal.GetType() != value.TypeInteger {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count64, ok := value.AsInteger(partVal)
			if !ok {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count := int(count64)
			if count < 0 || count > str.Length() {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "string", "out of range"})
			}
			// Użyj metody String() i substring
			runes := []rune(str.String())
			return value.StrVal(string(runes[:count])), nil
		}
		// Full copy
		return value.StrVal(str.String()), nil
	default:
		return value.NoneVal(), typeError("copy", "series", series)
	}
}

// seriesSelector is a function that selects a value from a series.
// For blocks: receives elements and returns the selected element.
// For strings: receives the string and returns a character as a value.
type seriesSelector func(series core.Value) (core.Value, error)

// seriesOperation provides a template for operations on series (blocks and strings).
// It handles arity checking, type validation, and empty series errors.
func seriesOperation(name string, args []core.Value, sel seriesSelector) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError(name, 1, len(args))
	}

	series := args[0]

	// Validate series type
	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		if blk.Length() == 0 {
			return value.NoneVal(), emptySeriesError(name)
		}
	case value.TypeString:
		str, _ := value.AsString(series)
		if str.Length() == 0 {
			return value.NoneVal(), emptySeriesError(name)
		}
	default:
		return value.NoneVal(), typeError(name, "series", series)
	}

	// Apply selector
	return sel(series)
}

// First implements the `first` native for series values.
//
// Contract: first series -> first element
// - Series must be block or string
// - Error on empty series
func First(args []core.Value) (core.Value, error) {
	return seriesOperation("first", args, func(series core.Value) (core.Value, error) {
		switch series.GetType() {
		case value.TypeBlock:
			blk, _ := value.AsBlock(series)
			return blk.First(), nil
		case value.TypeString:
			str, _ := value.AsString(series)
			return value.StrVal(string(str.First())), nil
		default:
			return value.NoneVal(), typeError("first", "series", series)
		}
	})
}

// Last implements the `last` native for series values.
func Last(args []core.Value) (core.Value, error) {
	return seriesOperation("last", args, func(series core.Value) (core.Value, error) {
		switch series.GetType() {
		case value.TypeBlock:
			blk, _ := value.AsBlock(series)
			return blk.Last(), nil
		case value.TypeString:
			str, _ := value.AsString(series)
			return value.StrVal(string(str.Last())), nil
		default:
			return value.NoneVal(), typeError("last", "series", series)
		}
	})
}

// Append implements the `append` native for series values.
//
// Contract: append series value -> modified series
func Append(args []core.Value) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("append", 2, len(args))
	}

	target := args[0]
	switch target.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(target)
		blk.Append(args[1])
		return target, nil
	case value.TypeString:
		str, _ := value.AsString(target)
		if args[1].GetType() != value.TypeString {
			return value.NoneVal(), typeError("append", "string", args[1])
		}
		insertStr, _ := value.AsString(args[1])
		str.Append(insertStr)
		return target, nil
	default:
		return value.NoneVal(), typeError("append", "series", target)
	}
}

// Insert implements the `insert` native for series values.
func Insert(args []core.Value) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("insert", 2, len(args))
	}

	target := args[0]
	switch target.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(target)
		blk.SetIndex(0)
		blk.Insert(args[1])
		return target, nil
	case value.TypeString:
		str, _ := value.AsString(target)
		if args[1].GetType() != value.TypeString {
			return value.NoneVal(), typeError("insert", "string", args[1])
		}
		insertStr, _ := value.AsString(args[1])
		str.SetIndex(0)
		str.Insert(insertStr)
		return target, nil
	default:
		return value.NoneVal(), typeError("insert", "series", target)
	}
}

// LengthQ implements the `length?` native for series values.
func LengthQ(args []core.Value) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("length?", 1, len(args))
	}

	series := args[0]
	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		return value.IntVal(int64(blk.Length())), nil
	case value.TypeString:
		str, _ := value.AsString(series)
		return value.IntVal(int64(str.Length())), nil
	default:
		return value.NoneVal(), typeError("length?", "series", series)
	}
}

func emptySeriesError(op string) error {
	return verror.NewScriptError(
		verror.ErrIDEmptySeries,
		[3]string{op, "", ""},
	)
}

// Find implements the `find` native for series values.
//
// Contract: find series value -> index (1-based) or none
// Contract: find --last series value -> last index (1-based) or none
func Find(args []core.Value, refinements map[string]core.Value) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("find", 2, len(args))
	}

	series := args[0]
	sought := args[1]
	lastVal, hasLast := refinements["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.LogicVal(true))

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		if isLast {
			for i := blk.Length() - 1; i >= 0; i-- {
				if blk.Elements[i].Equals(sought) {
					return value.IntVal(int64(i + 1)), nil
				}
			}
		} else {
			for i, v := range blk.Elements {
				if v.Equals(sought) {
					return value.IntVal(int64(i + 1)), nil
				}
			}
		}
		return value.NoneVal(), nil

	case value.TypeString:
		str, _ := value.AsString(series)
		soughtStr, ok := value.AsString(sought)
		if !ok {
			return value.NoneVal(), typeError("find", "string", sought)
		}

		runes := []rune(str.String())
		soughtRunes := []rune(soughtStr.String())

		if len(soughtRunes) == 0 {
			return value.NoneVal(), nil // Or error? Let's say none for now.
		}

		if isLast {
			for i := len(runes) - len(soughtRunes); i >= 0; i-- {
				match := true
				for j := range soughtRunes {
					if runes[i+j] != soughtRunes[j] {
						match = false
						break
					}
				}
				if match {
					return value.IntVal(int64(i + 1)), nil
				}
			}
		} else {
			for i := 0; i <= len(runes)-len(soughtRunes); i++ {
				match := true
				for j := range soughtRunes {
					if runes[i+j] != soughtRunes[j] {
						match = false
						break
					}
				}
				if match {
					return value.IntVal(int64(i + 1)), nil
				}
			}
		}
		return value.NoneVal(), nil

	default:
		return value.NoneVal(), typeError("find", "series", series)
	}
}

// Remove implements the `remove` native for series values.
//
// Contract: remove series -> modified series
// Contract: remove series --part count -> modified series
func Remove(args []core.Value, refinements map[string]core.Value) (core.Value, error) {
	partVal, hasPart := refinements["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	if len(args) != 1 {
		return value.NoneVal(), arityError("remove", 1, len(args))
	}

	series := args[0]
	count := 1

	if hasPart {
		if partVal.GetType() != value.TypeInteger {
			return value.NoneVal(), typeError("remove --part", "integer", partVal)
		}
		count64, _ := value.AsInteger(partVal)
		count = int(count64)
	}

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		if count < 0 || count > blk.Length() {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "block", "out of range"})
		}
		blk.SetIndex(0)
		blk.Remove(count)
		return series, nil
	case value.TypeString:
		str, _ := value.AsString(series)
		if count < 0 || count > str.Length() {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "string", "out of range"})
		}
		str.SetIndex(0)
		str.Remove(count)
		return series, nil
	default:
		return value.NoneVal(), typeError("remove", "series", series)
	}
}

// Skip implements the `skip` native for series values.
func Skip(args []core.Value) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("skip", 2, len(args))
	}
	series, countVal := args[0], args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NoneVal(), typeError("skip", "integer", countVal)
	}
	count64, _ := value.AsInteger(countVal)
	count := int(count64)

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		newIndex := blk.GetIndex() + count
		if newIndex < 0 || newIndex > blk.Length() {
			newIndex = blk.Length()
		}
		blk.SetIndex(newIndex)
		return series, nil
	case value.TypeString:
		str, _ := value.AsString(series)
		newIndex := str.Index() + count
		if newIndex < 0 || newIndex > str.Length() {
			newIndex = str.Length()
		}
		str.SetIndex(newIndex)
		return series, nil
	default:
		return value.NoneVal(), typeError("skip", "series", series)
	}
}

// Take implements the `take` native for series values.
func Take(args []core.Value) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("take", 2, len(args))
	}
	series, countVal := args[0], args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NoneVal(), typeError("take", "integer", countVal)
	}
	count64, _ := value.AsInteger(countVal)
	count := int(count64)

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		start := blk.GetIndex()
		end := min(start+count, blk.Length())
		newElements := blk.Elements[start:end]
		blk.SetIndex(end)
		return value.BlockVal(newElements), nil
	case value.TypeString:
		str, _ := value.AsString(series)
		start := str.Index()
		end := min(start+count, str.Length())
		newRunes := str.Runes()[start:end]
		str.SetIndex(end)
		return value.StrVal(string(newRunes)), nil
	default:
		return value.NoneVal(), typeError("take", "series", series)
	}
}

// Sort implements the `sort` native for series values.
func Sort(args []core.Value) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("sort", 1, len(args))
	}
	series := args[0]

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		if blk.Length() == 0 {
			return series, nil
		}
		// Check if all elements are of the same comparable type
		firstType := blk.Elements[0].GetType()
		for _, v := range blk.Elements {
			if v.GetType() != firstType || (v.GetType() != value.TypeInteger && v.GetType() != value.TypeString) {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDNotComparable, [3]string{"sort", "mixed types", ""})
			}
		}

		value.SortBlock(blk)
		return series, nil
	default:
		return value.NoneVal(), typeError("sort", "block", series)
	}
}

// Reverse implements the `reverse` native for series values.
func Reverse(args []core.Value) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("reverse", 1, len(args))
	}
	series := args[0]

	switch series.GetType() {
	case value.TypeBlock:
		blk, _ := value.AsBlock(series)
		for i, j := 0, len(blk.Elements)-1; i < j; i, j = i+1, j-1 {
			blk.Elements[i], blk.Elements[j] = blk.Elements[j], blk.Elements[i]
		}
		return series, nil
	case value.TypeString:
		str, _ := value.AsString(series)
		r := str.Runes()
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		str.SetRunes(r)
		return series, nil
	default:
		return value.NoneVal(), typeError("reverse", "series", series)
	}
}
