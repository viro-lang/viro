# viro Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-07

## Active Technologies
- Go 1.21+ (requires generics and improved error handling) + Go standard library + github.com/chzyer/readline (REPL command history and multi-line input) (001-implement-the-core)
- Go 1.21+ + Go standard library, `github.com/ericlagergren/decimal`, `gopkg.in/natefinch/lumberjack.v2` (002-implement-deferred-features)
- Local filesystem (sandbox root, whitelist TOML, trace logs) (002-implement-deferred-features)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Go 1.21+ (requires generics and improved error handling)

## Code Style
Go 1.21+ (requires generics and improved error handling): Follow standard conventions

## Testing Guidelines
- **ALWAYS prefer writing automated tests** over running the viro interpreter directly for validation
- Use Go's testing framework to create test cases that exercise viro functionality
- Only run the viro interpreter manually when absolutely necessary (e.g., for exploratory testing or user-facing demonstrations)
- Automated tests provide better coverage, reproducibility, and serve as documentation
- **When running viro binary for testing**: Always ensure test scripts end with 'quit' command to avoid entering the REPL loop
- Tests must never perform real network calls (no outbound HTTP, DNS, etc.); mock or stub all external interactions.
- When network-like interactions are needed, spin up a local in-process or 127.0.0.1 test server (stub/mocked handlers) instead of reaching external services.

## Recent Changes
- 002-implement-deferred-features: Added Go 1.21+ + Go standard library, `github.com/ericlagergren/decimal`, `gopkg.in/natefinch/lumberjack.v2`
- 001-implement-the-core: Added Go 1.21+ (requires generics and improved error handling) + Go standard library + github.com/chzyer/readline (REPL command history and multi-line input)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
