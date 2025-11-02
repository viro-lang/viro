// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// registerSeriesTypeImpls registers type-specific implementations into type frames.
// Called by RegisterSeriesNatives after type frames are initialized.
//
// Feature: 004-dynamic-function-invocation
func registerSeriesTypeImpls() {
	// Register block-specific implementations
	RegisterActionImpl(value.TypeBlock, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockFirst, false, nil))
	RegisterActionImpl(value.TypeBlock, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockLast, false, nil))
	RegisterActionImpl(value.TypeBlock, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BlockAppend, false, nil))
	RegisterActionImpl(value.TypeBlock, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BlockInsert, false, nil))
	RegisterActionImpl(value.TypeBlock, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockLength, false, nil))
	RegisterActionImpl(value.TypeBlock, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, BlockCopy, false, nil))
	RegisterActionImpl(value.TypeBlock, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, BlockFind, false, nil))
	RegisterActionImpl(value.TypeBlock, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, BlockRemove, false, nil))
	RegisterActionImpl(value.TypeBlock, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, BlockSkip, false, nil))
	RegisterActionImpl(value.TypeBlock, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockNext, false, nil))
	RegisterActionImpl(value.TypeBlock, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockHead, false, nil))
	RegisterActionImpl(value.TypeBlock, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockIndex, false, nil))
	RegisterActionImpl(value.TypeBlock, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, BlockTake, false, nil))
	RegisterActionImpl(value.TypeBlock, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockSort, false, nil))
	RegisterActionImpl(value.TypeBlock, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockReverse, false, nil))
	RegisterActionImpl(value.TypeBlock, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, BlockAt, false, nil))
	RegisterActionImpl(value.TypeBlock, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockTail, false, nil))

	// Register string-specific implementations
	RegisterActionImpl(value.TypeString, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringFirst, false, nil))
	RegisterActionImpl(value.TypeString, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringLast, false, nil))
	RegisterActionImpl(value.TypeString, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, StringAppend, false, nil))
	RegisterActionImpl(value.TypeString, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, StringInsert, false, nil))
	RegisterActionImpl(value.TypeString, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringLength, false, nil))
	RegisterActionImpl(value.TypeString, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, StringCopy, false, nil))
	RegisterActionImpl(value.TypeString, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, StringFind, false, nil))
	RegisterActionImpl(value.TypeString, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, StringRemove, false, nil))
	RegisterActionImpl(value.TypeString, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, StringSkip, false, nil))
	RegisterActionImpl(value.TypeString, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringNext, false, nil))
	RegisterActionImpl(value.TypeString, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringHead, false, nil))
	RegisterActionImpl(value.TypeString, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringIndex, false, nil))
	RegisterActionImpl(value.TypeString, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, StringAt, false, nil))
	RegisterActionImpl(value.TypeString, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringTail, false, nil))
	RegisterActionImpl(value.TypeString, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringSort, false, nil))
	RegisterActionImpl(value.TypeString, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringReverse, false, nil))
	RegisterActionImpl(value.TypeBinary, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryFirst, false, nil))
	RegisterActionImpl(value.TypeBinary, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryLast, false, nil))
	RegisterActionImpl(value.TypeBinary, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BinaryAppend, false, nil))
	RegisterActionImpl(value.TypeBinary, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BinaryInsert, false, nil))
	RegisterActionImpl(value.TypeBinary, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryLength, false, nil))
	RegisterActionImpl(value.TypeBinary, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, BinaryCopy, false, nil))
	RegisterActionImpl(value.TypeBinary, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, BinaryFind, false, nil))
	RegisterActionImpl(value.TypeBinary, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, BinaryRemove, false, nil))
	RegisterActionImpl(value.TypeBinary, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, BinarySkip, false, nil))
	RegisterActionImpl(value.TypeBinary, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryNext, false, nil))
	RegisterActionImpl(value.TypeBinary, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryHead, false, nil))
	RegisterActionImpl(value.TypeBinary, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryIndex, false, nil))
	RegisterActionImpl(value.TypeBinary, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, BinaryAt, false, nil))
	RegisterActionImpl(value.TypeBinary, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryTail, false, nil))
	RegisterActionImpl(value.TypeBinary, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinarySort, false, nil))
	RegisterActionImpl(value.TypeBinary, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinaryReverse, false, nil))
}

