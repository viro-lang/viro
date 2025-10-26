---
description: Generate a test summary with total tests, failed tests, and grouped failure details
---

# Test Summary Command

Generate a comprehensive summary of test results across the entire Viro codebase.
Run the tests with this command:

```bash
go test -json ./...
```

Analyze the tests output and errors to be able to generate the summary in required output.

You can store the test results in temporary file and use it instead of restarting tests multiple times. Remember to remove the file after you are finished. The file is going to be very large so use `head` to understand its structure and then use `jq` tool to work with it.

DO NOT edit any files.

## Output

The command provides:
- **Total test count**: Number of tests executed
- **Passed tests**: Number of successful tests
- **Failed tests**: Number of failing tests
- **Grouped failure details**: Failures organized by identified issue type for easier debugging
