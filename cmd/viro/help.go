package main

import "fmt"

func printHelp() {
	fmt.Print(`Viro - A REBOL-inspired programming language

USAGE:
    viro [OPTIONS] [FILE [ARGS...]]
    viro [OPTIONS] -- [ARGS...]
    viro -c EXPRESSION
    viro --check FILE
    viro --version
    viro --help

MODES:
    (default)           Start interactive REPL
    FILE [ARGS...]      Execute script file with arguments
    -- [ARGS...]        Start REPL with arguments in system.args
    -c EXPRESSION       Evaluate expression and print result
    --check FILE        Check syntax without executing

GLOBAL OPTIONS:
    --sandbox-root PATH        Sandbox root for file operations (default: current directory)
    --allow-insecure-tls       Disable TLS certificate verification (warning: security risk)
    --quiet                    Suppress non-error output
    --verbose                  Enable verbose output
    --help                     Show this help message
    --version                  Show version information

EVAL OPTIONS:
    --stdin                    Read additional input from stdin
    --no-print                 Don't print result of evaluation

REPL OPTIONS:
    --no-history               Disable command history
    --history-file PATH        History file location
    --prompt STRING            Custom REPL prompt
    --no-welcome               Skip welcome message

ENVIRONMENT VARIABLES:
    VIRO_SANDBOX_ROOT          Default sandbox root directory
    VIRO_ALLOW_INSECURE_TLS    Allow insecure TLS (set to "1" or "true")
    VIRO_HISTORY_FILE          REPL history file location

EXIT CODES:
    0     Success
    1     General error (script/math error)
    2     Syntax error (parse failure)
    3     Access error (permission denied, sandbox violation)
    64    Usage error (invalid CLI arguments)
    70    Internal error (interpreter crash)
    130   Interrupted (Ctrl+C)

EXAMPLES:
    # Start REPL
    viro

    # Start REPL with arguments
    viro -- arg1 arg2 arg3

    # Execute script with arguments
    viro script.viro arg1 arg2

    # Check syntax
    viro --check script.viro

    # Evaluate expression
    viro -c "3 + 4"

    # Use in pipeline
    echo "[1 2 3]" | viro -c "first" --stdin

    # Suppress output
    viro -c "pow 2 10" --no-print

    # REPL with arguments for testing
    viro -- user@example.com admin
    >> print ["Email:" first system.args]
    >> print ["Role:" last system.args]

For more information, visit: https://github.com/marcin-radoszewski/viro
`)
}
