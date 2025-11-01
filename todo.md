List of things to implement:

- extend help system for user-defined functions
  STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames

- comprehensive cli interface
  STATUS: NOT IMPLEMENTED - No main.go exists yet, so CLI flags (including trace-file and trace-max-size) are missing

- while should accept logic! or integer! (as it's documentation states)
  STATUS: NOT IMPLEMENTED - While only accepts blocks for both condition and body, not logic! or integer! values

- implement all the series functions

- the 'read' native should support reading directories and return block with filenames

- fix the examples (some are incorrect viro code and others simply are missing native function implementations )

- wherever a viro value is converted into string it should use either `mold` or `form` functions
  STATUS: NOT IMPLEMENTED - Many places in codebase use .String() directly instead of mold/form functions

- comprehensive cli interface
  STATUS: NOT IMPLEMENTED - No main.go exists yet, so CLI flags (including trace-file and trace-max-size) are missing

||||||| parent of fb24281 (Update)
- extend help system for user-defined functions
  STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames

=======
>>>>>>> fb24281 (Update)
- while should accept logic! or integer! (as it's documentation states)
  STATUS: NOT IMPLEMENTED - While only accepts blocks for both condition and body, not logic! or integer! values

- implement all the series functions

- the 'read' native should support reading directories and return block with filenames

- fix the examples (some are incorrect viro code and others simply are missing native function implementations )

- wherever a viro value is converted into string it should use either `mold` or `form` functions
  STATUS: NOT IMPLEMENTED - Many places in codebase use .String() directly instead of mold/form functions

- extend help system for user-defined functions
  STATUS: NOT IMPLEMENTED - Help system only shows native functions from root frame, doesn't include user-defined functions stored in frames
