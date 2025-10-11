package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Copy implements the `copy` native for series values.
//
// Contract: copy series -> new copy
// Contract: copy --part series count -> new copy of first count elements/chars
func Copy(args []value.Value, refinements map[string]value.Value) (value.Value, *verror.Error) {
	// --part refinement: copy only first N elements/chars
	partVal, hasPart := refinements["part"]
	// Check if refinement was actually provided (not just default none)
	hasPart = hasPart && partVal.Type != value.TypeNone
	if len(args) < 1 {
		return value.NoneVal(), arityError("copy", 1, len(args))
	}
	series := args[0]
	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if hasPart {
			if partVal.Type != value.TypeInteger {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count64, ok := partVal.AsInteger()
			if !ok {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count := int(count64)
			if count < 0 || count > blk.Length() {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "block", "out of range"})
			}
			elems := make([]value.Value, count)
			copy(elems, blk.Elements[:count])
			return value.BlockVal(elems), nil
		}
		// Full copy
		return value.BlockVal(append([]value.Value{}, blk.Elements...)), nil
	case value.TypeString:
		str, _ := series.AsString()
		if hasPart {
			if partVal.Type != value.TypeInteger {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count64, ok := partVal.AsInteger()
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
type seriesSelector func(series value.Value) (value.Value, *verror.Error)

// seriesOperation provides a template for operations on series (blocks and strings).
// It handles arity checking, type validation, and empty series errors.
func seriesOperation(name string, args []value.Value, sel seriesSelector) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError(name, 1, len(args))
	}

	series := args[0]

	// Validate series type
	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if blk.Length() == 0 {
			return value.NoneVal(), emptySeriesError(name)
		}
	case value.TypeString:
		str, _ := series.AsString()
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
func First(args []value.Value) (value.Value, *verror.Error) {
	return seriesOperation("first", args, func(series value.Value) (value.Value, *verror.Error) {
		switch series.Type {
		case value.TypeBlock:
			blk, _ := series.AsBlock()
			return blk.First(), nil
		case value.TypeString:
			str, _ := series.AsString()
			return value.StrVal(string(str.First())), nil
		default:
			return value.NoneVal(), typeError("first", "series", series)
		}
	})
}

// Last implements the `last` native for series values.
func Last(args []value.Value) (value.Value, *verror.Error) {
	return seriesOperation("last", args, func(series value.Value) (value.Value, *verror.Error) {
		switch series.Type {
		case value.TypeBlock:
			blk, _ := series.AsBlock()
			return blk.Last(), nil
		case value.TypeString:
			str, _ := series.AsString()
			return value.StrVal(string(str.Last())), nil
		default:
			return value.NoneVal(), typeError("last", "series", series)
		}
	})
}

// Append implements the `append` native for series values.
//
// Contract: append series value -> modified series
func Append(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("append", 2, len(args))
	}

	target := args[0]
	switch target.Type {
	case value.TypeBlock:
		blk, _ := target.AsBlock()
		blk.Append(args[1])
		return target, nil
	case value.TypeString:
		str, _ := target.AsString()
		if args[1].Type != value.TypeString {
			return value.NoneVal(), typeError("append", "string", args[1])
		}
		insertStr, _ := args[1].AsString()
		str.Append(insertStr)
		return target, nil
	default:
		return value.NoneVal(), typeError("append", "series", target)
	}
}

// Insert implements the `insert` native for series values.
func Insert(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("insert", 2, len(args))
	}

	target := args[0]
	switch target.Type {
	case value.TypeBlock:
		blk, _ := target.AsBlock()
		blk.SetIndex(0)
		blk.Insert(args[1])
		return target, nil
	case value.TypeString:
		str, _ := target.AsString()
		if args[1].Type != value.TypeString {
			return value.NoneVal(), typeError("insert", "string", args[1])
		}
		insertStr, _ := args[1].AsString()
		str.SetIndex(0)
		str.Insert(insertStr)
		return target, nil
	default:
		return value.NoneVal(), typeError("insert", "series", target)
	}
}

