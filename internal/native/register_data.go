// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// RegisterDataNatives registers all data and object-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterDataNatives(rootFrame *frame.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function
	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterDataNatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterDataNatives: duplicate registration of function '%s'", name))
		}

		// Add to global Registry for backward compatibility
		Registry[name] = fn

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// Helper function to wrap simple data functions
	registerSimpleDataFunc := func(name string, impl func([]value.Value) (value.Value, *verror.Error), arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := 0; i < arity; i++ {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"value", "word", "spec"}
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

	// ===== Group 6: Data operations (3 functions) =====
	// set and get need evaluator, type? doesn't
	fn := value.NewNativeFunction(
		"set",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
			value.NewParamSpec("value", true), // evaluated
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			// We need to pass a native.Evaluator to Set, but we have value.Evaluator
			// Create a reverse adapter that converts value.Evaluator back to native.Evaluator
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Set(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("set", fn)

	fn = value.NewNativeFunction(
		"get",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Get(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("get", fn)

	registerSimpleDataFunc("type?", TypeQ, 1, &NativeDoc{
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
	})

	// ===== Group 7: Object operations (5 functions - all need evaluator) =====
	fn = value.NewNativeFunction(
		"object",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Object(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("object", fn)

	fn = value.NewNativeFunction(
		"context",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Context(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("context", fn)

	fn = value.NewNativeFunction(
		"make",
		[]value.ParamSpec{
			value.NewParamSpec("parent", true), // evaluated
			value.NewParamSpec("spec", false),  // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Make(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("make", fn)

	fn = value.NewNativeFunction(
		"select",
		[]value.ParamSpec{
			value.NewParamSpec("target", true), // evaluated
			value.NewParamSpec("field", false), // NOT evaluated (word/string)
			value.NewRefinementSpec("default", true),
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Select(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Objects",
		Summary:  "Retrieves a field value from an object or block",
		Description: `Looks up a field in an object or searches for a key in a block.
For objects: returns the field value, checking parent prototypes if needed.
For blocks: searches for key-value pairs (alternating pattern) and returns the value.
Use --default refinement to provide a fallback when field/key is not found.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "object! or block!", Description: "The object or block to search", Optional: false},
			{Name: "field", Type: "word! or string!", Description: "The field name or key to look up", Optional: false},
			{Name: "--default", Type: "any-type!", Description: "Optional default value when field not found", Optional: true},
		},
		Returns: "[any-type!] The field/key value, or default, or none",
		Examples: []string{
			"obj: object [x: 10 y: 20]\nselect obj 'x  ; => 10",
			"select obj 'z --default 99  ; => 99 (field not found)",
			"data: ['name \"Alice\" 'age 30]\nselect data 'age  ; => 30",
		},
		SeeAlso: []string{"put", "get", "object"},
		Tags:    []string{"objects", "lookup", "field-access"},
	}
	registerAndBind("select", fn)

	fn = value.NewNativeFunction(
		"put",
		[]value.ParamSpec{
			value.NewParamSpec("object", true), // evaluated
			value.NewParamSpec("field", false), // NOT evaluated (word/string)
			value.NewParamSpec("value", true),  // evaluated
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Put(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Objects",
		Summary:  "Sets a field value in an object",
		Description: `Updates an existing field in an object with a new value.
The field must already exist in the object's manifest - dynamic field addition is not allowed.
If the field has a type hint, the new value must match that type.
Returns the assigned value.`,
		Parameters: []ParamDoc{
			{Name: "object", Type: "object!", Description: "The object to modify", Optional: false},
			{Name: "field", Type: "word! or string!", Description: "The field name to update", Optional: false},
			{Name: "value", Type: "any-type!", Description: "The new value to assign", Optional: false},
		},
		Returns: "[any-type!] The assigned value",
		Examples: []string{
			"obj: object [x: 10 y: 20]\nput obj 'x 42  ; => 42, obj.x is now 42",
			"person: object [name: \"Alice\" age: 30]\nput person 'age 31",
		},
		SeeAlso: []string{"select", "set", "object"},
		Tags:    []string{"objects", "mutation", "field-update"},
	}
	registerAndBind("put", fn)
}
