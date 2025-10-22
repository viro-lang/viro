List of things to implement:

- extend help system for user-defined functions
   STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames

- set new fields in objects
   STATUS: NOT IMPLEMENTED - Objects have fixed manifest of fields, no support for dynamically adding new fields (put function errors if field doesn't exist)

- comprehensive cli interface
   STATUS: PARTIALLY IMPLEMENTED - References to CLI flags exist (sandbox-root, trace-file, trace-max-size, allow-insecure-tls) but no main.go CLI implementation found

- while should accept logic! or integer! (as it's documentation states)
   STATUS: NOT IMPLEMENTED - While only accepts blocks for both condition and body, not logic! or integer! values

- '=' should work for everything - not only numbers
   STATUS: NOT IMPLEMENTED - Equal function only handles integers, should use Value.Equals method for all types

- make debugging step by step actually work
   STATUS: PARTIALLY IMPLEMENTED - Debug commands exist and set stepping flags, but no actual interactive step-by-step execution implemented in evaluator

- wherever a viro value is converted into string it should use either `mold` or `form` functions