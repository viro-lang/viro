# Viro TODO List

## High Priority

## Medium Priority

## Low Priority

- [ ] **Help System - User-Defined Functions**
  - STATUS: PARTIALLY COMPLETE - Infrastructure exists but no user interface
  - Help system dynamically builds registry from root frame at runtime
  - Includes both native and user-defined functions (see internal/native/help.go:10-22)
  - ❌ MISSING: Users cannot attach documentation to their functions
  - FuncDoc infrastructure exists (internal/docmodel/doc.go:12-21) but fn native passes nil for doc parameter (internal/native/function.go:66)
  - NEEDED: Add syntax/refinement to fn for users to provide documentation. Preferably the `--doc` refinement accepting a block like `[summary: "Does a thing" ...]`

- [ ] **String Conversion - Use mold/form**
  - STATUS: ACTIVE ISSUE
  - Found 85 uses of .String() across 20 files in internal/ directory
  - Need to audit codebase and replace with proper Viro string conversion (mold/form)
  - See: internal/native/format.go for mold/form implementations
