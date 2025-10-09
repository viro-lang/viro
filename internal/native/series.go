package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// First implements the `first` native for series values.
//
// Contract: first series -> first element
// - Series must be block or string
// - Error on empty series
func First(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("first", 1, len(args))
	}

	series := args[0]
	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if blk.Length() == 0 {
			return value.NoneVal(), emptySeriesError("first")
		}
		return blk.First(), nil
	case value.TypeString:
		str, _ := series.AsString()
		if str.Length() == 0 {
			return value.NoneVal(), emptySeriesError("first")
		}
		return value.StrVal(string(str.First())), nil
	default:
		return value.NoneVal(), typeError("first", "series", series)
	}
}

// Last implements the `last` native for series values.
func Last(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("last", 1, len(args))
	}

	series := args[0]
	switch series.Type {
	case value.TypeBlock:
		blk, _ := series.AsBlock()
		if blk.Length() == 0 {
			return value.NoneVal(), emptySeriesError("last")
		}
		return blk.Last(), nil
	case value.TypeString:
		str, _ := series.AsString()
		if str.Length() == 0 {
			return value.NoneVal(), emptySeriesError("last")
		}
		return value.StrVal(string(str.Last())), nil
	default:
		return value.NoneVal(), typeError("last", "series", series)
	}
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
