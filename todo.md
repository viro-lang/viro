List of things to implement:

- extend help system for user-defined functions
  STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames

- while should accept logic! or integer! (as it's documentation states)
  STATUS: NOT IMPLEMENTED - While only accepts blocks for both condition and body, not logic! or integer! values

- make debugging step by step actually work
  STATUS: PARTIALLY IMPLEMENTED - Debug commands exist and set stepping flags, but no actual interactive step-by-step execution implemented in evaluator

- wherever a viro value is converted into string it should use either `mold` or `form` functions
  STATUS: NOT IMPLEMENTED - Many places in codebase use .String() directly instead of mold/form functions

- implement all the series functions

- the 'read' native should support reading directories and return block with filenames

- fix the examples (some are incorrect viro code and others simply are missing native function implementations )
