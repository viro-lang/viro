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
func Copy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// --part refinement: copy only first N elements/chars
	partVal, hasPart := refValues["part"]
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
	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		if hasPart {
			if partVal.GetType() != value.TypeInteger {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count64, ok := value.AsInteger(partVal)
			if !ok {
				return value.NoneVal(), typeError("copy --part", "integer", partVal)
			}
			count := int(count64)
			if count < 0 || count > bin.Length() {
				return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"copy --part", "binary", "out of range"})
			}
			// Copy first count bytes
			bytes := make([]byte, count)
			copy(bytes, bin.Bytes()[:count])
			return value.BinaryVal(bytes), nil
		}
		// Full copy
		return value.BinaryVal(append([]byte{}, bin.Bytes()...)), nil
	default:
		return value.NoneVal(), typeError("copy", "series", series)
	}
}

// Find implements the `find` native for series values.
//
// Contract: find series value -> index (1-based) or none
// Contract: find --last series value -> last index (1-based) or none
func Find(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("find", 2, len(args))
	}

	series := args[0]
	sought := args[1]
	lastVal, hasLast := refValues["last"]
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

	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		soughtByte, ok := value.AsInteger(sought)
		if !ok || soughtByte < 0 || soughtByte > 255 {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"byte value (0-255)", value.TypeToString(sought.GetType()), ""})
		}

		bytes := bin.Bytes()
		targetByte := byte(soughtByte)

		if isLast {
			for i := len(bytes) - 1; i >= 0; i-- {
				if bytes[i] == targetByte {
					return value.IntVal(int64(i + 1)), nil
				}
			}
		} else {
			for i, b := range bytes {
				if b == targetByte {
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
func Remove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	partVal, hasPart := refValues["part"]
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
	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		if count < 0 || count > bin.Length() {
			return value.NoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"remove", "binary", "out of range"})
		}
		bin.SetIndex(0)
		bin.Remove(count)
		return series, nil
	default:
		return value.NoneVal(), typeError("remove", "series", series)
	}
}

// Skip implements the `skip` native for series values.
func Skip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		newIndex := bin.GetIndex() + count
		if newIndex < 0 || newIndex > bin.Length() {
			newIndex = bin.Length()
		}
		bin.SetIndex(newIndex)
		return series, nil
	default:
		return value.NoneVal(), typeError("skip", "series", series)
	}
}

// Take implements the `take` native for series values.
func Take(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		start := bin.GetIndex()
		end := min(start+count, bin.Length())
		newBytes := bin.Bytes()[start:end]
		bin.SetIndex(end)
		return value.BinaryVal(newBytes), nil
	default:
		return value.NoneVal(), typeError("take", "series", series)
	}
}

// Sort implements the `sort` native for series values.
func Sort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
func Reverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
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
	case value.TypeBinary:
		bin, _ := value.AsBinary(series)
		bytes := bin.Bytes()
		for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
			bytes[i], bytes[j] = bytes[j], bytes[i]
		}
		return series, nil
	default:
		return value.NoneVal(), typeError("reverse", "series", series)
	}
}
