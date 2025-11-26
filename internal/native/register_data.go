// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func RegisterDataNatives(rootFrame core.Frame) {
	registered := make(map[string]bool)

	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterDataNatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterDataNatives: duplicate registration of function '%s'", name))
		}

		rootFrame.Bind(name, value.NewFuncVal(fn))
		registered[name] = true
	}

	rootFrame.Bind("true", value.NewLogicVal(true))
	rootFrame.Bind("false", value.NewLogicVal(false))
	rootFrame.Bind("none", value.NewNoneVal())
	registerAndBind("set", value.NewNativeFunction(
		"set",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
			value.NewParamSpec("value", true), // evaluated
		},
		Set,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Sets a word to a value in the current context",
			Description: `Assigns a value to a word (variable) in the current frame.
The word is not evaluated; the value is evaluated before assignment. Returns the assigned value.`,
			Parameters: []ParamDoc{
				{Name: "word", Type: "word!", Description: "The word to set (not evaluated)", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The value to assign (evaluated)", Optional: false},
			},
			Returns:  "[any-type!] The value that was assigned",
			Examples: []string{"set 'x 42  ; => 42 (x is now 42)", "set 'name \"Alice\"  ; => \"Alice\"", "set 'data [1 2 3]  ; => [1 2 3]"},
			SeeAlso:  []string{"get", ":", "type?"}, Tags: []string{"data", "assignment", "variable"},
		},
	))

	registerAndBind("get", value.NewNativeFunction(
		"get",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
		},
		Get,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Gets the value of a word from the current context",
			Description: `Retrieves the value associated with a word (variable) in the current frame.
The word is not evaluated. Raises an error if the word is not bound to a value.`,
			Parameters: []ParamDoc{
				{Name: "word", Type: "word!", Description: "The word to look up (not evaluated)", Optional: false},
			},
			Returns:  "[any-type!] The value bound to the word",
			Examples: []string{"x: 42\nget 'x  ; => 42", "name: \"Bob\"\nget 'name  ; => \"Bob\""},
			SeeAlso:  []string{"set", ":", "type?"}, Tags: []string{"data", "access", "variable"},
		},
	))

	registerAndBind("type?", value.NewNativeFunction(
		"type?",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		TypeQ,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Returns the type of a value",
			Description: `Returns a word representing the type of the given value.
Possible types include: integer!, decimal!, string!, block!, word!, function!, object!, port!, logic!, none!`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to check the type of", Optional: false},
			},
			Returns:  "[word!] A word representing the value's type",
			Examples: []string{"type? 42  ; => integer!", `type? "hello"  ; => string!`, "type? [1 2 3]  ; => block!", "type? :print  ; => function!"},
			SeeAlso:  []string{"set", "get"}, Tags: []string{"data", "type", "introspection", "reflection"},
		},
	))

	registerAndBind("none?", value.NewNativeFunction(
		"none?",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		NoneQ,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Returns true if value is none",
			Description: `Returns true if the given value is none, false otherwise.
This is a type predicate function that checks if a value represents the absence of a value.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to check", Optional: false},
			},
			Returns:  "[logic!] True if value is none, false otherwise",
			Examples: []string{"none? none  ; => true", "none? 42  ; => false", `none? "hello"  ; => false`},
			SeeAlso:  []string{"type?"}, Tags: []string{"data", "type", "predicate", "none"},
		},
	))

	registerAndBind("form", value.NewNativeFunction(
		"form",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		Form,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Converts a value to a human-readable string",
			Description: `Returns a human-readable string representation of the value.
For blocks, omits outer brackets. For strings, omits quotes. Does not evaluate block contents.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to convert to string", Optional: false},
			},
			Returns:  "[string!] Human-readable string representation",
			Examples: []string{"form [1 2 3]  ; => \"1 2 3\"", `form "hello"  ; => "hello"`, "form 42  ; => \"42\""},
			SeeAlso:  []string{"mold", "type?"}, Tags: []string{"data", "string", "formatting"},
		},
	))

	registerAndBind("join", value.NewNativeFunction(
		"join",
		[]value.ParamSpec{
			value.NewParamSpec("value1", true), // evaluated
			value.NewParamSpec("value2", true), // evaluated
		},
		Join,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Concatenates two values into a string",
			Description: `Converts both values to strings using form and concatenates them.
Automatically converts numbers, blocks, and other types to their string representations.`,
			Parameters: []ParamDoc{
				{Name: "value1", Type: "any-type!", Description: "First value to concatenate", Optional: false},
				{Name: "value2", Type: "any-type!", Description: "Second value to concatenate", Optional: false},
			},
			Returns:  "[string!] Concatenated string",
			Examples: []string{`join "Hello" " World"  ; => "Hello World"`, `join "Number: " 42  ; => "Number: 42"`, `join "x: " [1 2 3]  ; => "x: 1 2 3"`},
			SeeAlso:  []string{"form", "mold"}, Tags: []string{"data", "string", "concatenation"},
		},
	))

	registerAndBind("rejoin", value.NewNativeFunction(
		"rejoin",
		[]value.ParamSpec{
			value.NewParamSpec("block", true),
		},
		Rejoin,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Evaluates a block and joins all results into a string",
			Description: `Evaluates each element in the block and concatenates all results into a single string without any separator.
This is equivalent to calling reduce on a block and then joining all results with no separator.`,
			Parameters: []ParamDoc{
				{Name: "block", Type: "block!", Description: "The block containing values to evaluate and join", Optional: false},
			},
			Returns: "[string!] Concatenated string of all evaluated values",
			Examples: []string{
				`rejoin ["Hello" " " "World"]  ; => "Hello World"`,
				`rejoin ["Number: " 42]  ; => "Number: 42"`,
				`rejoin ["Result: " 10 + 5]  ; => "Result: 15"`,
				`rejoin []  ; => ""`,
			},
			SeeAlso: []string{"join", "reduce", "form"},
			Tags:    []string{"data", "string", "concatenation", "evaluation"},
		},
	))

	registerAndBind("mold", value.NewNativeFunction(
		"mold",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		Mold,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Converts a value to a code-readable string",
			Description: `Returns a code-readable string representation of the value.
For blocks, includes outer brackets. For strings, includes quotes. Does not evaluate block contents.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to convert to string", Optional: false},
			},
			Returns:  "[string!] code-readable string representation",
			Examples: []string{"mold [1 2 3]  ; => \"[1 2 3]\"", `mold "hello"  ; => "\"hello\""`, "mold 42  ; => \"42\""},
			SeeAlso:  []string{"form", "type?"}, Tags: []string{"data", "string", "formatting", "serialization"},
		},
	))

	registerAndBind("reduce", value.NewNativeFunction(
		"reduce",
		[]value.ParamSpec{
			value.NewParamSpec("block", true),
		},
		Reduce,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Evaluates each element in a block and returns the results as a block",
			Description: `Takes a block and evaluates each element individually, collecting the results
into a new block. This is useful for computing values dynamically and building data structures.`,
			Parameters: []ParamDoc{
				{Name: "block", Type: "block!", Description: "The block containing elements to evaluate", Optional: false},
			},
			Returns:  "[block!] A new block containing the evaluated results",
			Examples: []string{"reduce [1 2 3]  ; => [1 2 3]", "reduce [1 + 2, 3 * 4]  ; => [3, 12]", "reduce []  ; => []"},
			SeeAlso:  []string{"form", "mold"}, Tags: []string{"data", "evaluation", "block", "reduce"},
		},
	))

	// ===== Group 7: Object operations (5 functions - all need evaluator) =====
	registerAndBind("object", value.NewNativeFunction(
		"object",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		Object,
		false,
		&NativeDoc{
			Category: "Objects",
			Summary:  "Creates a new object from a block of definitions",
			Description: `Creates a new object (context) by evaluating a block of word-value pairs.
The block is evaluated in a new frame, and all word definitions become fields of the object.
Returns the newly created object.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "block!", Description: "A block containing word definitions to become object fields", Optional: false},
			},
			Returns:  "[object!] The newly created object",
			Examples: []string{"obj: object [x: 10 y: 20]  ; => object with fields x and y", "person: object [name: \"Alice\" age: 30]"},
			SeeAlso:  []string{"context", "make"}, Tags: []string{"objects", "context", "creation"},
		},
	))

	registerAndBind("context", value.NewNativeFunction(
		"context",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		Context,
		false,
		&NativeDoc{
			Category: "Objects",
			Summary:  "Creates a new context (alias for object)",
			Description: `Creates a new object (context) by evaluating a block of word-value pairs.
This is an alias for the 'object' function. The block is evaluated in a new frame,
and all word definitions become fields of the context.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "block!", Description: "A block containing word definitions to become context fields", Optional: false},
			},
			Returns:  "[object!] The newly created context",
			Examples: []string{"ctx: context [counter: 0 increment: fn [] [counter: counter + 1]]", "config: context [debug: true port: 8080]"},
			SeeAlso:  []string{"object", "make"}, Tags: []string{"objects", "context", "creation"},
		},
	))

	registerAndBind("make", value.NewNativeFunction(
		"make",
		[]value.ParamSpec{
			value.NewParamSpec("parent", true), // evaluated
			value.NewParamSpec("spec", false),  // NOT evaluated (block)
		},
		Make,
		false,
		&NativeDoc{
			Category: "Objects",
			Summary:  "Creates a derived object from a parent object",
			Description: `Creates a new object that inherits from a parent object and adds or overrides fields.
The first argument is the parent object (prototype), and the second is a block of
new or overriding field definitions. The new object shares the parent's fields but can shadow them.`,
			Parameters: []ParamDoc{
				{Name: "parent", Type: "object!", Description: "The parent object to derive from", Optional: false},
				{Name: "spec", Type: "block!", Description: "A block of field definitions to add or override", Optional: false},
			},
			Returns:  "[object!] The newly created derived object",
			Examples: []string{"base: object [x: 1 y: 2]\nderived: make base [z: 3]  ; => object with x, y, z", "point: object [x: 0 y: 0]\npoint3d: make point [z: 0]"},
			SeeAlso:  []string{"object", "context"}, Tags: []string{"objects", "inheritance", "derivation"},
		},
	))

	RegisterActionImpl(value.TypeObject, "select", value.NewNativeFunction(
		"select",
		[]value.ParamSpec{
			value.NewParamSpec("target", true), // evaluated
			value.NewParamSpec("field", false), // NOT evaluated (word/string)
			value.NewRefinementSpec("default", true),
		},
		Select,
		false,
		nil)) // No doc needed since it's type-specific

	registerAndBind("put", value.NewNativeFunction(
		"put",
		[]value.ParamSpec{
			value.NewParamSpec("target", true),
			value.NewParamSpec("key", true),
			value.NewParamSpec("value", true),
		},
		Put,
		false,
		&NativeDoc{
			Category: "Objects",
			Summary:  "Sets a field value in an object or key/value pair in a block",
			Description: `For objects: Sets or updates an object field with a new value.
Creates the field if it doesn't exist, allowing dynamic field addition.
If the field has a type hint, the new value must match that type.

For blocks: Treats the block as an association list of alternating key/value pairs.
Updates or appends key/value pairs, or removes pairs when value is none.
Respects the block's current index when searching for keys.
Keys are matched using the same logic as select (word-like symbol comparison or general equality).
Returns the assigned value (or none for removal).`,
			Parameters: []ParamDoc{
				{Name: "target", Type: "object! or block!", Description: "The object or block to modify", Optional: false},
				{Name: "key", Type: "any-type!", Description: "The field name (for objects) or key (for blocks) to update", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The new value to assign (use none to remove from blocks)", Optional: false},
			},
			Returns: "[any-type!] The assigned value",
			Examples: []string{
				"obj: object [x: 10 y: 20]\nput obj 'x 42  ; => 42, obj.x is now 42",
				"blk: [a 1 b 2]\nput blk 'a 99  ; => 99, blk is now [a 99 b 2]",
				"blk: [a 1 b 2]\nput blk 'c 3  ; => 3, blk is now [a 1 b 2 c 3]",
				"blk: [a 1 b 2]\nput blk 'a none  ; => none, blk is now [b 2]",
			},
			SeeAlso: []string{"select", "set", "object"},
			Tags:    []string{"objects", "blocks", "mutation", "field-update", "association-lists"},
		},
	))

	registerAndBind("to-integer", value.NewNativeFunction(
		"to-integer",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		ToInteger,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Converts a value to an integer",
			Description: `Converts integer (pass-through), decimal (truncate), or string (parse) to integer.
Decimal values are truncated towards zero. String values must contain valid integer format.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal! string!", Description: "The value to convert", Optional: false},
			},
			Returns:  "[integer!] The converted integer value",
			Examples: []string{"to-integer 42  ; => 42", "to-integer 3.7  ; => 3", `to-integer "123"  ; => 123`},
			SeeAlso:  []string{"to-decimal", "to-string", "type?"},
			Tags:     []string{"data", "conversion", "type"},
		},
	))

	registerAndBind("to-decimal", value.NewNativeFunction(
		"to-decimal",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		ToDecimal,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Converts a value to a decimal",
			Description: `Converts integer (exact), decimal (pass-through), or string (parse) to decimal.
Integer values are converted to exact decimal representation. String values must contain valid decimal format.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal! string!", Description: "The value to convert", Optional: false},
			},
			Returns:  "[decimal!] The converted decimal value",
			Examples: []string{"to-decimal 42  ; => 42.0", "to-decimal 3.7  ; => 3.7", `to-decimal "12.34"  ; => 12.34`},
			SeeAlso:  []string{"to-integer", "to-string", "type?"},
			Tags:     []string{"data", "conversion", "type"},
		},
	))

	registerAndBind("to-string", value.NewNativeFunction(
		"to-string",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		ToString,
		false,
		&NativeDoc{
			Category: "Data",
			Summary:  "Converts a value to a string",
			Description: `Converts any value to a string using its Form representation.
For strings, returns the value unchanged. For integers and decimals, returns string representation.
For blocks and other types, returns their human-readable form.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to convert", Optional: false},
			},
			Returns:  "[string!] The converted string value",
			Examples: []string{"to-string 42  ; => \"42\"", "to-string 3.7  ; => \"3.70\"", `to-string [1 2 3]  ; => "1 2 3"`},
			SeeAlso:  []string{"to-integer", "to-decimal", "form", "mold"},
			Tags:     []string{"data", "conversion", "type", "string"},
		},
	))
}
