# Implementation Plan: Consolidate Duplicate Path Eval Parsing Tests

**Plan ID**: 029  
**GitHub Issue**: #52  
**Type**: Test Refactoring  
**Complexity**: Low  
**Estimated Effort**: 30 minutes  
**Risk Level**: Very Low (test-only change, no production code modification)

---

## Feature Summary

Consolidate duplicate path evaluation segment parsing tests by removing the redundant `eval_path_test.go` file and relying on the comprehensive test coverage already present in `path_test.go`. This refactoring was identified by viro-reviewer during code review.

### Current State

- **`internal/parse/path_test.go`**: Contains comprehensive path parsing tests including:
  - Lines 93-122: Basic eval segment test (`foo.(field).bar`)
  - Lines 124-167: Nested eval segment test (`foo.(bar.(baz)).qux`)
  - Lines 293-318: Error cases for leading eval segments
  - Multiple integration tests covering path, set-path, and get-path variants

- **`internal/parse/eval_path_test.go`**: Contains only 3 basic test cases:
  - `data.(idx)` - simple eval (path)
  - `data.(idx):` - set-path eval
  - `:data.(idx)` - get-path eval

### Problem

The tests in `eval_path_test.go` provide redundant coverage that is already comprehensively tested in `path_test.go`. The `path_test.go` file includes:
- More complex eval segments (with nested paths)
- Better structured table-driven tests
- Detailed segment-level validation
- Error case coverage
- Integration with other syntax elements

---

## Research Findings

### Coverage Analysis

1. **`eval_path_test.go` Test Cases**:
   - `data.(idx)` - Basic path with eval segment
   - `data.(idx):` - Set-path with eval segment  
   - `:data.(idx)` - Get-path with eval segment

2. **`path_test.go` Equivalent/Superior Coverage**:
   - Line 93-122: `foo.(field).bar` - Tests basic eval segments with word lookups
   - Line 124-167: `foo.(bar.(baz)).qux` - Tests nested eval segments (more complex)
   - Line 239-258: Set-path tests including `obj.name: "Alice"`
   - The comprehensive tests validate segment types, values, and nesting

3. **Validation Differences**:
   - `eval_path_test.go`: Only validates parse success and value type
   - `path_test.go`: Validates segment types, values, nesting structure, and error conditions

### Dependencies Check

```bash
# No code references to eval_path_test.go found
rg "eval_path_test" --type go
# Result: No matches
```

The file is standalone with no external dependencies.

### Test Execution Verification

```bash
# All tests pass independently
go test -v ./internal/parse -run "TestParsePathWithEvalSegment"  # 3/3 pass
go test -v ./internal/parse -run "TestPathTokenization"          # 6/6 pass
```

---

## Architecture Overview

This is a **pure test consolidation refactoring** with:
- No production code changes
- No API changes
- No behavioral changes
- Maintains 100% test coverage with better-structured tests

The architecture remains unchanged. The refactoring simply removes redundant test coverage.

---

## Implementation Roadmap

### Phase 1: Pre-Deletion Verification (5 minutes)

**Objective**: Confirm that removing `eval_path_test.go` will not reduce test coverage.

**Steps**:

1. **Run baseline test coverage**:
   ```bash
   go test -coverprofile=/tmp/before_coverage.out ./internal/parse
   go tool cover -func=/tmp/before_coverage.out > /tmp/before_coverage.txt
   ```
   - Record total coverage percentage
   - Identify any lines only covered by `eval_path_test.go`

2. **Verify test overlap**:
   - Compare test cases in `eval_path_test.go` (lines 14-16) with `path_test.go`
   - Confirm that:
     - Basic eval segments: Covered by line 93-122 (`foo.(field).bar`)
     - Set-path eval: Covered by line 239-258 (set-path tests)
     - Get-path eval: Covered by integration tests in `path_test.go`
   - Document any gaps (expected: none)

