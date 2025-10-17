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
	}, BlockFirst))
	RegisterActionImpl(value.TypeBlock, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockLast))
	RegisterActionImpl(value.TypeBlock, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BlockAppend))
	RegisterActionImpl(value.TypeBlock, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, BlockInsert))
	RegisterActionImpl(value.TypeBlock, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockLength))

	// Register string-specific implementations
	RegisterActionImpl(value.TypeString, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringFirst))
	RegisterActionImpl(value.TypeString, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringLast))
	RegisterActionImpl(value.TypeString, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, StringAppend))
	RegisterActionImpl(value.TypeString, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, StringInsert))
	RegisterActionImpl(value.TypeString, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringLength))
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

	// Helper function to wrap simple series functions
	registerSimpleSeriesFunc := func(name string, impl core.NativeFunc, arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := range arity {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"series", "value"}
			for i := range arity {
				if i < len(paramNames) {
					params[i] = value.NewParamSpec(paramNames[i], true)
				} else {
					params[i] = value.NewParamSpec("arg", true)
				}
			}
		}

		fn := value.NewNativeFunction(
			name,
			params,
			impl,
		)
		fn.Doc = doc
		registerAndBind(name, value.FuncVal(fn))
	}

	// Register type-specific implementations into type frames
	registerSeriesTypeImpls()

	// ===== Group 5: Series operations (5 actions) =====
	// These are now actions that dispatch to type-specific implementations

	// first - action
	registerAndBind("first", CreateAction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the first element of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to get first element from"},
		},
		Returns:  "any! The first element of the series",
		Examples: []string{"first [1 2 3]  ; => 1", `first "hello"  ; => "h"`},
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
			{Name: "series", Type: "block! string!", Description: "The series to get last element from"},
		},
		Returns:  "any! The last element of the series",
		Examples: []string{"last [1 2 3]  ; => 3", `last "hello"  ; => "o"`},
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
			{Name: "series", Type: "block! string!", Description: "The series to append to"},
			{Name: "value", Type: "any!", Description: "The value to append"},
		},
		Returns:  "block! string! The modified series",
		Examples: []string{"append [1 2] 3  ; => [1 2 3]", `append "hel" "lo"  ; => "hello"`},
		SeeAlso:  []string{"insert", "skip", "take"},
		Tags:     []string{"series", "modification"},
	}))

	// insert - action
	registerAndBind("insert", CreateAction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Inserts a value at the beginning of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to insert into"},
			{Name: "value", Type: "any!", Description: "The value to insert"},
		},
		Returns:  "block! string! The modified series",
		Examples: []string{"insert [2 3] 1  ; => [1 2 3]", `insert "ello" "h"  ; => "hello"`},
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
			{Name: "series", Type: "block! string!", Description: "The series to get length of"},
		},
		Returns:  "integer! The number of elements in the series",
		Examples: []string{"length? [1 2 3]  ; => 3", `length? "hello"  ; => 5`},
		SeeAlso:  []string{"first", "last", "skip", "take"},
		Tags:     []string{"series", "query"},
	}))
	registerSimpleSeriesFunc("skip", Skip, 2, &NativeDoc{
		Category: "Series",
		Summary:  "Skips n elements in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to skip from"},
			{Name: "n", Type: "integer!", Description: "Number of elements to skip"},
		},
		Returns:  "[block! string!] Series with first n elements removed",
		Examples: []string{"skip [1 2 3 4] 2  ; => [3 4]"},
		SeeAlso:  []string{"take", "first", "last"},
		Tags:     []string{"series"},
	})
	registerSimpleSeriesFunc("take", Take, 2, &NativeDoc{
		Category: "Series",
		Summary:  "Takes n elements from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to take from"},
			{Name: "n", Type: "integer!", Description: "Number of elements to take"},
		},
		Returns:  "[block! string!] Series with first n elements",
		Examples: []string{"take [1 2 3 4] 2  ; => [1 2]"},
		SeeAlso:  []string{"skip", "first", "last"},
		Tags:     []string{"series"},
	})
	registerSimpleSeriesFunc("sort", Sort, 1, &NativeDoc{
		Category: "Series",
		Summary:  "Sorts a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to sort"},
		},
		Returns:  "[block! string!] The sorted series",
		Examples: []string{"sort [3 1 2]  ; => [1 2 3]"},
		SeeAlso:  []string{"reverse"},
		Tags:     []string{"series"},
	})
	registerSimpleSeriesFunc("reverse", Reverse, 1, &NativeDoc{
		Category: "Series",
		Summary:  "Reverses a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to reverse"},
		},
		Returns:  "[block! string!] The reversed series",
		Examples: []string{"reverse [1 2 3]  ; => [3 2 1]"},
		SeeAlso:  []string{"sort"},
		Tags:     []string{"series"},
	})

	// Functions with refinements need special handling
	// Copy function
	{
		params := []value.ParamSpec{
			value.NewParamSpec("series", true),
			value.NewRefinementSpec("part", true),
		}
		fn := value.NewNativeFunction(
			"copy",
			params,
			Copy,
		)
		fn.Doc = &NativeDoc{
			Category: "Series",
			Summary:  "Copies a series",
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to copy"},
				{Name: "--part", Type: "integer!", Description: "Copy only first N elements", Optional: true},
			},
			Returns:  "[block! string!] A copy of the series",
			Examples: []string{"copy [1 2 3]  ; => [1 2 3]"},
			SeeAlso:  []string{"append", "insert"},
			Tags:     []string{"series"},
		}
		registerAndBind("copy", value.FuncVal(fn))
	}

	// Find function
	{
		params := []value.ParamSpec{
			value.NewParamSpec("series", true),
			value.NewParamSpec("value", true),
			value.NewRefinementSpec("last", false),
		}
		fn := value.NewNativeFunction(
			"find",
			params,
			Find,
		)
		fn.Doc = &NativeDoc{
			Category: "Series",
			Summary:  "Finds a value in a series",
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to search"},
				{Name: "value", Type: "any-type!", Description: "The value to find"},
				{Name: "--last", Type: "", Description: "Find last occurrence instead of first", Optional: true},
			},
			Returns:  "[block! string! none!] Series from found position or none",
			Examples: []string{"find [1 2 3] 2  ; => [2 3]"},
			SeeAlso:  []string{"append", "insert"},
			Tags:     []string{"series"},
		}
		registerAndBind("find", value.FuncVal(fn))
	}

	// Remove function
	{
		params := []value.ParamSpec{
			value.NewParamSpec("series", true),
			value.NewRefinementSpec("part", true),
		}
		fn := value.NewNativeFunction(
			"remove",
			params,
			Remove,
		)
		fn.Doc = &NativeDoc{
			Category: "Series",
			Summary:  "Removes elements from a series",
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to remove from"},
				{Name: "--part", Type: "integer!", Description: "Remove n elements", Optional: true},
			},
			Returns:  "[block! string!] The modified series",
			Examples: []string{"remove [1 2 3]  ; => [2 3]", "remove --part 2 [1 2 3]  ; => [3]"},
			SeeAlso:  []string{"append", "insert"},
			Tags:     []string{"series"},
		}
		registerAndBind("remove", value.FuncVal(fn))
	}
}