// RegisterSeriesNatives registers all series-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterSeriesNatives(rootFrame core.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function or action
	registerAndBind := func(name string, val core.Value) {
		if val.GetType() == value.TypeNone {
			panic(fmt.Sprintf("RegisterSeriesNatives: attempted to register nil value for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterSeriesNatives: duplicate registration of '%s'", name))
		}

		// Bind to root frame
		rootFrame.Bind(name, val)

		// Mark as registered
		registered[name] = true
	}

	// Register type-specific implementations into type frames
	registerSeriesTypeImpls()

	// ===== Group 5: Series operations (12 actions) =====
	// All series operations now use action dispatch to type-specific implementations

	// first - action
	registerAndBind("first", CreateAction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the first element of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get first element from"},
		},
		Returns:  "any! The first element of the series",
		Examples: []string{"first [1 2 3]  ; => 1", `first "hello"  ; => "h"`, "first #{DEADBEEF}  ; => 222"},
		SeeAlso:  []string{"last", "skip", "take"},
		Tags:     []string{"series"},
	}))

	// last - action
	registerAndBind("last", CreateAction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the last element of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get last element from"},
		},
		Returns:  "any! The last element of the series",
		Examples: []string{"last [1 2 3]  ; => 3", `last "hello"  ; => "o"`, "last #{DEADBEEF}  ; => 239"},
		SeeAlso:  []string{"first", "skip", "take"},
		Tags:     []string{"series"},
	}))

	// append - action
	registerAndBind("append", CreateAction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Appends a value to the end of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to append to"},
			{Name: "value", Type: "any!", Description: "The value to append"},
		},
		Returns:  "block! string! binary! The modified series",
		Examples: []string{"append [1 2] 3  ; => [1 2 3]", `append "hel" "lo"  ; => "hello"`, "append #{DEAD} 190  ; => #{DEADBE}"},
		SeeAlso:  []string{"insert", "skip", "take"},
		Tags:     []string{"series", "modification"},
	}))

	// at - action
	registerAndBind("at", CreateAction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the element at the specified 1-based index from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get element from"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to return"},
		},
		Returns:  "any! The element at the specified index",
		Examples: []string{"at [1 2 3] 2  ; => 2", `at "hello" 1  ; => "h"`, "at #{DEADBEEF} 3  ; => 190"},
		SeeAlso:  []string{"first", "last", "skip", "take"},
		Tags:     []string{"series", "indexing"},
	}))

	// insert - action
	registerAndBind("insert", CreateAction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Inserts a value at the beginning of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to insert into"},
			{Name: "value", Type: "any!", Description: "The value to insert"},
		},
		Returns:  "block! string! binary! The modified series",
		Examples: []string{"insert [2 3] 1  ; => [1 2 3]", `insert "ello" "h"  ; => "hello"`, "insert #{ADBE} #{DE}  ; => #{DEADBE}"},
		SeeAlso:  []string{"append", "skip", "take"},
		Tags:     []string{"series", "modification"},
	}))

	// length? - action
	registerAndBind("length?", CreateAction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the length of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get length of"},
		},
		Returns:  "integer! The number of elements in the series",
		Examples: []string{"length? [1 2 3]  ; => 3", `length? "hello"  ; => 5`, "length? #{DEADBEEF}  ; => 4"},
		SeeAlso:  []string{"first", "last", "skip", "take"},
		Tags:     []string{"series", "query"},
	}))

	// copy - action
	registerAndBind("copy", CreateAction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Copies a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to copy"},
			{Name: "--part", Type: "integer!", Description: "Copy only first N elements", Optional: true},
		},
		Returns:  "block! string! binary! A copy of the series",
		Examples: []string{"copy [1 2 3]  ; => [1 2 3]", `copy "hello"  ; => "hello"`, "copy #{DEADBEEF}  ; => #{DEADBEEF}"},
		SeeAlso:  []string{"append", "insert"},
		Tags:     []string{"series"},
	}))

	// find - action
	registerAndBind("find", CreateAction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Finds a value in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to search"},
			{Name: "value", Type: "any!", Description: "The value to find"},
			{Name: "--last", Type: "", Description: "Find last occurrence instead of first", Optional: true},
		},
		Returns:  "integer! 1-based index or none",
		Examples: []string{"find [1 2 3] 2  ; => 2", `find "hello" "l"  ; => 3`, "find #{DEADBEEF} 190  ; => 3"},
		SeeAlso:  []string{"first", "last"},
		Tags:     []string{"series", "search"},
	}))

	// remove - action
	registerAndBind("remove", CreateAction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes elements from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to remove from"},
			{Name: "--part", Type: "integer!", Description: "Remove n elements", Optional: true},
		},
		Returns:  "block! string! binary! The modified series",
		Examples: []string{"remove [1 2 3]  ; => [2 3]", "remove --part 2 [1 2 3]  ; => [3]", `remove "hello"  ; => "ello"`, "remove #{DEADBEEF}  ; => #{ADBE}"},
		SeeAlso:  []string{"append", "insert"},
		Tags:     []string{"series", "modification"},
	}))

	// skip - action
	registerAndBind("skip", CreateAction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Skips n elements in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to skip from"},
			{Name: "count", Type: "integer!", Description: "Number of elements to skip"},
		},
		Returns:  "block! string! binary! Series with index advanced by count",
		Examples: []string{"skip [1 2 3 4] 2  ; => [1 2 3 4] (index at 3)", `skip "hello" 2  ; => "hello" (index at 3)`, "skip #{DEADBEEF} 2  ; => #{DEADBEEF} (index at 3)"},
		SeeAlso:  []string{"take", "first", "last"},
		Tags:     []string{"series"},
	}))

	// next - action
	registerAndBind("next", CreateAction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference advanced by one position",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to advance"},
		},
		Returns:  "block! string! binary! New series reference at next position",
		Examples: []string{"next [1 2 3]  ; => [1 2 3] (index at 2)", `next "hello"  ; => "hello" (index at 2)`, "next #{DEADBEEF}  ; => #{DEADBEEF} (index at 2)"},
		SeeAlso:  []string{"skip", "back", "head", "tail"},
		Tags:     []string{"series", "navigation"},
	}))

	// head - action
	registerAndBind("head", CreateAction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference positioned at the head (position 0)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to position at head"},
		},
		Returns:  "block! string! binary! New series reference at head position",
		Examples: []string{"head [1 2 3]  ; => [1 2 3] (index at 1)", `head "hello"  ; => "hello" (index at 1)`, "head #{DEADBEEF}  ; => #{DEADBEEF} (index at 1)"},
		SeeAlso:  []string{"tail", "next", "back"},
		Tags:     []string{"series", "navigation"},
	}))

	// tail - action
	registerAndBind("tail", CreateAction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series containing all elements except the first one",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get tail from"},
		},
		Returns:  "block! string! binary! New series containing all elements except the first",
		Examples: []string{"tail [1 2 3 4]  ; => [2 3 4]", `tail "hello"  ; => "ello"`, "tail #{DEADBEEF}  ; => #{ADBE}"},
		SeeAlso:  []string{"head", "first", "last"},
		Tags:     []string{"series", "navigation"},
	}))

	// index? - action
	registerAndBind("index?", CreateAction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the current index position of a series (1-based)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get index from"},
		},
		Returns:  "integer! The current index position (1-based)",
		Examples: []string{"index? [1 2 3]  ; => 1", `index? next "hello"  ; => 2`, "index? skip #{DEADBEEF} 2  ; => 3"},
		SeeAlso:  []string{"head", "next", "skip"},
		Tags:     []string{"series", "query"},
	}))

	// take - action
	registerAndBind("take", CreateAction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Takes n elements from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to take from"},
			{Name: "count", Type: "integer!", Description: "Number of elements to take"},
		},
		Returns:  "block! string! binary! Series containing first count elements",
		Examples: []string{"take [1 2 3 4] 2  ; => [1 2]", `take "hello" 2  ; => "he"`, "take #{DEADBEEF} 2  ; => #{DEAD}"},
		SeeAlso:  []string{"skip", "first", "last"},
		Tags:     []string{"series"},
	}))

	// sort - action
	registerAndBind("sort", CreateAction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Sorts a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to sort"},
		},
		Returns: "block! string! binary! The sorted series",
		Examples: []string{
			"sort [3 1 2]  ; => [1 2 3]",
			`sort "cba"  ; => "abc"`,
			"sort #{030201}  ; => #{010203}",
		},
		SeeAlso: []string{"reverse"},
		Tags:    []string{"series", "sorting"},
	}))

	// reverse - action
	registerAndBind("reverse", CreateAction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Reverses a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to reverse"},
		},
		Returns:  "block! string! binary! The reversed series",
		Examples: []string{"reverse [1 2 3]  ; => [3 2 1]", `reverse "hello"  ; => "olleh"`, "reverse #{DEADBEEF}  ; => #{EFBEADDE}"},
		SeeAlso:  []string{"sort"},
		Tags:     []string{"series"},
	}))
}