3. **Check for test-specific logic**:
   ```bash
   rg "TestParsePathWithEvalSegment" --type go
   ```
   - Verify no other code depends on this test function
   - Expected: Only the test file itself

4. **Run all parse tests to establish baseline**:
   ```bash
   go test -json ./internal/parse | jq -r 'select(.Action == "pass" or .Action == "fail") | "\(.Action): \(.Test // "package")"' > /tmp/baseline_tests.txt
   ```
   - Count total passing tests
   - Verify all tests pass before deletion

**Success Criteria**:
- Coverage report shows eval segment parsing is tested in `path_test.go`
- No unique test scenarios found in `eval_path_test.go`
- All 100% of current tests pass
- Coverage percentage ≥ 74.9% (current baseline)

**Decision Point**: If any unique coverage is found, add those specific test cases to `path_test.go` before deletion.

---

### Phase 2: File Deletion (2 minutes)

**Objective**: Remove the redundant test file.

**Steps**:

1. **Delete the file**:
   ```bash
   rm internal/parse/eval_path_test.go
   ```

2. **Verify deletion**:
   ```bash
   git status
   ls -la internal/parse/*test.go
   ```
   - Confirm `eval_path_test.go` is marked as deleted
   - Confirm other test files remain

**Success Criteria**:
- File is deleted and staged for commit
- Other test files are unchanged

---

### Phase 3: Post-Deletion Validation (10 minutes)

**Objective**: Verify that test coverage and functionality remain intact.

**Steps**:

1. **Run all parse tests**:
   ```bash
   go test -v ./internal/parse
   ```
   - Verify all remaining tests pass
   - Note: Test count should decrease by exactly 3 (the test cases from `eval_path_test.go`)

2. **Generate new coverage report**:
   ```bash
   go test -coverprofile=/tmp/after_coverage.out ./internal/parse
   go tool cover -func=/tmp/after_coverage.out > /tmp/after_coverage.txt
   ```

3. **Compare coverage**:
   ```bash
   diff /tmp/before_coverage.txt /tmp/after_coverage.txt
   ```
   - Verify coverage percentage is identical or improved
   - Verify no uncovered lines were introduced
   - Expected: No material changes (same coverage from better tests)

4. **Run structured test output**:
   ```bash
   go test -json ./internal/parse | jq -r 'select(.Action == "pass" or .Action == "fail") | "\(.Action): \(.Test // "package")"' > /tmp/after_tests.txt
   wc -l /tmp/baseline_tests.txt /tmp/after_tests.txt
   ```
   - Verify test count decreased by exactly 4 lines (3 subtests + parent test)

5. **Run full test suite**:
   ```bash
   go test ./...
   ```
   - Verify no other packages were affected
   - Confirm all tests still pass

6. **Verify integration tests**:
   ```bash
   go test -v ./test/integration -run ".*Path.*"
   ```
   - Ensure integration-level path tests still pass
   - These provide additional coverage verification

**Success Criteria**:
- All remaining tests pass (100% pass rate)
- Coverage percentage unchanged or improved (≥ 74.9%)
- Test count decreased by exactly 4 (3 subtests + 1 parent)
- No new uncovered lines introduced
- Integration tests remain passing

---

### Phase 4: Documentation and Commit (3 minutes)

**Objective**: Document the change and create a clean commit.

**Steps**:

1. **Verify change scope**:
   ```bash
   git status
   git diff --stat
   ```
   - Should show only 1 file deleted: `internal/parse/eval_path_test.go`

2. **Create commit** (following Viro guidelines):
   ```bash
   git add internal/parse/eval_path_test.go
   git commit -m "test: consolidate duplicate path eval parsing tests

Removed eval_path_test.go as all its test cases are already covered
by the more comprehensive tests in path_test.go. The removed tests
(data.(idx), data.(idx):, :data.(idx)) are redundant with existing
coverage for eval segments in paths, set-paths, and get-paths.

Closes #52"
   ```

