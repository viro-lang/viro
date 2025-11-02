package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func registerBlockSeriesActions() {
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
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeBlock, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeBlock, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeBlock, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
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
	RegisterActionImpl(value.TypeBlock, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeBlock, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, seriesPoke, false, nil))
	RegisterActionImpl(value.TypeBlock, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesSelect, false, nil))
	RegisterActionImpl(value.TypeBlock, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeBlock, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeBlock, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))

	RegisterActionImpl(value.TypeBlock, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeBlock, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeBlock, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
}

func registerStringSeriesActions() {
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
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeString, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeString, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeString, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
	RegisterActionImpl(value.TypeString, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, StringAt, false, nil))
	RegisterActionImpl(value.TypeString, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeString, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, seriesPoke, false, nil))
	RegisterActionImpl(value.TypeString, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesSelect, false, nil))
	RegisterActionImpl(value.TypeString, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeString, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeString, "trim", value.NewNativeFunction("trim", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTrim, false, nil))
	RegisterActionImpl(value.TypeString, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))
	RegisterActionImpl(value.TypeString, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeString, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeString, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
	RegisterActionImpl(value.TypeString, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringSort, false, nil))
	RegisterActionImpl(value.TypeString, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringReverse, false, nil))
	RegisterActionImpl(value.TypeString, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, StringTake, false, nil))
}

func registerBinarySeriesActions() {
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
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeBinary, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeBinary, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeBinary, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
	RegisterActionImpl(value.TypeBinary, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, BinaryAt, false, nil))
	RegisterActionImpl(value.TypeBinary, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeBinary, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, seriesPoke, false, nil))
	RegisterActionImpl(value.TypeBinary, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesSelect, false, nil))
	RegisterActionImpl(value.TypeBinary, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeBinary, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeBinary, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))
	RegisterActionImpl(value.TypeBinary, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeBinary, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeBinary, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
	RegisterActionImpl(value.TypeBinary, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinarySort, false, nil))
	RegisterActionImpl(value.TypeBinary, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, BinaryTake, false, nil))
}

func registerSeriesTypeImpls() {
	registerBlockSeriesActions()
	registerStringSeriesActions()
	registerBinarySeriesActions()
}

