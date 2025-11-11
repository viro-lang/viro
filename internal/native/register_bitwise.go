package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func RegisterBitwiseNatives(rootFrame core.Frame) {
	bitFrame := frame.NewFrame(frame.FrameObject, -1)
	andFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.and",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		BitAnd,
		true,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Performs bitwise AND operation",
			Description: `Performs bitwise AND on two values of the same type.

For integers: standard bitwise AND using two's complement.
For binaries: byte-by-byte AND from right (LSB first), zeros remaining bytes from longer operand.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! binary!", Description: "First operand", Optional: false},
				{Name: "right", Type: "integer! binary!", Description: "Second operand (must match left type)", Optional: false},
			},
			Returns: "Same type as input",
			Examples: []string{
				"2 bit.and 3  ; => 2",
				"#{FF00} bit.and #{0FF0}  ; => #{0F00}",
				"#{FFFF} bit.and #{FF}  ; => #{00FF}",
			},
			SeeAlso: []string{"bit.or", "bit.xor", "bit.not"},
			Tags:    []string{"bitwise", "logic"},
		},
	))

	bitFrame.Bind("and", andFunc)

	orFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.or",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		BitOr,
		true,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Performs bitwise OR operation",
			Description: `Performs bitwise OR on two values of the same type.

For integers: standard bitwise OR using two's complement.
For binaries: byte-by-byte OR from right (LSB first), copies remaining bytes from longer operand.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! binary!", Description: "First operand", Optional: false},
				{Name: "right", Type: "integer! binary!", Description: "Second operand (must match left type)", Optional: false},
			},
			Returns: "Same type as input",
			Examples: []string{
				"2 bit.or 4  ; => 6",
				"#{0F00} bit.or #{F00F}  ; => #{FF0F}",
				"#{FFFF} bit.or #{FF}  ; => #{FFFF}",
			},
			SeeAlso: []string{"bit.and", "bit.xor", "bit.not"},
			Tags:    []string{"bitwise", "logic"},
		},
	))

	bitFrame.Bind("or", orFunc)

	xorFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.xor",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		BitXor,
		true,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Performs bitwise XOR operation",
			Description: `Performs bitwise XOR on two values of the same type.

For integers: standard bitwise XOR using two's complement.
For binaries: byte-by-byte XOR from right (LSB first), copies remaining bytes from longer operand.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! binary!", Description: "First operand", Optional: false},
				{Name: "right", Type: "integer! binary!", Description: "Second operand (must match left type)", Optional: false},
			},
			Returns: "Same type as input",
			Examples: []string{
				"6 bit.xor 3  ; => 5",
				"#{FF00} bit.xor #{0FF0}  ; => #{F0F0}",
				"#{FFFF} bit.xor #{FF}  ; => #{FF00}",
			},
			SeeAlso: []string{"bit.and", "bit.or", "bit.not"},
			Tags:    []string{"bitwise", "logic"},
		},
	))

	bitFrame.Bind("xor", xorFunc)

	notFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.not",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
		},
		BitNot,
		false,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Performs bitwise NOT operation",
			Description: `Performs bitwise NOT (complement) on a value.

For integers: bitwise complement using two's complement.
For binaries: flips all bits in all bytes.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! binary!", Description: "Value to complement", Optional: false},
			},
			Returns: "Same type as input",
			Examples: []string{
				"bit.not 0  ; => -1",
				"bit.not #{FF}  ; => #{00}",
				"bit.not #{00FF}  ; => #{FF00}",
			},
			SeeAlso: []string{"bit.and", "bit.or", "bit.xor"},
			Tags:    []string{"bitwise", "logic"},
		},
	))

	bitFrame.Bind("not", notFunc)

	shlFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.shl",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
			value.NewParamSpec("count", true),
		},
		BitShl,
		true,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Shifts bits left",
			Description: `Shifts bits to the left by the specified count.

For integers: Standard left shift using Go's << operator.
For binaries: Shifts all bytes left within the series boundaries.
  - Overflow beyond the series length is lost.
  - No new bytes are created.
  - Result has same length as input.

Left shift by N positions is equivalent to multiplying by 2^N for integers.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! binary!", Description: "Value to shift", Optional: false},
				{Name: "count", Type: "integer!", Description: "Number of bit positions to shift (must be non-negative)", Optional: false},
			},
			Returns: "Same type as input value",
			Examples: []string{
				"1 bit.shl 3  ; => 8",
				"#{01} bit.shl 2  ; => #{04}",
				"#{80} bit.shl 1  ; => #{00} (overflow lost)",
				"#{0100} bit.shl 8  ; => #{0001} (multi-byte shift)",
			},
			SeeAlso: []string{"bit.shr", "<<", ">>"},
			Tags:    []string{"bitwise", "shift"},
		},
	))
	bitFrame.Bind("shl", shlFunc)

	shrFunc := value.NewFuncVal(value.NewNativeFunction(
		"bit.shr",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
			value.NewParamSpec("count", true),
		},
		BitShr,
		true,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Shifts bits right",
			Description: `Shifts bits to the right by the specified count.

