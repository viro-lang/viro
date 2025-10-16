List of things to impelment:

- extend help system for user-defined functions
- regression: introducing actions hid the documentation for them (docs are only for functions, not actions)
  it would be better to reuse the FunctionValue but as an Action type - it could normally store the documentation.
  as types have numbers going up from 0 then the type frame registry could be a simple array instead of a map. This would improve the performance.

- merge branch to main, upload to github
- register native functions within the root frame and drop special native fns handling
- mold i form functions (print should use form)
- set new fields in objects
- support for comments within the parser
- comprehensive cli interface
- while should accept logic! or integer! (as it's documentation states)
- operations on series are implemented using switch, but this should be dynamic dispatch based on type (possible to extend from viro)
- all series functions should work on ALL series datatypes (blocks, strings, ports etc.)
- '=' should work for everything - not only numbers
- make debugging step by step actually work
