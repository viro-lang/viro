List of things to implement:

- extend help system for user-defined functions
  STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames

- mold i form functions (print should use form)
  STATUS: NOT IMPLEMENTED - No mold or form functions exist. Print uses val.String() for output formatting

- set new fields in objects
  STATUS: NOT IMPLEMENTED - Objects have fixed manifest of fields, no support for dynamically adding new fields

- support for comments within the parser
  STATUS: NOT IMPLEMENTED - Parser has no comment support

- comprehensive cli interface
  STATUS: PARTIALLY IMPLEMENTED - Basic CLI exists with sandbox-root, trace-file, trace-max-size, allow-insecure-tls flags, but not comprehensive

- while should accept logic! or integer! (as it's documentation states)
  STATUS: NOT IMPLEMENTED - While only accepts blocks for both condition and body, not logic! or integer! values

- operations on series are implemented using switch, but this should be dynamic dispatch based on type (possible to extend from viro)
  STATUS: NOT IMPLEMENTED - Series operations still use switch statements on value.GetType(), not action-based dynamic dispatch

- all series functions should work on ALL series datatypes (blocks, strings, ports etc.)
  STATUS: NOT IMPLEMENTED - Series operations implemented with type switches, only work on specific hardcoded types

- '=' should work for everything - not only numbers
  STATUS: NOT IMPLEMENTED - Equal function only handles integers, uses type switches for other types but falls back to false

- make debugging step by step actually work
  STATUS: PARTIALLY IMPLEMENTED - Debug commands exist and set stepping flags, but no actual interactive step-by-step execution implemented in evaluator
