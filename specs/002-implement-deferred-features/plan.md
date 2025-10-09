# Implementation Plan: Deferred Language Capabilities (002)

**Branch**: `002-deferred-language-capabilities` | **Date**: 2025-10-08 | **Spec**: `/specs/002-implement-deferred-features/spec.md`
**Input**: Feature specification, clarification answers, research summary, data model, contracts, quickstart.

## Summary

Implement the deferred interpreter capabilities: decimal128 arithmetic, sandboxed ports with TLS controls, objects/path semantics, parse dialect, observability (trace/debug), and reflection helpers. Strategy follows research decisions—use `github.com/ericlagergren/decimal`, share HTTP client pools for TLS toggles, enforce sandbox root resolution, and emit rotating trace logs via lumberjack.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Go standard library, `github.com/ericlagergren/decimal`, `gopkg.in/natefinch/lumberjack.v2`  
**Storage**: Local filesystem (sandbox root, whitelist TOML, trace logs)  
**Testing**: `go test ./...` with contract suites (`test/contract`) and integration validation (`test/integration`)  
**Target Platform**: Cross-platform CLI (macOS/Linux)  
**Project Type**: Single Go CLI with layered `internal` packages  
**Performance Goals**: Decimal ops ≤1.5× integer baseline; trace overhead <5%; port I/O throughput within 10% of stdlib wrappers  
**Constraints**: Constitution mandates (type dispatch, stack/index safety, structured errors, observability, YAGNI), sandbox root enforcement, TLS secure-by-default, deterministic whitelist  
**Scale/Scope**: ~60 new/updated natives + supporting infrastructure; 10 new success criteria (SC-011–SC-020)

## Constitution Check

Gates satisfied after Phase 1 design:
- **Principle III**: Register new value types (decimal, object, port, path) in dispatch tables before evaluation.
- **Principle IV**: Objects rely on frame indices; ports encapsulate driver state without raw pointer leakage.
- **Principle V**: Map new failure modes to Math (400), Access (500), Script (300), Syntax (200) categories with near/where context.
- **Principle VI**: Trace/debug provide JSON event stream and inspector commands to maintain transparency.
- **Principle VII**: Scope limited to spec—no FTP/WebSocket ports or module system yet.

Re-run constitution checklist before implementation sign-off; current design introduces no violations.

## Project Structure

### Documentation (feature 002)

```
specs/002-implement-deferred-features/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── README.md
│   ├── math-decimal.md
│   ├── ports.md
│   ├── objects.md
│   ├── parse.md
│   ├── trace-debug.md
│   └── reflection.md
└── tasks.md  (Phase 2)
```

### Source Code Targets

```
cmd/
└── viro/
    ├── main.go        # CLI flags: sandbox root, trace file overrides, TLS toggles
    └── repl.go        # Debug prompt, trace commands

internal/
├── eval/             # Path evaluation, parse integration, trace hooks
├── frame/            # Object manifest support
├── native/
│   ├── control.go    # Debug command dispatch
│   ├── data.go       # Reflection natives
│   ├── function.go   # Function spec exposure
│   ├── io.go         # Ports (open/read/write/query)
│   ├── math.go       # Decimal promotion + rounding surface
│   ├── math_decimal.go  # NEW advanced math natives
│   ├── series.go     # Path-aware series operations
│   └── trace.go      # NEW trace/debug session state
├── parse/            # Parse dialect engine
├── repl/             # Trace/debug REPL UX
├── stack/            # Ensure path operations respect index-based safety
└── value/            # New Value types (decimal/object/port/path)

test/
├── contract/
│   ├── math_decimal_test.go
│   ├── ports_test.go
│   ├── objects_test.go
│   ├── parse_test.go
│   ├── trace_debug_test.go
│   └── reflection_test.go
└── integration/
    ├── sc011_validation_test.go
    ├── sc012_validation_test.go
    ├── sc013_validation_test.go
    ├── sc014_validation_test.go
    ├── sc015_validation_test.go
    ├── sc016_validation_test.go
    ├── sc017_validation_test.go
    ├── sc018_validation_test.go
    ├── sc019_validation_test.go
    └── sc020_validation_test.go
```

**Structure Decision**: Retain existing single-project Go layout. Add feature-specific files within `internal` packages and mirror coverage in contract/integration tests; no new Go modules/packages beyond those listed.

## Complexity Tracking

_No constitution exceptions or additional governance approvals required._