// LengthQ implements the `length?` native for series values.
func LengthQ(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("length?", 1, len(args))
	}

	series := args[0]
	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		return value.IntVal(int64(blk.Length())), nil
	case value.TypeString:
		str, _ := series.AsString()
		return value.IntVal(int64(str.Length())), nil
	default:
		return value.NoneVal(), typeError("length?", "series", series)
	}
}

func emptySeriesError(op string) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDEmptySeries,
		[3]string{op, "", ""},
	)
}

// Find implements the `find` native for series values.
//
// Contract: find series value -> index (1-based) or none
// Contract: find --last series value -> last index (1-based) or none
func Find(args []value.Value, refinements map[string]value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("find", 2, len(args))
	}

	series := args[0]
	sought := args[1]
	lastVal, hasLast := refinements["last"]
	isLast := hasLast && lastVal.Type == value.TypeLogic && lastVal.Equals(value.LogicVal(true))

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
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
		str, _ := series.AsString()
		soughtStr, ok := sought.AsString()
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
				for j := 0; j < len(soughtRunes); j++ {
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
				for j := 0; j < len(soughtRunes); j++ {
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
func Remove(args []value.Value, refinements map[string]value.Value) (value.Value, *verror.Error) {
	partVal, hasPart := refinements["part"]
	hasPart = hasPart && partVal.Type != value.TypeNone

	if len(args) != 1 {
		return value.NoneVal(), arityError("remove", 1, len(args))
	}

	series := args[0]
	count := 1

	if hasPart {
		if partVal.Type != value.TypeInteger {
			return value.NoneVal(), typeError("remove --part", "integer", partVal)
		}
		count64, _ := partVal.AsInteger()
		count = int(count64)
	}

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if count < 0 || count > blk.Length() {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "block", "out of range"})
		}
		blk.SetIndex(0)
		blk.Remove(count)
		return series, nil
	case value.TypeString:
		str, _ := series.AsString()
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
func Skip(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("skip", 2, len(args))
	}
	series, countVal := args[0], args[1]
	if countVal.Type != value.TypeInteger {
		return value.NoneVal(), typeError("skip", "integer", countVal)
	}
	count64, _ := countVal.AsInteger()
	count := int(count64)

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		newIndex := blk.GetIndex() + count
		if newIndex < 0 || newIndex > blk.Length() {
			newIndex = blk.Length()
		}
		blk.SetIndex(newIndex)
		return series, nil
	case value.TypeString:
		str, _ := series.AsString()
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
func Take(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("take", 2, len(args))
	}
	series, countVal := args[0], args[1]
	if countVal.Type != value.TypeInteger {
		return value.NoneVal(), typeError("take", "integer", countVal)
	}
	count64, _ := countVal.AsInteger()
	count := int(count64)

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		start := blk.GetIndex()
		end := start + count
		if end > blk.Length() {
			end = blk.Length()
		}
		newElements := blk.Elements[start:end]
		blk.SetIndex(end)
		return value.BlockVal(newElements), nil
	case value.TypeString:
		str, _ := series.AsString()
		start := str.Index()
		end := start + count
		if end > str.Length() {
			end = str.Length()
		}
		newRunes := str.Runes()[start:end]
		str.SetIndex(end)
		return value.StrVal(string(newRunes)), nil
	default:
		return value.NoneVal(), typeError("take", "series", series)
	}
}

// Sort implements the `sort` native for series values.
func Sort(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("sort", 1, len(args))
	}
	series := args[0]

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if blk.Length() == 0 {
			return series, nil
		}
		// Check if all elements are of the same comparable type
		firstType := blk.Elements[0].Type
		for _, v := range blk.Elements {
			if v.Type != firstType || (v.Type != value.TypeInteger && v.Type != value.TypeString) {
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
func Reverse(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("reverse", 1, len(args))
	}
	series := args[0]

	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		for i, j := 0, len(blk.Elements)-1; i < j; i, j = i+1, j-1 {
			blk.Elements[i], blk.Elements[j] = blk.Elements[j], blk.Elements[i]
		}
		return series, nil
	case value.TypeString:
		str, _ := series.AsString()
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
