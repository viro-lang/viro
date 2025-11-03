# Viro TODO List

## High Priority

- [ ] **CLI Interface Improvements**
  - STATUS: PARTIALLY COMPLETE - Basic CLI exists in cmd/viro/main.go with REPL, script execution, and multiple modes
  - REMAINING: Consider adding trace-file and trace-max-size flags if needed for file-based tracing
  - NOTE: Most CLI functionality is already implemented (see AGENTS.md for full CLI reference)

- [ ] **Fix Examples**
  - STATUS: NOT STARTED
  - Fix incorrect Viro code in examples/ directory
  - Implement missing native functions that examples rely on

## Medium Priority

- [ ] **Series Functions - Complete Implementation**
  - STATUS: PARTIALLY COMPLETE - Many series functions exist (first, last, next, skip, take, copy, etc.)
  - REMAINING: Implement pick, poke, select, clear, change, trim
  - See: specs/001-implement-the-core/contracts/ for specifications

- [ ] **While Loop - Accept logic! and integer!**
  - STATUS: NOT IMPLEMENTED
  - Currently only accepts blocks for condition and body
  - Should accept logic! or integer! values as documentation states

- [ ] **Read Native - Directory Support**
  - STATUS: NOT IMPLEMENTED
  - Currently reads files, should also support reading directories
  - Should return block with filenames when given a directory path

## Low Priority

- [ ] **Help System - User-Defined Functions**
  - STATUS: NOT IMPLEMENTED
  - Currently only shows native functions from root frame
  - Should include user-defined functions stored in frames

- [ ] **String Conversion - Use mold/form**
  - STATUS: NOT IMPLEMENTED
  - Many places use .String() directly instead of mold/form functions
  - Need to audit codebase and replace with proper Viro string conversion
