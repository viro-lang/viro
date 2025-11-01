# Plan 019: Remove REBOL Mentions from Codebase

## Context

The codebase contains numerous references to "REBOL" throughout documentation, source code comments, and specifications. These mentions pollute LLM context and can cause confusion, leading LLMs to generate REBOL code instead of Viro code. While Viro was inspired by REBOL, it is a distinct language with its own semantics and design decisions.

## Goal

Remove all REBOL mentions from the codebase while preserving:
1. Technical accuracy of documentation
2. Design rationale explanations
3. Historical context where genuinely needed
4. External references (URLs to REBOL documentation can remain)

## Scope Analysis

Based on `rg` analysis, REBOL is mentioned in **42 files** with approximately **150+ occurrences**:

### High-Impact Files (Documentation)
- `README.md` (8 mentions) - Main project description
- `RELEASE_NOTES.md` (9 mentions) - Release documentation
- `AGENTS.md` (1 mention) - Agent guidelines
- `CLAUDE.md` (4 mentions) - Claude-specific instructions
- `todo.md` (1 mention) - Already notes removal needed

### Documentation Files
- `docs/repl-usage.md` (5 mentions)
- `docs/scoping-differences.md` (28 mentions) - Heavy REBOL comparison
- `docs/operator-precedence.md` (8 mentions)

### Specification Files
- `specs/001-implement-the-core/` - Multiple files with REBOL references
- `specs/002-implement-deferred-features/` - Data model and contracts
- `specs/004-dynamic-function-invocation/` - Research and spec files

### Source Code Files
- `cmd/viro/help.go` - Help text
- `internal/value/string.go` (2 mentions) - Comments
- `internal/value/word.go` (1 mention) - Comments
- `internal/value/types.go` (1 mention) - Comments
- `internal/frame/frame.go` (1 mention) - Comment
- `internal/native/*.go` - Various comments

### Example Files
- `examples/05_data_manipulation.viro` (1 mention)
- `examples/README.md` (2 mentions)

### Plan Files
- `plans/003_cli.md` (2 mentions)

## Replacement Strategy

### 1. Direct Descriptions
Replace "REBOL-inspired" and "REBOL-like" with descriptive phrases:
- "REBOL-inspired programming language" → "homoiconic programming language"
- "REBOL-style evaluation" → "left-to-right evaluation"
- "REBOL semantics" → "Viro semantics" or specific behavior description
- "REBOL-readable format" → "code-readable format" or "serialized format"

### 2. Comparison Documents
For files that compare Viro with REBOL (e.g., `scoping-differences.md`):
- **Option A**: Rewrite to describe Viro's scoping without comparison
- **Option B**: Rename to describe the feature, not the comparison
- **Recommended**: Rewrite to focus on Viro's behavior with brief "unlike some languages" mentions

### 3. Code Comments
Replace technical comments:
- "REBOL series semantics" → "character-based series semantics"
- "per REBOL semantics" → "per Viro semantics"
- "REBOL character series" → "character series"

### 4. External References
Keep URLs to REBOL documentation in research files but mark them as "historical reference" or "inspiration source"

### 5. Type Names and Native Names
Keep references to naming conventions:
- "Use REBOL-style native names" → "Use Viro-style native names (lowercase, with ?, !)"

## Implementation Plan

### Phase 1: Documentation (High Priority)
1. **README.md**
   - Remove "REBOL-Inspired Interpreter" from title
   - Change description to focus on Viro's features
   - Remove comparison bullet points
   - Keep feature descriptions standalone

2. **RELEASE_NOTES.md**
   - Remove "Differences from REBOL" section
   - Reframe features as Viro features, not comparisons

3. **AGENTS.md**
   - Change "REBOL-style native names" to "Viro-style native names"

4. **CLAUDE.md**
   - Remove "It is NOT a REBOL interpreter" clarification
   - Describe Viro standalone

5. **cmd/viro/help.go**
   - Update help text description

### Phase 2: User-Facing Documentation
6. **docs/repl-usage.md**
   - Remove "Known Differences from REBOL" section
   - Focus on Viro's actual behavior

