// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// RegisterIONatives registers all I/O and port-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterIONatives(rootFrame core.Frame, eval core.Evaluator) {
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

		// Bind to root frame
		rootFrame.Bind(name, value.NewFuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// Helper function to wrap simple I/O functions (no evaluator needed)
	registerSimpleIOFunc := func(name string, impl core.NativeFunc, arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := range arity {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"value", "source", "target", "file", "spec", "data", "port"}
			for i := range arity {
				if i < len(paramNames) {
					params[i] = value.NewParamSpec(paramNames[i], true)
				} else {
					params[i] = value.NewParamSpec("arg", true)
				}
			}
		}

		registerAndBind(name, value.NewNativeFunction(
			name,
			params,
			impl,
			false,
			doc,
		))
	}

	// ===== Group 8: I/O operations (3 functions - print/prin need evaluator) =====
	registerAndBind("print", value.NewNativeFunction(
		"print",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		Print,
		false,
		&NativeDoc{
			Category: "I/O",
			Summary:  "Prints a value to standard output",
			Description: `Evaluates and prints a value to standard output, followed by a newline.
 Blocks are formatted with spaces between elements. Returns none.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to print (will be evaluated)", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`print "Hello, world!"  ; prints: Hello, world!`, "print 42  ; prints: 42", "print [1 2 3]  ; prints: 1 2 3"},
			SeeAlso:  []string{"prin", "input"}, Tags: []string{"io", "output", "print", "display"},
		},
	))

	registerAndBind("prin", value.NewNativeFunction(
		"prin",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		Prin,
		false,
		&NativeDoc{
			Category: "I/O",
			Summary:  "Prints a value to standard output without newline",
			Description: `Evaluates and prints a value to standard output without a trailing newline.
 Blocks are formatted with spaces between elements. Returns none.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to print (will be evaluated)", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`prin "Hello, world!"  ; prints: Hello, world! (no newline)`, "prin 42  ; prints: 42 (no newline)", "prin [1 2 3]  ; prints: 1 2 3 (no newline)"},
			SeeAlso:  []string{"print", "input"}, Tags: []string{"io", "output", "prin", "display"},
		},
	))

	registerSimpleIOFunc("input", Input, 0, &NativeDoc{
		Category: "I/O",
		Summary:  "Reads a line of text from standard input",
		Description: `Reads a line of text from standard input (stdin) and returns it as a string.
The trailing newline is removed. Blocks until input is received.`,
		Parameters: []ParamDoc{},
		Returns:    "[string!] The line of text read from standard input",
		Examples:   []string{`name: input  ; waits for user input`, `print "Enter your name:"\nname: input\nprint ["Hello" name]`},
		SeeAlso:    []string{"print", "prin", "read"}, Tags: []string{"io", "input", "stdin", "read"},
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

	registerAndBind("read", value.NewNativeFunction(
		"read",
		[]value.ParamSpec{
			value.NewParamSpec("source", true),       // evaluated
			value.NewRefinementSpec("binary", false), // --binary flag
			value.NewRefinementSpec("lines", false),  // --lines flag
			value.NewRefinementSpec("part", true),    // --part length
			value.NewRefinementSpec("seek", true),    // --seek index
			value.NewRefinementSpec("as", true),      // --as encoding
		},
		ReadNative,
		false,
		&NativeDoc{
			Category: "Ports",
			Summary:  "Reads data from a port or file",
			Description: `Reads all data from a port or directly from a file path.
If given a port, reads from that open port. If given a string (file path),
opens the file, reads its contents, and closes it automatically.
Returns the data as a string by default, or as binary! when --binary is used.

Refinements:
  --binary: Return data as binary! instead of string!
  --lines: Return block of lines instead of single string
  --part length: Read only specified number of units (bytes or lines)
  --seek index: Start reading from specific byte position
  --as encoding: Read with specified encoding (default: utf-8)`,
			Parameters: []ParamDoc{
				{Name: "source", Type: "port! string!", Description: "A port or file path to read from", Optional: false},
			},
			Returns: "[string! binary! block!] The data read from the source",
			Examples: []string{
				`content: read "file://data.txt"  ; read as string`,
				`data: read --binary "file://image.png"  ; read as binary`,
				`lines: read --lines "file://data.txt"  ; read as block of lines`,
				`partial: read --part 100 "file://data.txt"  ; read first 100 bytes`,
				`lines: read --lines --part 5 "file://data.txt"  ; read first 5 lines`,
				`data: read --seek 1000 "file://data.txt"  ; read from byte 1000`,
				`p: open "file://data.txt"\ndata: read p\nclose p`,
			},
			SeeAlso: []string{"write", "load", "open", "close"}, Tags: []string{"ports", "io", "read", "file", "binary"},
		},
	))

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

	// ===== Group 10: Parser operations (4 functions) =====
	registerSimpleIOFunc("tokenize", NativeTokenize, 1, &NativeDoc{
		Category: "Parser",
		Summary:  "Tokenizes a viro source string into token objects",
		Description: `Tokenizes a viro source code string and returns a block of token objects.
Each token object has fields: type (word), value (string), line (integer), column (integer).
This is the first stage of the two-stage parser.`,
		Parameters: []ParamDoc{
			{Name: "source", Type: "string!", Description: "The viro source code to tokenize", Optional: false},
		},
		Returns:  "[block!] A block of token objects",
		Examples: []string{`tokens: tokenize "x: 42"  ; => [object! object! object!]`, `tokens: tokenize "[1 2 3]"`},
		SeeAlso:  []string{"parse", "load-string", "classify"}, Tags: []string{"parser", "tokenize", "lexer"},
	})

	registerSimpleIOFunc("parse", NativeParse, 1, &NativeDoc{
		Category: "Parser",
		Summary:  "Parses token objects into viro values",
		Description: `Takes a block of token objects (from tokenize) and parses them into viro values.
This is the second stage of the two-stage parser. Returns a block of parsed values.`,
		Parameters: []ParamDoc{
			{Name: "tokens", Type: "block!", Description: "A block of token objects from tokenize", Optional: false},
		},
		Returns:  "[block!] A block of parsed viro values",
		Examples: []string{`tokens: tokenize "x: 42"\nvalues: parse tokens  ; => [x: 42]`, `values: parse tokenize "[1 2 3]"`},
		SeeAlso:  []string{"tokenize", "load-string", "classify"}, Tags: []string{"parser", "parse", "semantic"},
	})

	registerSimpleIOFunc("load-string", NativeLoadString, 1, &NativeDoc{
		Category: "Parser",
		Summary:  "Parses a viro source string directly into values",
		Description: `Combines tokenize and parse in one step. Takes a source code string,
tokenizes it, parses it, and returns a block of viro values. This is equivalent to
calling parse on the result of tokenize.`,
		Parameters: []ParamDoc{
			{Name: "source", Type: "string!", Description: "The viro source code to parse", Optional: false},
		},
		Returns:  "[block!] A block of parsed viro values",
		Examples: []string{`values: load-string "x: 42"  ; => [x: 42]`, `values: load-string "[1 2 3]"  ; => [[1 2 3]]`},
		SeeAlso:  []string{"tokenize", "parse", "classify", "load"}, Tags: []string{"parser", "load", "parse"},
	})

	registerSimpleIOFunc("classify", NativeClassify, 1, &NativeDoc{
		Category: "Parser",
		Summary:  "Classifies a literal string into its viro value type",
		Description: `Takes a literal string and determines what viro type it represents,
returning the corresponding typed value. For example, "42" becomes integer! 42,
"true" becomes logic! true, etc. This is useful for dynamic type conversion.`,
		Parameters: []ParamDoc{
			{Name: "literal", Type: "string!", Description: "The literal string to classify", Optional: false},
		},
		Returns:  "[any-type!] The classified viro value",
		Examples: []string{`classify "42"  ; => 42`, `classify "true"  ; => true`, `classify "hello"  ; => hello (word)`, `classify ":x"  ; => :x (get-word)`},
		SeeAlso:  []string{"tokenize", "parse", "load-string"}, Tags: []string{"parser", "classify", "type", "conversion"},
	})

	// Create and bind standard I/O ports
	stdoutPort := value.NewPort("stdio", "stdout", &stdioWriterDriver{writer: eval.GetOutputWriter()})
	stderrPort := value.NewPort("stdio", "stderr", &stdioWriterDriver{writer: eval.GetErrorWriter()})
	stdinPort := value.NewPort("stdio", "stdin", &stdioReaderDriver{reader: eval.GetInputReader()})

	rootFrame.Bind("stdout", value.PortVal(stdoutPort))
	rootFrame.Bind("stderr", value.PortVal(stderrPort))
	rootFrame.Bind("stdin", value.PortVal(stdinPort))
}
