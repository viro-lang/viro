// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// RegisterSeriesNatives registers all series-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterSeriesNatives(rootFrame *frame.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function
	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterSeriesNatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterSeriesNatives: duplicate registration of function '%s'", name))
		}

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))

		// TEMPORARY: Also populate deprecated Registry for help system
		Registry[name] = fn

		// Mark as registered
		registered[name] = true
	}

	// Helper function to wrap simple series functions
	registerSimpleSeriesFunc := func(name string, impl func([]value.Value) (value.Value, *verror.Error), arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := 0; i < arity; i++ {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"series", "value"}
			for i := 0; i < arity; i++ {
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
			func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := impl(args)
				if err == nil {
					return result, nil
				}
				return result, err
			},
		)
		fn.Doc = doc
		registerAndBind(name, fn)
	}

	// ===== Group 5: Series operations (5 functions) =====
	registerSimpleSeriesFunc("first", First, 1, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the first element of a series",
		Description: `Gets the first element of a block or string. Raises an error if the series is empty.
For strings, returns the first character as a string.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to get the first element from", Optional: false},
		},
		Returns:  "[any-type!] The first element of the series",
		Examples: []string{"first [1 2 3]  ; => 1", `first "hello"  ; => "h"`, "first [[a b] c]  ; => [a b]"},
		SeeAlso:  []string{"last", "length?", "append", "insert"}, Tags: []string{"series", "access", "first"},
	})
	registerSimpleSeriesFunc("last", Last, 1, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the last element of a series",
		Description: `Gets the last element of a block or string. Raises an error if the series is empty.
For strings, returns the last character as a string.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to get the last element from", Optional: false},
		},
		Returns:  "[any-type!] The last element of the series",
		Examples: []string{"last [1 2 3]  ; => 3", `last "hello"  ; => "o"`, "last [[a b] c]  ; => c"},
		SeeAlso:  []string{"first", "length?", "append", "insert"}, Tags: []string{"series", "access", "last"},
	})
	registerSimpleSeriesFunc("append", Append, 2, &NativeDoc{
		Category: "Series",
		Summary:  "Appends a value to the end of a series",
		Description: `Adds a value to the end of a block or string, modifying the series in place.
Returns the modified series. For strings, the value is converted to a string before appending.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to append to (modified in place)", Optional: false},
			{Name: "value", Type: "any-type!", Description: "The value to append", Optional: false},
		},
		Returns:  "[block! string!] The modified series",
		Examples: []string{"data: [1 2 3]\nappend data 4  ; => [1 2 3 4]", `text: "hello"\nappend text " world"  ; => "hello world"`},
		SeeAlso:  []string{"insert", "first", "last", "length?"}, Tags: []string{"series", "modification", "append"},
	})
	registerSimpleSeriesFunc("insert", Insert, 2, &NativeDoc{
		Category: "Series",
		Summary:  "Inserts a value at the beginning of a series",
		Description: `Adds a value to the start of a block or string, modifying the series in place.
Returns the modified series. For strings, the value is converted to a string before inserting.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to insert into (modified in place)", Optional: false},
			{Name: "value", Type: "any-type!", Description: "The value to insert", Optional: false},
		},
		Returns:  "[block! string!] The modified series",
		Examples: []string{"data: [2 3 4]\ninsert data 1  ; => [1 2 3 4]", `text: "world"\ninsert text "hello "  ; => "hello world"`},
		SeeAlso:  []string{"append", "first", "last", "length?"}, Tags: []string{"series", "modification", "insert"},
	})
	registerSimpleSeriesFunc("length?", LengthQ, 1, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the number of elements in a series",
		Description: `Counts the elements in a block or characters in a string.
Returns an integer representing the length of the series.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string!", Description: "The series to measure", Optional: false},
		},
		Returns:  "[integer!] The number of elements in the series",
		Examples: []string{"length? [1 2 3 4]  ; => 4", `length? "hello"  ; => 5`, "length? []  ; => 0"},
		SeeAlso:  []string{"first", "last", "append", "insert"}, Tags: []string{"series", "query", "length", "count"},
	})
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
			func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := Copy(args, refValues)
				if err == nil {
					return result, nil
				}
				return result, err
			},
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
		registerAndBind("copy", fn)
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
			func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := Find(args, refValues)
				if err == nil {
					return result, nil
				}
				return result, err
			},
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
		registerAndBind("find", fn)
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
			func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := Remove(args, refValues)
				if err == nil {
					return result, nil
				}
				return result, err
			},
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
		registerAndBind("remove", fn)
	}
}
