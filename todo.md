
List of things to impelment:

- pliki `register_*` do rejestrowania funkcji natywnych wszystkie definiują swoje funkcje pomocnicze do rejestrowania
  warto to wyciągnąć do wspólnego kodu
- extend help system for user-defined functions

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