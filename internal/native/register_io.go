// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// RegisterIONatives registers all I/O and port-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterIONatives(rootFrame *frame.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function
	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterIONatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterIONatives: duplicate registration of function '%s'", name))
		}

		// Add to global Registry for backward compatibility
		Registry[name] = fn

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// Helper function to wrap simple I/O functions (no evaluator needed)
	registerSimpleIOFunc := func(name string, impl func([]value.Value) (value.Value, *verror.Error), arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := 0; i < arity; i++ {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"value", "source", "target", "file", "spec", "data", "port"}
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

	// ===== Group 8: I/O operations (2 functions - print needs evaluator) =====
	fn := value.NewNativeFunction(
		"print",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Print(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "I/O",
		Summary:  "Prints a value to standard output",
		Description: `Evaluates and prints a value to standard output, followed by a newline.
Blocks are formatted with spaces between elements. Returns none.`,
		Parameters: []ParamDoc{
			{Name: "value", Type: "any-type!", Description: "The value to print (will be evaluated)", Optional: false},
		},
		Returns:  "[none!] Always returns none",
		Examples: []string{`print "Hello, world!"  ; prints: Hello, world!`, "print 42  ; prints: 42", "print [1 2 3]  ; prints: 1 2 3"},
		SeeAlso:  []string{"input"}, Tags: []string{"io", "output", "print", "display"},
	}
	registerAndBind("print", fn)

	registerSimpleIOFunc("input", Input, 0, &NativeDoc{
		Category: "I/O",
		Summary:  "Reads a line of text from standard input",
		Description: `Reads a line of text from standard input (stdin) and returns it as a string.
The trailing newline is removed. Blocks until input is received.`,
		Parameters: []ParamDoc{},
		Returns:    "[string!] The line of text read from standard input",
		Examples:   []string{`name: input  ; waits for user input`, `print "Enter your name:"\nname: input\nprint ["Hello" name]`},
		SeeAlso:    []string{"print", "read"}, Tags: []string{"io", "input", "stdin", "read"},
	})

	// ===== Group 9: Port operations (8 functions) =====
	registerSimpleIOFunc("open", OpenNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Opens a port for file or network I/O",
		Description: `Opens a port specified by a URL or file path string. Supports file:// URLs and
potentially other schemes. Returns a port value that can be used with read, write, close, etc.
File operations are subject to sandbox restrictions if configured.`,
		Parameters: []ParamDoc{
			{Name: "spec", Type: "string!", Description: "A URL or file path (e.g., \"file://data.txt\")", Optional: false},
		},
		Returns:  "[port!] An open port ready for I/O operations",
		Examples: []string{`p: open "file://data.txt"  ; => port`, `p: open "file:///tmp/output.log"`},
		SeeAlso:  []string{"close", "read", "write", "save", "load"}, Tags: []string{"ports", "io", "file", "open"},
	})

	registerSimpleIOFunc("close", CloseNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Closes an open port",
		Description: `Closes a previously opened port, releasing any associated resources.
After closing, the port should not be used for further I/O operations. Returns none.`,
		Parameters: []ParamDoc{
			{Name: "port", Type: "port!", Description: "The port to close", Optional: false},
		},
		Returns:  "[none!] Always returns none",
		Examples: []string{`p: open "file://data.txt"\nclose p  ; closes the port`},
		SeeAlso:  []string{"open", "read", "write"}, Tags: []string{"ports", "io", "close", "cleanup"},
	})

	registerSimpleIOFunc("read", ReadNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Reads data from a port or file",
		Description: `Reads all data from a port or directly from a file path.
If given a port, reads from that open port. If given a string (file path),
opens the file, reads its contents, and closes it automatically. Returns the data as a string.`,
		Parameters: []ParamDoc{
			{Name: "source", Type: "port! string!", Description: "A port or file path to read from", Optional: false},
		},
		Returns:  "[string!] The data read from the source",
		Examples: []string{`content: read "file://data.txt"  ; read entire file`, `p: open "file://data.txt"\ndata: read p\nclose p`},
		SeeAlso:  []string{"write", "load", "open", "close"}, Tags: []string{"ports", "io", "read", "file"},
	})

	registerSimpleIOFunc("write", WriteNative, 2, &NativeDoc{
		Category: "Ports",
		Summary:  "Writes data to a port or file",
		Description: `Writes data to a port or directly to a file path.
If the target is a port, writes to that open port. If given a string (file path),
opens the file, writes the data, and closes it automatically. Overwrites existing content.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "port! string!", Description: "A port or file path to write to", Optional: false},
			{Name: "data", Type: "string!", Description: "The data to write", Optional: false},
		},
		Returns:  "[none!] Always returns none",
		Examples: []string{`write "file://output.txt" "Hello, world!"  ; write to file`, `p: open "file://output.txt"\nwrite p "data"\nclose p`},
		SeeAlso:  []string{"read", "save", "open", "close"}, Tags: []string{"ports", "io", "write", "file"},
	})

	registerSimpleIOFunc("save", SaveNative, 2, &NativeDoc{
		Category: "Ports",
		Summary:  "Saves a value to a file in viro format",
		Description: `Serializes a viro value (block, object, etc.) and writes it to a file.
The value is converted to viro source code format that can be loaded back with 'load'.
This is the recommended way to persist viro data structures.`,
		Parameters: []ParamDoc{
			{Name: "file", Type: "string!", Description: "The file path to save to", Optional: false},
			{Name: "value", Type: "any-type!", Description: "The value to save", Optional: false},
		},
		Returns:  "[none!] Always returns none",
		Examples: []string{`save "file://config.viro" [debug: true port: 8080]`, `data: object [x: 1 y: 2]\nsave "file://data.viro" data`},
		SeeAlso:  []string{"load", "write", "read"}, Tags: []string{"ports", "io", "save", "serialize", "persist"},
	})

	registerSimpleIOFunc("load", LoadNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Loads and parses a viro source file",
		Description: `Reads a file containing viro source code, parses it, and returns the parsed value.
This is the recommended way to load data structures saved with 'save'.
Returns the parsed viro value (block, object, etc.).`,
		Parameters: []ParamDoc{
			{Name: "file", Type: "string!", Description: "The file path to load from", Optional: false},
		},
		Returns:  "[any-type!] The parsed viro value from the file",
		Examples: []string{`config: load "file://config.viro"  ; load and parse`, `data: load "file://data.viro"`},
		SeeAlso:  []string{"save", "read"}, Tags: []string{"ports", "io", "load", "parse", "deserialize"},
	})

	registerSimpleIOFunc("query", QueryNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Queries metadata about a port or file",
		Description: `Returns metadata about a port or file, such as size, modification time, or status.
The exact information returned depends on the port type. Returns an object with metadata fields.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "port! string!", Description: "A port or file path to query", Optional: false},
		},
		Returns:  "[object!] An object containing metadata about the target",
		Examples: []string{`info: query "file://data.txt"  ; get file info`, `p: open "file://data.txt"\ninfo: query p\nclose p`},
		SeeAlso:  []string{"open", "read"}, Tags: []string{"ports", "io", "metadata", "query", "info"},
	})

	registerSimpleIOFunc("wait", WaitNative, 1, &NativeDoc{
		Category: "Ports",
		Summary:  "Waits for a port to be ready or for a timeout",
		Description: `Waits for a port to become ready for I/O operations, or for a specified duration.
If given a number, waits for that many seconds. If given a port, waits until the port is ready.
Returns the port that became ready, or none if a timeout occurred.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "port! integer! decimal!", Description: "A port to wait on or a duration in seconds", Optional: false},
		},
		Returns:  "[port! none!] The ready port or none on timeout",
		Examples: []string{"wait 2  ; wait for 2 seconds", "wait 0.5  ; wait for half a second", `p: open "file://data.txt"\nwait p  ; wait until port is ready`},
		SeeAlso:  []string{"open", "read", "write"}, Tags: []string{"ports", "io", "wait", "delay", "timeout"},
	})
}