func RegisterSeriesNatives(rootFrame core.Frame) {
	registered := make(map[string]bool)

	registerAndBind := func(name string, val core.Value) {
		if val.GetType() == value.TypeNone {
			panic(fmt.Sprintf("RegisterSeriesNatives: attempted to register nil value for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterSeriesNatives: duplicate registration of '%s'", name))
		}

		rootFrame.Bind(name, val)

		registered[name] = true
	}

	registerSeriesTypeImpls()

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

	registerAndBind("pick", CreateAction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the element at the specified 1-based index from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get element from"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to return"},
		},
		Returns:  "any! The element at the specified index, or none if index is out of bounds",
		Examples: []string{"pick [1 2 3] 2  ; => 2", `pick "hello" 1  ; => "h"`, "pick #{DEADBEEF} 3  ; => 190", "pick [1 2 3] 10  ; => none"},
		SeeAlso:  []string{"at", "first", "last", "poke"},
		Tags:     []string{"series", "indexing"},
	}))

	registerAndBind("poke", CreateAction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Sets the element at the specified 1-based index in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to modify"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to set"},
			{Name: "value", Type: "any!", Description: "The new value to set at the index"},
		},
		Returns:  "any! The value that was set",
		Examples: []string{"poke [1 2 3] 2 99  ; => 99, series becomes [1 99 3]", `poke "hello" 1 "H"  ; => "H", series becomes "Hello"`},
		SeeAlso:  []string{"at", "change", "insert"},
		Tags:     []string{"series", "modification", "indexing"},
	}))

	registerAndBind("select", CreateAction("select", []value.ParamSpec{
		value.NewParamSpec("target", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("default", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Finds a value in a series or field in an object and returns associated value",
		Description: `Polymorphic lookup action that works on series and objects.

For blocks: searches for the value and returns the next element (key-value pairs).
For strings/binary: finds the pattern and returns the remaining portion after it.
For objects: looks up the field name and returns its value (searches prototype chain).

The --default refinement provides a fallback when the value/field is not found.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "block! string! binary! object!", Description: "The series or object to search"},
			{Name: "value", Type: "any!", Description: "For series: value to find. For objects: field name (word or string)"},
			{Name: "--default", Type: "any!", Description: "Optional fallback value when search/lookup fails"},
		},
		Returns: "any! The found value, or default, or none",
		Examples: []string{
			"select [a 1 b 2] 'b  ; => 2",
			`select "hello world" " "  ; => "world"`,
			"obj: object [x: 10]\nselect obj 'x  ; => 10",
			"select obj 'missing --default 99  ; => 99",
		},
		SeeAlso: []string{"find", "at", "index?", "put", "get"},
		Tags:    []string{"series", "search", "objects", "lookup"},
	}))

	registerAndBind("clear", CreateAction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes all elements from a series and resets index to head",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to clear"},
		},
		Returns:  "block! string! binary! The cleared series (same reference)",
		Examples: []string{"clear [1 2 3]  ; => [], series becomes empty with index at head", `clear "hello"  ; => "", series becomes empty with index at head`},
		SeeAlso:  []string{"append", "insert", "remove"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("change", CreateAction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Replaces the element at the current index with a new value",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to modify"},
			{Name: "value", Type: "any!", Description: "The new value to set at current index (string: single character, binary: 0-255)"},
		},
		Returns:  "any! The value that was set",
		Examples: []string{"change next [1 2 3] 99  ; => 99, series becomes [1 99 3]", `change next "hello" "H"  ; => "H", series becomes "Hello" (single char only)`},
		SeeAlso:  []string{"poke", "at", "insert"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("trim", CreateAction("trim", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("head", false),  // flag: remove from head only
		value.NewRefinementSpec("tail", false),  // flag: remove from tail only
		value.NewRefinementSpec("auto", false),  // flag: auto indent relative to first line
		value.NewRefinementSpec("lines", false), // flag: remove line breaks and extra spaces
		value.NewRefinementSpec("all", false),   // flag: remove all whitespace
		value.NewRefinementSpec("with", true),   // value: remove characters in string instead of whitespace
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes whitespace from strings with various trimming options",
		Parameters: []ParamDoc{
			{Name: "series", Type: "string!", Description: "The string to trim"},
			{Name: "--head", Type: "flag", Description: "Remove whitespace from head only", Optional: true},
			{Name: "--tail", Type: "flag", Description: "Remove whitespace from tail only", Optional: true},
			{Name: "--auto", Type: "flag", Description: "Auto indent lines relative to first line", Optional: true},
			{Name: "--lines", Type: "flag", Description: "Remove all line breaks and extra spaces", Optional: true},
			{Name: "--all", Type: "flag", Description: "Remove all whitespace", Optional: true},
			{Name: "--with", Type: "string!", Description: "Remove characters in this string instead of whitespace", Optional: true},
		},
		Returns: "string! The trimmed string",
		Examples: []string{
			`trim "  hello  "  ; => "hello"`,
			`trim/head "  hello  "  ; => "hello  "`,
			`trim/tail "  hello  "  ; => "  hello"`,
			`trim/all "  hello world  "  ; => "helloworld"`,
			`trim/with "a-b-c" "-"  ; => "abc"`,
		},
		SeeAlso: []string{"clear", "change"},
		Tags:    []string{"series", "string", "modification"},
	}))

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

	registerAndBind("back", CreateAction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference moved backward by one position",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to move backward"},
		},
		Returns:  "block! string! binary! New series reference at previous position",
		Examples: []string{"back next [1 2 3]  ; => [1 2 3] (index at 1)", `back next "hello"  ; => "hello" (index at 1)`, "back next #{DEADBEEF}  ; => #{DEADBEEF} (index at 1)"},
		SeeAlso:  []string{"next", "head", "tail", "skip"},
		Tags:     []string{"series", "navigation"},
	}))

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

	registerAndBind("tail", CreateAction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series positioned at the tail",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to position at tail"},
		},
		Returns:  "block! string! binary! New series reference at tail position",
		Examples: []string{"tail [1 2 3 4]  ; => [1 2 3 4] (index at 4)", `tail "hello"  ; => "hello" (index at 5)`, "tail #{DEADBEEF}  ; => #{DEADBEEF} (index at 4)"},
		SeeAlso:  []string{"head", "next", "back"},
		Tags:     []string{"series", "navigation"},
	}))

	registerAndBind("empty?", CreateAction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true if the series has zero elements",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if series is empty, false otherwise",
		Examples: []string{"empty? []  ; => true", "empty? [1 2 3]  ; => false", `empty? ""  ; => true`, `empty? "hello"  ; => false`},
		SeeAlso:  []string{"length?", "head?", "tail?"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("head?", CreateAction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true if the series index is at position 0 (head)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if series is at head position, false otherwise",
		Examples: []string{"head? [1 2 3]  ; => true", "head? next [1 2 3]  ; => false", `head? "hello"  ; => true`, `head? next "hello"  ; => false`},
		SeeAlso:  []string{"tail?", "index?", "head"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("tail?", CreateAction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true if the series index is at the end (index == length)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if series is at tail position, false otherwise",
		Examples: []string{"tail? [1 2 3]  ; => false", "tail? tail [1 2 3]  ; => true", `tail? "hello"  ; => false`, `tail? tail "hello"  ; => true`},
		SeeAlso:  []string{"head?", "index?", "tail"},
		Tags:     []string{"series", "query"},
	}))

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
