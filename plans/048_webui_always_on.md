# Plan 048: WebUI Always-On Implementation

## Overview
Implement WebUI functionality as always-available without build tags, removing conditional compilation and noop implementations. This makes WebUI natives permanently registered in the standard evaluator.

## Inventory of Files/Build Tags
- **Noop files to delete**:
  - `internal/native/webui_noop.go`
  - `internal/native/register_webui_noop.go`
  - `internal/native/webui_runtime_noop.go`
  - `internal/native/webui_runtime_test.go` (if irrelevant)

- **Build tag removals**:
  - Remove `//go:build webui` from:
    - `internal/native/webui.go`
    - `internal/native/register_webui.go`
    - `internal/bootstrap/webui_registration.go`
    - `internal/value/webui.go`

- **Error category removal**:
  - Remove `verror.ErrIDWebUIUnavailable` and references

## Code Changes
1. Delete noop implementation files
2. Remove build tags from WebUI source files
3. Inline `RegisterWebUINatives` into bootstrap unconditionally
4. Remove conditional registration logic (`registerWebUINativesIfAvailable`)
5. Clean up error handling for unavailable WebUI

## Documentation Updates
- Update `docs/webui.md` to state always-built, mention CGO prerequisites
- Update `docs/repl-usage.md` to remove noop build instructions
- Update `README.md` and `RELEASE_NOTES.md` accordingly
- Update `specs/002-implement-deferred-features/contracts/webui.md`

## Build System Changes
- Remove `-tags webui` from `Makefile` and scripts
- Ensure CGO requirements are documented
- Update build instructions to reflect always-on WebUI

## Testing Strategy
1. **Contract tests first**: Add `test/contract/webui_availability_test.go`
   - Verify WebUI natives resolve to `native!` type
   - Ensure functions like `webui-new-window` are available
   - Test with standard evaluator (no build tags)
   - Avoid opening real windows (mock or skip UI operations)

2. **Integration tests**: Run affected packages and full test suite
3. **Cross-reference checks**: Use `rg` to ensure no remaining references to `webui-unavailable` or `-tags webui`

## Validation Steps
- Build succeeds with CGO enabled
- All WebUI natives are registered unconditionally
- No references to conditional WebUI availability remain
- Documentation accurately reflects new behavior

## Risk Assessment
- CGO dependency becomes mandatory for WebUI builds
- Potential build failures on systems without CGO/WebKit
- Need to ensure error handling for runtime WebUI failures (not availability)