3. **Verify commit**:
   ```bash
   git show --stat
   ```
   - Review commit message follows Viro conventions
   - Verify only the test file was deleted

**Success Criteria**:
- Clean commit with only the test file deletion
- Commit message follows Viro guidelines (lowercase prefix, explains "why")
- References GitHub issue #52

---

## Integration Points

### Affected Components

- **Parser Tests** (`internal/parse/`): Test file count reduced by 1
- **CI/CD Pipeline**: Test count will decrease (expected behavior)
- **Coverage Reports**: Should remain stable or improve

### No Impact On

- Production code (`internal/parse/parser.go`, `tokenizer.go`, etc.)
- Other test files
- Runtime behavior
- Public APIs
- Documentation (tests are self-documenting)

---

## Testing Strategy

### Pre-Deletion Testing

1. **Baseline Coverage**:
   ```bash
   go test -coverprofile=/tmp/baseline.out ./internal/parse
   ```
   - Establish coverage baseline (expected: 74.9%)

2. **Baseline Test Count**:
   ```bash
   go test -json ./internal/parse | jq -s 'map(select(.Test and .Action == "pass")) | length'
   ```
   - Count passing tests (expected: ~40-50 tests)

### Post-Deletion Testing

1. **Coverage Verification**:
   - Run same coverage command
   - Compare line-by-line coverage
   - Verify no regression

2. **Test Count Verification**:
   - Run same test count command
   - Verify decrease of exactly 4 tests (3 subtests + parent)

3. **Full Suite Validation**:
   ```bash
   go test ./...
   make test
   ```
   - Ensure no ripple effects

### Acceptance Criteria

- ✅ All tests pass after deletion
- ✅ Coverage ≥ 74.9% (no decrease)
- ✅ Test count decreased by exactly 4
- ✅ No new uncovered lines
- ✅ Integration tests pass
- ✅ Clean commit with proper message

---

## Potential Challenges

### Challenge 1: Hidden Test Dependencies

**Risk**: Low  
**Likelihood**: Very Low

**Description**: Another test might indirectly depend on `TestParsePathWithEvalSegment`.

**Mitigation**:
- Pre-deletion grep for test name references
- Run full test suite after deletion
- Check for import cycles or test helpers

**Detection**:
```bash
rg "TestParsePathWithEvalSegment" --type go
```

**Resolution**: If found, evaluate dependency and either:
- Remove dependency (if test coupling)
- Extract shared helper function (if legitimate shared logic)

---

### Challenge 2: Coverage Decrease

**Risk**: Very Low  
**Likelihood**: Near Zero

**Description**: Removing tests might reveal gaps in `path_test.go` coverage.

**Mitigation**:
- Detailed pre-deletion coverage comparison
- Line-by-line diff of coverage reports
- Manual review of test case mapping

**Detection**:
```bash
diff /tmp/before_coverage.txt /tmp/after_coverage.txt
```

**Resolution**: If coverage decreases:
1. Identify uncovered lines
2. Add specific test cases to `path_test.go`
3. Re-run deletion process

---

### Challenge 3: Test Count Mismatch

**Risk**: Very Low  
**Likelihood**: Low

**Description**: Test count doesn't decrease by exactly 4.

**Mitigation**:
- Use structured JSON test output
- Count exact test hierarchy before/after
- Verify no test name conflicts

**Detection**:
```bash
diff /tmp/baseline_tests.txt /tmp/after_tests.txt
```

**Resolution**: If mismatch:
1. Review diff to identify discrepancy
2. Check for parallel test execution issues
3. Re-run with `-count=1` to disable caching

---

### Challenge 4: Git Merge Conflicts

**Risk**: Very Low  
**Likelihood**: Low (if active development on parse package)

**Description**: Concurrent changes to `eval_path_test.go` cause conflicts.

**Mitigation**:
- Check for open PRs touching this file
- Complete refactoring quickly (30 min window)
- Coordinate with team if needed

**Detection**:
```bash
git pull --rebase
```