For integers: Arithmetic right shift (sign-extending) using Go's >> operator.
For binaries: Shifts all bytes right within the series boundaries.
  - Underflow beyond the series length is lost.
  - No new bytes are created.
  - Result has same length as input.

Right shift by N positions is equivalent to dividing by 2^N for integers (with truncation).`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! binary!", Description: "Value to shift", Optional: false},
				{Name: "count", Type: "integer!", Description: "Number of bit positions to shift (must be non-negative)", Optional: false},
			},
			Returns: "Same type as input value",
			Examples: []string{
				"8 bit.shr 2  ; => 2",
				"#{08} bit.shr 2  ; => #{02}",
				"#{01} bit.shr 1  ; => #{00} (underflow lost)",
				"#{0080} bit.shr 8  ; => #{8000} (multi-byte shift)",
			},
			SeeAlso: []string{"bit.shl", "<<", ">>"},
			Tags:    []string{"bitwise", "shift"},
		},
	))
	bitFrame.Bind("shr", shrFunc)

	bitFrame.Bind("on", value.NewFuncVal(value.NewNativeFunction(
		"bit.on",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
			value.NewParamSpec("position", true),
		},
		BitOn,
		false,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Sets a specific bit to 1",
			Description: `Sets the bit at the specified position to 1 in an integer value.
Returns the modified integer with the bit set.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer!", Description: "Integer value to modify", Optional: false},
				{Name: "position", Type: "integer!", Description: "Bit position to set (0-based)", Optional: false},
			},
			Returns: "integer! Modified value with bit set",
			Examples: []string{
				"bit.on 0 0  ; => 1",
				"bit.on 0 3  ; => 8",
				"bit.on 5 0  ; => 5 (already set)",
			},
			SeeAlso: []string{"bit.off", "bit.count"},
			Tags:    []string{"bitwise", "bit-manipulation"},
		},
	)))

	bitFrame.Bind("off", value.NewFuncVal(value.NewNativeFunction(
		"bit.off",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
			value.NewParamSpec("position", true),
		},
		BitOff,
		false,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Clears a specific bit to 0",
			Description: `Clears the bit at the specified position to 0 in an integer value.
Returns the modified integer with the bit cleared.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer!", Description: "Integer value to modify", Optional: false},
				{Name: "position", Type: "integer!", Description: "Bit position to clear (0-based)", Optional: false},
			},
			Returns: "integer! Modified value with bit cleared",
			Examples: []string{
				"bit.off 1 0  ; => 0",
				"bit.off 15 3  ; => 7",
				"bit.off 4 0  ; => 4 (already clear)",
			},
			SeeAlso: []string{"bit.on", "bit.count"},
			Tags:    []string{"bitwise", "bit-manipulation"},
		},
	)))

	bitFrame.Bind("count", value.NewFuncVal(value.NewNativeFunction(
		"bit.count",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
		},
		BitCount,
		false,
		&NativeDoc{
			Category: "Bitwise",
			Summary:  "Counts set bits (1-bits)",
			Description: `Counts the number of bits set to 1 in the value.

For integers: Counts set bits in the 64-bit two's complement representation.
For binaries: Counts set bits across all bytes in the binary series.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! binary!", Description: "Value to count bits in", Optional: false},
			},
			Returns: "integer! Number of set bits",
			Examples: []string{
				"bit.count 0  ; => 0",
				"bit.count 15  ; => 4 (0b1111)",
				"bit.count -1  ; => 64 (all bits set)",
				"bit.count #{FF}  ; => 8",
				"bit.count #{FF000F}  ; => 12",
			},
			SeeAlso: []string{"bit.on", "bit.off"},
			Tags:    []string{"bitwise", "counting"},
		},
	)))

	bitObj := value.NewObject(bitFrame)

	rootFrame.Bind("bit", bitObj)

	rootFrame.Bind("<<", shlFunc)
	rootFrame.Bind(">>", shrFunc)

	rootFrame.Bind("&", andFunc)
	rootFrame.Bind("|", orFunc)
	rootFrame.Bind("^", xorFunc)
}
