package native

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Print implements the `print` native.
//
// Contract: print value
// - Accepts any value
// - For blocks: reduce elements (evaluate each) and join with spaces
// - Writes result to stdout followed by newline
// - Returns none
func Print(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"print", "1", formatInt(len(args))},
		)
	}

	output, err := buildPrintOutput(args[0], eval)
	if err != nil {
		return value.NoneVal(), err
	}

	if _, writeErr := fmt.Fprintln(os.Stdout, output); writeErr != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("print output error: %v", writeErr), "", ""},
		)
	}

	return value.NoneVal(), nil
}

// Input implements the `input` native.
//
// Contract: input
// - Reads a line from stdin
// - Returns the line as string value without trailing newline
func Input(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 0 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"input", "0", formatInt(len(args))},
		)
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return value.NoneVal(), verror.NewAccessError(
				verror.ErrIDInvalidOperation,
				[3]string{fmt.Sprintf("input read error: %v", err), "", ""},
			)
		}
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return value.StrVal(line), nil
}

func buildPrintOutput(val value.Value, eval Evaluator) (string, *verror.Error) {
	if val.Type == value.TypeBlock {
		blk, ok := val.AsBlock()
		if !ok {
			return "", verror.NewInternalError("block value missing payload in print", [3]string{})
		}

		if len(blk.Elements) == 0 {
			return "", nil
		}

		parts := make([]string, 0, len(blk.Elements))
		for idx, elem := range blk.Elements {
			evaluated, err := eval.Do_Next(elem)
			if err != nil {
				if err.Near == "" {
					err.SetNear(verror.CaptureNear(blk.Elements, idx))
				}
				return "", err
			}
			parts = append(parts, valueToPrintString(evaluated))
		}
		return strings.Join(parts, " "), nil
	}

	return valueToPrintString(val), nil
}

func valueToPrintString(val value.Value) string {
	if val.Type == value.TypeString {
		if str, ok := val.AsString(); ok {
			return str.String()
		}
	}
	return val.String()
}