**Resolution**: If conflicts:
1. Review conflicting changes
2. Incorporate into `path_test.go` if new test cases
3. Complete deletion
4. Notify PR author of consolidation

---

## Viro Guidelines Reference

### Relevant Guidelines

1. **TDD Mandatory** (AGENTS.md):
   - ✅ Tests exist first (we're consolidating, not removing coverage)
   - ✅ All tests pass before and after changes

2. **Code Style - NO COMMENTS** (AGENTS.md):
   - ✅ Test names are self-documenting
   - ✅ Table-driven tests with clear names

3. **Table-Driven Tests** (AGENTS.md):
   - ✅ `path_test.go` uses proper table-driven structure
   - ✅ `eval_path_test.go` also uses table-driven tests (but redundant)

4. **Workflow - Automated Tests Preferred** (AGENTS.md):
   - ✅ Using automated test suite verification
   - ✅ Coverage reports for validation
   - ✅ No manual execution required

5. **Planning - Sequential Numbering** (AGENTS.md):
   - ✅ This plan is `029_consolidate_path_eval_tests.md`
   - ✅ Follows sequential pattern after plan 028

### Adherence Verification

- **Test Coverage**: Maintained via comprehensive `path_test.go` tests
- **No Production Changes**: Only test file deletion
- **Automated Validation**: Full test suite run before/after
- **Documentation**: This plan documents the "why" (redundant tests)
- **Clean Commits**: Single-purpose commit with clear message

---

## Decision Framework for Implementation

### When to Proceed with Deletion

**Proceed if**:
- ✅ Coverage diff shows no unique lines in `eval_path_test.go`
- ✅ All test cases map to equivalent or better tests in `path_test.go`
- ✅ No external references to the test function
- ✅ Full test suite passes

**Do NOT proceed if**:
- ❌ Any test case in `eval_path_test.go` lacks equivalent coverage
- ❌ Coverage would decrease
- ❌ Other code references `TestParsePathWithEvalSegment`
- ❌ Active development conflicts detected

### Handling Edge Cases

**Case 1: Slightly Different Test Approach**

If `eval_path_test.go` uses a different validation approach that might catch different bugs:

**Decision**: 
- Merge the validation approach into `path_test.go`
- Add explicit test cases if needed
- Then proceed with deletion

**Case 2: Historical Test Documentation**

If the test file has valuable comments explaining design decisions:

**Decision**:
- Extract comments to `path_test.go` or documentation
- Update this plan with historical context
- Proceed with deletion

**Case 3: Performance Implications**

If fewer tests significantly impact CI time:

**Decision**:
- Measure actual time difference (expected: negligible, ~1ms)
- Document improvement in commit message
- Proceed with deletion

---

## Validation Checkpoints

### Checkpoint 1: Pre-Deletion (After Phase 1)

**Questions to Answer**:
- ✅ Is coverage ≥ 74.9%?
- ✅ Are all 3 test cases covered by `path_test.go`?
- ✅ Do all tests pass?
- ✅ Are there any external references?

**Go/No-Go Decision**: Proceed to Phase 2 if all ✅

---

### Checkpoint 2: Post-Deletion (After Phase 3)

**Questions to Answer**:
- ✅ Do all remaining tests pass?
- ✅ Is coverage still ≥ 74.9%?
- ✅ Did test count decrease by exactly 4?
- ✅ Do integration tests still pass?

**Go/No-Go Decision**: Proceed to Phase 4 if all ✅

---

### Checkpoint 3: Pre-Commit (During Phase 4)

**Questions to Answer**:
- ✅ Is the commit message clear and follows Viro style?
- ✅ Does `git diff` show only the deleted file?
- ✅ Does the commit close issue #52?
- ✅ Have all validation steps passed?

**Go/No-Go Decision**: Commit if all ✅

---

## Success Metrics

### Quantitative Metrics

| Metric | Before | After | Status |
|--------|--------|-------|--------|
| Test Files | 5 | 4 | ✅ Reduced |
| Test Count | ~45 | ~41 | ✅ Decreased by 4 |
| Coverage % | 74.9% | ≥74.9% | ✅ Maintained |
| Passing Tests | 100% | 100% | ✅ No regression |
| LOC (test) | ~450 | ~400 | ✅ Reduced duplication |

### Qualitative Metrics

- ✅ **Reduced Maintenance**: One fewer test file to maintain
- ✅ **Better Test Organization**: All path tests in one logical file
- ✅ **Clearer Intent**: Comprehensive tests are more descriptive
- ✅ **Faster Test Discovery**: Developers find all path tests in one place

---

## Rollback Plan

### If Issues Are Discovered

**Immediate Rollback** (within same session):
```bash
git reset --hard HEAD~1
```

**Rollback After Push** (if merged to main):
```bash
git revert <commit-hash>
git push origin main
```

### When to Rollback

- Coverage decreases unexpectedly
- Tests start failing in CI/CD
- Integration tests break
- Hidden dependencies discovered

### Rollback Validation

After rollback:
```bash
go test ./internal/parse
go test ./...
```

Verify all tests pass and coverage is restored.

---

## Timeline Estimate

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 1: Pre-Deletion Verification | 5 min | 5 min |
| Phase 2: File Deletion | 2 min | 7 min |
| Phase 3: Post-Deletion Validation | 10 min | 17 min |
| Phase 4: Documentation and Commit | 3 min | 20 min |
| **Buffer for Issues** | 10 min | **30 min** |

---

## Next Steps for Coder Agent

### Immediate Actions

1. **Start with Phase 1**: Run baseline coverage and test counts
2. **Analyze Coverage**: Compare `eval_path_test.go` vs `path_test.go` coverage
3. **Make Go/No-Go Decision**: Based on Checkpoint 1 criteria
4. **Execute Deletion**: If validated, proceed with Phase 2-4
5. **Verify Results**: Ensure all checkpoints pass
6. **Commit Changes**: With proper message referencing #52

### Command Sequence

```bash
# Phase 1: Verification
go test -coverprofile=/tmp/before_coverage.out ./internal/parse
go tool cover -func=/tmp/before_coverage.out > /tmp/before_coverage.txt
go test -json ./internal/parse | jq -r 'select(.Action == "pass" or .Action == "fail") | "\(.Action): \(.Test // "package")"' > /tmp/baseline_tests.txt
rg "TestParsePathWithEvalSegment" --type go

# Phase 2: Deletion
rm internal/parse/eval_path_test.go
git status

# Phase 3: Validation
go test -v ./internal/parse
go test -coverprofile=/tmp/after_coverage.out ./internal/parse
go tool cover -func=/tmp/after_coverage.out > /tmp/after_coverage.txt
diff /tmp/before_coverage.txt /tmp/after_coverage.txt
go test ./...

# Phase 4: Commit
git add internal/parse/eval_path_test.go
git commit -m "test: consolidate duplicate path eval parsing tests

Removed eval_path_test.go as all its test cases are already covered
by the more comprehensive tests in path_test.go. The removed tests
(data.(idx), data.(idx):, :data.(idx)) are redundant with existing
coverage for eval segments in paths, set-paths, and get-paths.

Closes #52"
```

### Expected Output

- ✅ All tests pass
- ✅ Coverage remains ≥ 74.9%
- ✅ Test count decreases by 4
- ✅ Clean commit with one deleted file
- ✅ Issue #52 closed

---

## Conclusion

This is a **low-risk, high-value refactoring** that:

- Eliminates test duplication
- Improves test organization
- Maintains 100% coverage
- Follows Viro development guidelines
- Requires no production code changes

The comprehensive validation steps ensure no regression in test coverage or functionality. The coder agent should follow the phased approach with clear checkpoints to verify success at each stage.

**Recommendation**: Proceed with implementation following the outlined phases and validation checkpoints.