7. **docs/scoping-differences.md**
   - **MAJOR REWRITE NEEDED**
   - Rename to `docs/scoping-model.md` or similar
   - Describe Viro's local-by-default scoping
   - Briefly mention "unlike global-by-default languages"
   - Remove all REBOL code examples
   - Keep Viro examples and expand them

8. **docs/operator-precedence.md**
   - Remove REBOL comparisons
   - Focus on Viro's left-to-right evaluation model
   - Keep examples but frame them as Viro-specific

### Phase 3: Specifications
9. **specs/001-implement-the-core/**
   - `spec.md` - Remove "REBOL semantics" mentions
   - `plan.md` - Rewrite design decision rationales
   - `quickstart.md` - Remove REBOL references
   - `contracts/*.md` - Replace "REBOL" with descriptive terms
   - Keep external URLs but mark as historical

10. **specs/002-implement-deferred-features/**
    - Similar treatment as 001 specs

11. **specs/004-dynamic-function-invocation/**
    - Similar treatment as 001 specs

### Phase 4: Source Code
12. **internal/value/string.go**
    - "REBOL series semantics" → "character series semantics"
    - "REBOL strings are character series" → "strings are character series"

13. **internal/value/word.go**
    - "follow REBOL semantics" → "follow Viro semantics"

14. **internal/value/types.go**
    - "align with REBOL's type system" → "Viro type system"

15. **internal/frame/frame.go**
    - "differs from REBOL's global-by-default" → "local-by-default scoping"

16. **internal/native/*.go**
    - "REBOL-readable" → "code-readable" or "serialized"
    - Update function documentation

17. **test/contract/*.go**
    - Update test comments
    - "REBOL-style" → "left-to-right" or specific behavior

### Phase 5: Examples and Misc
18. **examples/05_data_manipulation.viro**
    - "REBOL-readable" → "code format" or "serialized"

19. **examples/README.md**
    - Similar replacements

20. **plans/003_cli.md**
    - Historical document, low priority
    - Update if clarity needed

21. **todo.md**
    - Remove completed task

## Testing Strategy

After each phase:
1. Run `make test` to ensure no behavioral changes
2. Run `make build` to verify compilation
3. Verify documentation renders correctly
4. Spot-check that descriptions remain accurate

## Verification

After completion, verify removal:
```bash
rg -i "rebol" --type go --type md --type viro
```

Should return:
- Zero results in source code comments
- Zero results in user-facing documentation
- Possibly remaining in:
  - External URLs (acceptable)
  - Historical plan documents (acceptable if marked)

## Success Criteria

1. ✅ No REBOL mentions in source code comments
2. ✅ No REBOL mentions in user-facing documentation (README, docs/)
3. ✅ No REBOL mentions in specifications
4. ✅ All tests pass after changes
5. ✅ Documentation remains technically accurate
6. ✅ Design rationales are preserved with descriptive language
7. ✅ LLMs reading the codebase won't be exposed to REBOL terminology

## Risks and Mitigation

**Risk**: Loss of historical context for design decisions
**Mitigation**: Keep detailed descriptions of *why* features work the way they do, just without naming REBOL

**Risk**: Breaking existing documentation clarity
**Mitigation**: Make descriptions more detailed and self-explanatory

**Risk**: Confusion about naming conventions (?, ! suffixes)
**Mitigation**: Document as "Viro conventions" with examples

## Estimated Effort

- Phase 1 (Documentation): 30 minutes
- Phase 2 (User Docs): 45 minutes (scoping-differences.md rewrite is major)
- Phase 3 (Specifications): 60 minutes (many files, careful editing needed)
- Phase 4 (Source Code): 20 minutes (mostly comment updates)
- Phase 5 (Examples): 10 minutes
- Testing & Verification: 15 minutes

**Total**: ~3 hours

## Notes

- The hardest file is `docs/scoping-differences.md` (28 mentions) which needs complete rewrite
- Many specifications have "REBOL semantics" which should become "Viro semantics"
- Keep the *intent* of comparisons but frame as "Viro's approach" vs "other languages"
- External documentation URLs can remain as "reference" or "inspiration" links
