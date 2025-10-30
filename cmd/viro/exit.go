package main

import (
	"fmt"
	"os"

	"github.com/marcin-radoszewski/viro/internal/verror"
)

const (
	ExitSuccess   = 0
	ExitError     = 1
	ExitSyntax    = 2
	ExitAccess    = 3
	ExitUsage     = 64
	ExitInternal  = 70
	ExitInterrupt = 130
)

func handleError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintf(os.Stderr, "%v", vErr)
		return categoryToExitCode(vErr.Category)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return ExitError
}

func categoryToExitCode(cat verror.ErrorCategory) int {
	switch cat {
	case verror.ErrSyntax:
		return ExitSyntax
	case verror.ErrAccess:
		return ExitAccess
	case verror.ErrInternal:
		return ExitInternal
	default:
		return ExitError
	}
}

func printParseError(err error) {
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintf(os.Stderr, "%v", vErr)
	} else {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
	}
}

func printRuntimeError(err error) {
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintf(os.Stderr, "%v", vErr)
	} else {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
	}
}
