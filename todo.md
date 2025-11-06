# Viro TODO List

## High Priority

- [x] **CLI Interface Improvements**
  - STATUS: ✅ COMPLETE
  - Full CLI implementation exists in cmd/viro/main.go and run.go
  - All modes implemented: REPL, script execution, eval, check, version, help
  - Configuration system with flags and environment variables working
  - See: cmd/viro/run.go:38-60 for mode execution

- [x] **Fix Examples**
  - STATUS: ✅ COMPLETE
  - All 9 example files run successfully without errors
  - Examples cover: basics, control flow, functions, series, data manipulation, objects, advanced patterns, practical algorithms, and script arguments
  - Located: examples/*.viro

## Medium Priority

- [x] **Series Functions - Complete Implementation**
  - STATUS: ✅ COMPLETE
  - All series functions fully implemented and registered
  - Implemented: pick, poke, select, clear, change, trim (all with full documentation)
  - See: internal/native/register_series.go:95-124 for complete implementation

- [x] **While Loop - Accept logic! and integer!**
  - STATUS: ✅ COMPLETE
  - While loop now accepts blocks, logic!, and integer! values for condition
  - Blocks are re-evaluated on each iteration, other values used directly
  - See: internal/native/control.go:159-222 for implementation

- [x] **Read Native - Directory Support**
  - STATUS: ✅ COMPLETE
  - Implemented directory reading in internal/native/io.go:556-577
  - Returns block with filenames when given a directory path
  - Respects sandbox rules for security
  - Tests added in test/contract/ports_test.go:500-610

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
