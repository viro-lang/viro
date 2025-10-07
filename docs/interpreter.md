## Design Specification: REBOL-Inspired Interpreter for Sonnet 4.5

### Core Architecture

**Evaluation Engine**
- Implement a type-based dispatch system similar to `Do_Next()` and `Do_Blk()` [2](#0-1) 
- Create an evaluation type map that classifies values (words, functions, paths, literals) and routes them to appropriate handlers [3](#0-2) 
- Support recursive evaluation for nested blocks and parentheses <cite />

**Stack Management**
- Design a unified stack for both data values and function call frames <cite />
- Implement stack frame layout with: return value slot, prior frame info, function metadata, and arguments <cite />
- Include automatic stack expansion with index-based (not pointer-based) access for safety <cite />

**Frame & Context System**
- Create frames as the fundamental unit for variable storage [4](#0-3) 
- Each frame contains: a word list (variable names) and value list (actual data) <cite />
- Support multiple frame types: objects, modules, function arguments, closures <cite />
- Implement binding operations: `Bind_Block()`, `Get_Var()`, `Set_Var()` <cite />

### Native Functions System

**Function Categories** (600+ functions organized as):
- Control flow: `if`, `either`, `case`, `loop`, `while`, `foreach`, `try`, `catch` [5](#0-4) 
- Data operations: `bind`, `parse`, `get`, `set`, `type?` [6](#0-5) 
- I/O operations: `print`, `read`, `write`, `now` <cite />
- Math operations: `cosine`, `sine`, `sqrt`, `exp` <cite />
- Series operations: `first`, `last`, `append`, `insert` <cite />

**Function Specification Format**:
```rebol
function-name: native [
    "Description"
    arg1 [type1! type2!] "Arg description"
    /refinement "Refinement description"
    ref-arg [type!]
]
``` [7](#0-6) 

### Error Handling System

**Error Categories** (0-900 range):
- Throw (0): Loop control errors <cite />
- Note (100): Warnings <cite />
- Syntax (200): Parsing errors <cite />
- Script (300): Runtime errors <cite />
- Math (400): Arithmetic errors <cite />
- Access (500): I/O/security errors <cite />
- Internal (900): System errors <cite />

**Error Structure**:
```rebol
error: context [
    code: 0
    type: 'user
    id: 'message
    arg1: arg2: arg3: near: where: none
]
``` [8](#0-7) 

### Parse Dialect

Implement a pattern-matching DSL with:
- String and block parsing modes [9](#0-8) 
- Commands: `to`, `thru`, `copy`, `set`, `some`, `any`, `opt` <cite />
- Support for literal matching, type matching, and recursive rules <cite />
- Parse flags: case-sensitive, all-characters modes <cite />

### Module System

**Module Loading**:
- Header processing with validation and checksums [10](#0-9) 
- Delayed vs immediate loading strategies <cite />
- Export/import word management <cite />
- Context resolution and binding <cite />

**Boot Sequence**:
1. Load base constants and functions [11](#0-10) 
2. Initialize sys context [12](#0-11) 
3. Load lib context (mezzanine functions) [13](#0-12) 
4. Initialize protocols [14](#0-13) 

### Build System

**Configuration**:
- Platform database with OS targets and compiler flags [15](#0-14) 
- Support multiple platforms: Linux, Windows, OSX, embedded systems <cite />
- Compile flags: optimization levels, visibility, architecture <cite />

**Generation Process**:
- Generate C enums from native function specs <cite />
- Create dispatch tables for function calls <cite />
- Build platform-specific makefiles <cite />

### Implementation Recommendations for Sonnet 4.5

1. **Start with Core Evaluator**: Implement `Do_Next()` and `Do_Blk()` as the foundation <cite />
2. **Type System First**: Define value types and evaluation strategies before functions <cite />
3. **Minimal Native Set**: Begin with ~50 essential natives, expand incrementally <cite />
4. **Frame System Early**: Critical for variable binding and function calls <cite />
5. **Error Handling**: Implement structured errors from the start for debugging <cite />
6. **Parse Later**: Complex subsystem, add after core interpreter works <cite />

### Key Design Principles

- **Homoiconicity**: Code is data, data is code <cite />
- **Type-based dispatch**: Each value type has specific evaluation rules <cite />
- **Stack-based execution**: Unified stack for data and control flow <cite />
- **Context-based binding**: Words bound to frames, not global namespace <cite />
- **Layered architecture**: Natives → Mezzanines → User code <cite />

## Notes

This specification is based on REBOL R3's proven architecture [16](#0-15) . The system is highly tuned and should be modified carefully <cite />. The frame system [17](#0-16)  and control flow primitives [18](#0-17)  are particularly critical components. Consider starting with a subset of features and expanding based on use cases. <cite />

Wiki pages you might want to explore:
- [Overview (rebol/rebol)](/wiki/rebol/rebol#1)
- [Core Runtime Engine (rebol/rebol)](/wiki/rebol/rebol#3)
- [Language Features (rebol/rebol)](/wiki/rebol/rebol#4)

### Citations

**File:** src/core/c-do.c (L1-32)
```c
/***********************************************************************
**
**  REBOL [R3] Language Interpreter and Run-time Environment
**
**  Copyright 2012 REBOL Technologies
**  REBOL is a trademark of REBOL Technologies
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**  http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
**
************************************************************************
**
**  Module:  c-do.c
**  Summary: the core interpreter - the heart of REBOL
**  Section: core
**  Author:  Carl Sassenrath
**  Notes:
**    WARNING WARNING WARNING
**    This is highly tuned code that should only be modified by experts
**    who fully understand its design. It is very easy to create odd
**    side effects so please be careful and extensively test all changes!
**
***********************************************************************/
```

**File:** src/core/c-do.c (L649-740)
```c
	tos = DS_NEXT;
	DSP += ds;
	for (; ds > 0; ds--) SET_NONE(tos++);

	// Go thru the word list args:
	ds = dsp;
	for (; NOT_END(args); args++, ds++) {

		//if (Trace_Flags) Trace_Arg(ds - dsp, args, path);

		// Process each formal argument:
		switch (VAL_TYPE(args)) {

		case REB_WORD:		// WORD - Evaluate next value
			index = Do_Next(block, index, IS_OP(func));
			// THROWN is handled after the switch.
			if (index == END_FLAG) Trap2(RE_NO_ARG, Func_Word(dsf), args);
			DS_Base[ds] = *DS_POP;
			break;

		case REB_LIT_WORD:	// 'WORD - Just get next value
			if (index < BLK_LEN(block)) {
				value = BLK_SKIP(block, index);
				if (IS_PAREN(value) || IS_GET_WORD(value) || IS_GET_PATH(value)) {
					index = Do_Next(block, index, IS_OP(func));
					// THROWN is handled after the switch.
					DS_Base[ds] = *DS_POP;
				}
				else {
					index++;
					DS_Base[ds] = *value;
				}
			} else
				SET_UNSET(&DS_Base[ds]); // allowed to be none
			break;

		case REB_GET_WORD:	// :WORD - Get value
			if (index < BLK_LEN(block)) {
				DS_Base[ds] = *BLK_SKIP(block, index);
				index++;
			} else
				SET_UNSET(&DS_Base[ds]); // allowed to be none
			break;
/*
				value = BLK_SKIP(block, index);
				index++;
				if (IS_WORD(value) && VAL_WORD_FRAME(value)) value = Get_Var(value);
				DS_Base[ds] = *value;
*/
		case REB_REFINEMENT: // /WORD - Function refinement
			if (!path || IS_END(path)) return index;
			if (IS_WORD(path)) {
				// Optimize, if the refinement is the next arg:
				if (SAME_SYM(path, args)) {
					SET_TRUE(DS_VALUE(ds)); // set refinement stack value true
					path++;				// remove processed refinement
					continue;
				}
				// Refinement out of sequence, resequence arg order:
more_path:
				ds = dsp;
				args = BLK_SKIP(words, 1);
				for (; NOT_END(args); args++, ds++) {
					if (IS_REFINEMENT(args) && VAL_WORD_CANON(args) == VAL_WORD_CANON(path)) {
						SET_TRUE(DS_VALUE(ds)); // set refinement stack value true
						path++;				// remove processed refinement
						break;
					}
				}
				// Was refinement found? If not, error:
				if (IS_END(args)) Trap2(RE_NO_REFINE, Func_Word(dsf), path);
				continue;
			}
			else Trap1(RE_BAD_REFINE, path);
			break;

		case REB_SET_WORD:	// WORD: - reserved for special features
		default:
			Trap_Arg(args);
		}

		if (THROWN(DS_VALUE(ds))) {
			// Store THROWN value in TOS, so that Do_Next can handle it.
			*DS_TOP = *DS_VALUE(ds);
			return index;
		}

		// If word is typed, verify correct argument datatype:
		if (!TYPE_CHECK(args, VAL_TYPE(DS_VALUE(ds))))
			Trap3(RE_EXPECT_ARG, Func_Word(dsf), args, Of_Type(DS_VALUE(ds)));
	}

```

**File:** src/core/c-do.c (L803-979)
```c
*/	REBCNT Do_Next(REBSER *block, REBCNT index, REBFLG op)
/*
**		Evaluate the code block until we have:
**			1. An irreducible value (return next index)
**			2. Reached the end of the block (return END_FLAG)
**			3. Encountered an error
**
**		Index is a zero-based index into the block.
**		Op indicates infix operator is being evaluated (precedence);
**		The value (or error) is placed on top of the data stack.
**
***********************************************************************/
{
	REBVAL *value;
	REBVAL *word = 0;
	REBINT ftype;
	REBCNT dsf;

	//CHECK_MEMORY(1);
	CHECK_STACK(&value);
	if ((DSP + 20) > (REBINT)SERIES_REST(DS_Series)) Expand_Stack(STACK_MIN); //Trap0(RE_STACK_OVERFLOW);
	if (--Eval_Count <= 0 || Eval_Signals) Do_Signals();

	value = BLK_SKIP(block, index);
	//if (Trace_Flags) Trace_Eval(block, index);

reval:
	if (Trace_Flags) Trace_Line(block, index, value);

	//getchar();
	switch (EVAL_TYPE(value)) {

	case ET_WORD:
		value = Get_Var(word = value);
		if (IS_UNSET(value)) Trap1(RE_NO_VALUE, word);
		if (VAL_TYPE(value) >= REB_NATIVE && VAL_TYPE(value) <= REB_FUNCTION) goto reval; // || IS_LIT_PATH(value)
		DS_PUSH(value);
		if (IS_LIT_WORD(value)) VAL_SET(DS_TOP, REB_WORD);
		if (IS_FRAME(value)) Init_Obj_Value(DS_TOP, VAL_WORD_FRAME(word));
		index++;
		break;

	case ET_SELF:
		DS_PUSH(value);
		index++;
		break;

	case ET_SET_WORD:
		word = value;
		//if (!VAL_WORD_FRAME(word)) Trap1(RE_NOT_DEFINED, word); (checked in set_var)
		index = Do_Next(block, index+1, 0);
		// THROWN is handled in Set_Var.
		if (index == END_FLAG || VAL_TYPE(DS_TOP) <= REB_UNSET) Trap1(RE_NEED_VALUE, word);
		Set_Var(word, DS_TOP);
		//Set_Word(word, DS_TOP); // (value stays on stack)
		//Dump_Frame(Main_Frame);
		break;

	case ET_FUNCTION:
eval_func0:
		ftype = VAL_TYPE(value) - REB_NATIVE; // function type
		if (!word) word = ROOT_NONAME;
		dsf = Push_Func(FALSE, block, index, VAL_WORD_SYM(word), value);
eval_func:
		value = DSF_FUNC(dsf); // a safe copy of function
		if (VAL_TYPE(value) < REB_NATIVE) {
			Debug_Value(word, 4, 0);
			Dump_Values(value, 4);
		}
		index = Do_Args(value, 0, block, index+1); // uses old DSF, updates DSP
eval_func2:
		// Evaluate the function:
		DSF = dsf;	// Set new DSF
		if (!THROWN(DS_TOP)) {
			if (Trace_Flags) Trace_Func(word, value);
			Func_Dispatch[ftype](value);
		}
		else {
			*DS_RETURN = *DS_TOP;
		}

		// Reset the stack to prior function frame, but keep the
		// return value (function result) on the top of the stack.
		DSP = dsf;
		DSF = PRIOR_DSF(dsf);
		if (Trace_Flags) Trace_Return(word, DS_TOP);

		// The return value is a FUNC that needs to be re-evaluated.
		if (VAL_GET_OPT(DS_TOP, OPTS_REVAL) && ANY_FUNC(DS_TOP)) {
			value = DS_POP; // WARNING: value is volatile on TOS1 !
			word = Get_Type_Word(VAL_TYPE(value));
			index--;		// Backup block index to re-evaluate.
			if (IS_OP(value)) Trap_Type(value); // not allowed
			goto eval_func0;
		}
		break;

	case ET_OPERATOR:
		// An operator can be native or function, so its true evaluation
		// datatype is stored in the extended flags part of the value.
		if (!word) word = ROOT_NONAME;
		if (DSP <= 0 || index == 0) Trap1(RE_NO_OP_ARG, word);
		ftype = VAL_GET_EXT(value) - REB_NATIVE;
		dsf = Push_Func(TRUE, block, index, VAL_WORD_SYM(word), value); // TOS has first arg
		DS_PUSH(DS_VALUE(dsf)); // Copy prior to first argument
		goto eval_func;

	case ET_PATH:  // PATH, SET_PATH
		ftype = VAL_TYPE(value);
		word = value; // a path
		//index++; // now done below with +1

		//Debug_Fmt("t: %r", value);
		if (ftype == REB_SET_PATH) {
			index = Do_Next(block, index+1, 0);
			// THROWN is handled in Do_Path.
			if (index == END_FLAG || VAL_TYPE(DS_TOP) <= REB_UNSET) Trap1(RE_NEED_VALUE, word);
			Do_Path(&word, DS_TOP);
		} else {
			// Can be a path or get-path:
			value = Do_Path(&word, 0); // returns in word the path item, DS_TOP has value
			//Debug_Fmt("v: %r", value);
			// Value returned only for functions that need evaluation (but not GET_PATH):
			if (value && ANY_FUNC(value)) {
				if (IS_OP(value)) Trap_Type(value); // (because prior value is wiped out above)
				// Can be object/func or func/refinements or object/func/refinement:
				dsf = Push_Func(TRUE, block, index, VAL_WORD_SYM(word), value); // Do not unset TOS1 (it is the value)
				value = DS_TOP;
				index = Do_Args(value, word+1, block, index+1);
				ftype = VAL_TYPE(value)-REB_NATIVE;
				goto eval_func2;
			} else
				index++;
		}
		break;

	case ET_PAREN:
		DO_BLK(value);
		DSP++; // keep it on top
		index++;
		break;

	case ET_LIT_WORD:
		DS_PUSH(value);
		VAL_SET(DS_TOP, REB_WORD);
		index++;
		break;

	case ET_GET_WORD:
		DS_PUSH(Get_Var(value));
		index++;
		break;

	case ET_LIT_PATH:
		DS_PUSH(value);
		VAL_SET(DS_TOP, REB_PATH);
		index++;
		break;

	case ET_END:
		 return END_FLAG;

	default:
		//Debug_Fmt("Bad eval: %d %s", VAL_TYPE(value), Get_Type_Name(value));
		Crash(RP_BAD_EVALTYPE, VAL_TYPE(value));
		//return -index;
	}

	// If normal eval (not higher precedence of infix op), check for op:
	if (!op) {
		value = BLK_SKIP(block, index);
		if (IS_WORD(value) && VAL_WORD_FRAME(value) && IS_OP(Get_Var(value)))
			goto reval;
	}

	return index;
}
```

**File:** src/core/c-frame.c (L1-86)
```c
/***********************************************************************
**
**  REBOL [R3] Language Interpreter and Run-time Environment
**
**  Copyright 2012 REBOL Technologies
**  REBOL is a trademark of REBOL Technologies
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**  http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
**
************************************************************************
**
**  Module:  c-frame.c
**  Summary: frame management
**  Section: core
**  Author:  Carl Sassenrath
**  Notes:
**
***********************************************************************/
/*
		This structure is used for:

			1. Modules
			2. Objects
			3. Function frame (arguments)
			4. Closures

		A frame is a block that begins with a special FRAME! value
		(a datatype that links to the frame word list). That value
		(SELF) is followed by the values of the words for the frame.

		FRAME BLOCK:                            WORD LIST:
		+----------------------------+          +----------------------------+
		|    Frame Datatype Value    |--Series->|         SELF word          |
		+----------------------------+          +----------------------------+
		|          Value 1           |          |          Word 1            |
		+----------------------------+          +----------------------------+
		|          Value 2           |          |          Word 2            |
		+----------------------------+          +----------------------------+
		|          Value ...         |          |          Word ...          |
		+----------------------------+          +----------------------------+

		The word list holds word datatype values of the structure:

				Type:   word, 'word, :word, word:, /word
				Symbol: actual symbol
				Canon:  canonical symbol
				Typeset: index of the value's typeset, or zero

		This list is used for binding, evaluation, type checking, and
		can also be used for molding.

		When a frame is cloned, only the value block itself need be
		created. The word list remains the same. For functions, the
		value block can be pushed on the stack.

		Frame creation patterns:

			1. Function specification to frame. Spec is scanned for
			words and datatypes, from which the word list is created.
			Closures are identical.

			2. Object specification to frame. Spec is scanned for
			word definitions and merged with parent defintions. An
			option is to allow the words to be typed.

			3. Module words to frame. They are not normally known in
			advance, they are collected during the global binding of a
			newly loaded block. This requires either preallocation of
			the module frame, or some kind of special scan to track
			the new words.

			4. Special frames, such as system natives and actions
			may be created by specific block scans and appending to
			a given frame.
*/

```

**File:** src/boot/natives.r (L19-60)
```r
;-- Control Natives - nat_control.c

ajoin: native [
	{Reduces and joins a block of values into a new string.}
	block [block!]
]

also: native [
	{Returns the first value, but also evaluates the second.}
	value1 [any-type!]
	value2 [any-type!]
]

all: native [
	{Shortcut AND. Evaluates and returns at the first FALSE or NONE.}
	block [block!] {Block of expressions}
]

any: native [
	{Shortcut OR. Evaluates and returns the first value that is not FALSE or NONE.}
	block [block!] {Block of expressions}
]

apply: native [
	{Apply a function to a reduced block of arguments.}
	func [any-function!] "Function value to apply"
	block [block!] "Block of args, reduced first (unless /only)"
	/only "Use arg values as-is, do not reduce the block"
]

assert: native [
	"Assert that condition is true, else cause an assertion error."
	conditions [block!]
	/type "Safely check datatypes of variables (words and paths)"
]

attempt: native [
	"Tries to evaluate a block and returns result or NONE on error."
	block [block!]
]

break: native [
```

**File:** src/boot/natives.r (L541-600)
```r

in: native [
	{Returns the word or block in the object's context.}
	object [any-object! block!]
	word [any-word! block! paren!]  {(modified if series)}
]

parse: native [
	{Parses a string or block series according to grammar rules.}
	input [series!] {Input series to parse}
	rules [block! string! char! none!] {Rules to parse by (none = ",;")}
	/all {For simple rules (not blocks) parse all chars including whitespace}
	/case {Uses case-sensitive comparison}
]

set: native [
	{Sets a word, path, block of words, or object to specified value(s).}
	word [any-word! any-path! block! object!] {Word, block of words, path, or object to be set (modified)}
	value [any-type!] {Value or block of values}
	/any {Allows setting words to any value, including unset}
	/pad {For objects, if block is too short, remaining words are set to NONE}
]

to-hex: native [
	{Converts numeric value to a hex issue! datatype (with leading # and 0's).}
	value [integer! tuple!] {Value to be converted}
	/size {Specify number of hex digits in result}
	len [integer!]
]

type?: native [
	{Returns the datatype of a value.}
	value [any-type!]
	/word {Returns the datatype as a word}
]

unset: native [
	{Unsets the value of a word (in its current context.)}
	word [word! block!] {Word or block of words}
]

utf?: native [
	{Returns UTF BOM (byte order marker) encoding; + for BE, - for LE.}
	data [binary!]
]

invalid-utf?: native [
	{Checks UTF encoding; if correct, returns none else position of error.}
	data [binary!]
	/utf "Check encodings other than UTF-8"
	num [integer!] "Bit size - positive for BE negative for LE"
]

value?: native [
	{Returns TRUE if the word has a value.}
	value
]

;-- IO Natives - nat_io.c

```

**File:** src/boot/sysobj.r (L151-210)
```r
standard: context [

	error: context [ ; Template used for all errors:
		code: 0
		type: 'user
		id:   'message
		arg1:
		arg2:
		arg3:
		near:
		where:
			none
	]
 
	script: context [
		title:
		header:
		parent: 
		path:
		args:
			none
	]

	header: context [
		title: {Untitled}
		name:
		type:
		version:
		date:
		file:
		author:
		needs:
		options:
		checksum:
;		compress:
;		exports:
;		content:
			none
	]

	scheme: context [
		name:		; word of http, ftp, sound, etc.
		title:		; user-friendly title for the scheme
		spec:		; custom spec for scheme (if needed)
		info:		; prototype info object returned from query
;		kind:		; network, file, driver
;		type:		; bytes, integers, objects, values, block
		actor:		; standard action handler for scheme port functions
		awake:		; standard awake handler for this scheme's ports
			none
	]

	port: context [ ; Port specification object
		spec:		; published specification of the port
		scheme:		; scheme object used for this port
		actor:		; port action handler (script driven)
		awake:		; port awake function (event driven)
		state:		; internal state values (private)
		data:		; data buffer (usually binary or block)
		locals:		; user-defined storage of local data
```

**File:** src/core/u-parse.c (L1-33)
```c
/***********************************************************************
**
**  REBOL [R3] Language Interpreter and Run-time Environment
**
**  Copyright 2012 REBOL Technologies
**  REBOL is a trademark of REBOL Technologies
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**  http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
**
************************************************************************
**
**  Module:  u-parse.c
**  Summary: parse dialect interpreter
**  Section: utility
**  Author:  Carl Sassenrath
**  Notes:
**
***********************************************************************/

#include "sys-core.h"
#include "sys-state.h"

// Parser flags:
```

**File:** src/mezz/sys-load.r (L1-60)
```r
REBOL [
	System: "REBOL [R3] Language Interpreter and Run-time Environment"
	Title: "REBOL 3 Boot Sys: Load, Import, Modules"
	Rights: {
		Copyright 2012 REBOL Technologies
		REBOL is a trademark of REBOL Technologies
	}
	License: {
		Licensed under the Apache License, Version 2.0
		See: http://www.apache.org/licenses/LICENSE-2.0
	}
	Context: sys
	Note: {
		The boot binding of this module is SYS then LIB deep.
		Any non-local words not found in those contexts WILL BE
		UNBOUND and will error out at runtime!

		These functions are kept in a single file because they
		are inter-related.
	}
]

; BASICS:
; Code gets loaded in two ways:
;   1. As user code/data - residing in user context
;   2. As module code/data - residing in its own context
; Module loading can be delayed. This allows special modules like CGI, protocols,
; or HTML formatters to be available, but not require extra space.
; The system/modules list holds modules for fully init'd modules, otherwise it
; holds their headers, along with the binary or block that will be used to init them.

intern: function [
	"Imports (internalizes) words/values from the lib into the user context."
	data [block! any-word!] "Word or block of words to be added (deeply)"
][
	index: 1 + length? usr: system/contexts/user ; for optimization below (index for resolve)
	data: bind/new :data usr   ; Extend the user context with new words
	resolve/only usr lib index ; Copy only the new values into the user context
	:data
]

bind-lib: func [
	"Bind only the top words of the block to the lib context (used to load mezzanines)."
	block [block!]
][
	bind/only/set block lib ; Note: not bind/new !
	bind block lib
	block
]

export-words: func [
	"Exports the words of a context into both the system lib and user contexts."
	ctx [module! object!] "Module context"
	words [block! none!] "The exports words block of the module"
][
	if words [
		resolve/extend/only lib ctx words  ; words already set in lib are not overriden
		resolve/extend/only system/contexts/user lib words  ; lib, because of above
	]
]
```

**File:** src/mezz/boot-files.r (L15-23)
```r
;-- base: low-level boot in lib context:
[
	%base-constants.r
	%base-funcs.r
	%base-series.r
	%base-files.r
	%base-debug.r
	%base-defs.r
]
```

**File:** src/mezz/boot-files.r (L25-32)
```r
;-- sys: low-level sys context:
[
	%sys-base.r
	%sys-ports.r
	%sys-codec.r ; export to lib!
	%sys-load.r
	%sys-start.r
]
```

**File:** src/mezz/boot-files.r (L34-49)
```r
;-- lib: mid-level lib context:
[
	%mezz-types.r
	%mezz-func.r
	%mezz-debug.r
	%mezz-control.r
	%mezz-save.r
	%mezz-series.r
	%mezz-files.r
	%mezz-shell.r
	%mezz-math.r
	%mezz-help.r ; move dump-obj!
	%mezz-banner.r
	%mezz-colors.r
	%mezz-tail.r
]
```

**File:** src/mezz/boot-files.r (L51-54)
```r
;-- protocols:
[
	%prot-http.r
```

**File:** src/tools/systems.r (L21-60)
```r
systems: [
	[plat  os-name   os-base  build-flags]
	[0.1.03 "amiga"      posix  [HID NPS +SC CMT COP -SP -LM]]
	[0.2.04 "osx"        posix  [+OS NCM -LM]]			; no shared lib possible
	[0.2.05 "osxi"       posix  [ARC +O1 NPS PIC NCM HID STX -LM]]
	[0.3.01 "win32"      win32  [+O2 UNI W32 CON S4M EXE DIR -LM]]
	[0.4.02 "linux"      posix  [+O2 LDL ST1 -LM]]		; libc 2.3
	[0.4.03 "linux"      posix  [+O2 HID LDL ST1 -LM]]	; libc 2.5
	[0.4.04 "linux"      posix  [+O2 HID LDL ST1 M32 -LM]]	; libc 2.11
	[0.4.10 "linux_ppc"  posix  [+O1 HID LDL ST1 -LM]]
	[0.4.20 "linux_arm"  posix  [+O2 HID LDL ST1 -LM]]
	[0.4.21 "linux_arm"  posix  [+O2 HID LDL ST1 -LM PIE]]  ; bionic (Android)
	[0.4.30 "linux_mips" posix  [+O2 HID LDL ST1 -LM]]  ; glibc does not need C++
	[0.5.75 "haiku"      posix  [+O2 ST1 NWK]]
	[0.7.02 "freebsd"    posix  [+O1 C++ ST1 -LM]]
	[0.9.04 "openbsd"    posix  [+O1 C++ ST1 -LM]]
	[0.13.01 "android_arm"  android  [HID F64 LDL LLOG -LM CST]]
]

compile-flags: [
	+OS: "-Os"                    ; size optimize
	+O1: "-O1"                    ; full optimize
	+O2: "-O2"                    ; full optimize
	UNI: "-DUNICODE"              ; win32 wants it
	CST: "-DCUSTOM_STARTUP"		  ; include custom startup script at host boot
	HID: "-fvisibility=hidden"    ; all syms are hidden
	F64: "-D_FILE_OFFSET_BITS=64" ; allow larger files
	NPS: "-Wno-pointer-sign"      ; OSX fix
	NSP: "-fno-stack-protector"   ; avoid insert of functions names
	PIC: "-fPIC"                  ; position independent (used for libs)
	PIE: "-fPIE"                  ; position independent (executables)
	DYN: "-dynamic"               ; optimize for dll??
	NCM: "-fno-common"            ; lib cannot have common vars
	PAK: "-fpack-struct"          ; pack structures
	ARC: "-arch i386"             ; x86 32 bit architecture (OSX)
	M32: "-m32"                   ; use 32-bit memory model
]

linker-flags: [
	MAP: "-Wl,-M"  ; output a map
```

**File:** src/core/n-control.c (L1-61)
```c
/***********************************************************************
**
**  REBOL [R3] Language Interpreter and Run-time Environment
**
**  Copyright 2012 REBOL Technologies
**  REBOL is a trademark of REBOL Technologies
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**  http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
**
************************************************************************
**
**  Module:  n-control.c
**  Summary: native functions for control flow
**  Section: natives
**  Author:  Carl Sassenrath
**  Notes:
**    Warning: Do not cache pointer to stack ARGS (stack may expand).
**
***********************************************************************/

#include "sys-core.h"


// Local flags used for Protect functions below:
enum {
	PROT_SET,
	PROT_DEEP,
	PROT_HIDE,
	PROT_WORD,
};


/***********************************************************************
**
*/	void Protected(REBVAL *word)
/*
**		Throw an error if word is protected.
**
***********************************************************************/
{
	REBSER *frm;
	REBINT index = VAL_WORD_INDEX(word);

	if (index > 0) {
		frm = VAL_WORD_FRAME(word);
		if (VAL_PROTECTED(FRM_WORDS(frm)+index))
			Trap1(RE_LOCKED_WORD, word);
	}
	else if (index == 0) Trap0(RE_SELF_PROTECTED);
}

```
