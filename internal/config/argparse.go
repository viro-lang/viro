package config

import "strings"

var flagsWithValues = map[string]bool{
	"-c":             true,
	"--sandbox-root": true,
	"--history-file": true,
	"--prompt":       true,
}

type ParsedArgs struct {
	ScriptIdx   int
	ReplArgsIdx int
	ScriptArgs  []string
}

func splitCommandLineArgs(args []string) *ParsedArgs {
	result := &ParsedArgs{
		ScriptIdx:   -1,
		ReplArgsIdx: -1,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			result.ReplArgsIdx = i
			break
		}

		if flagsWithValues[arg] {
			if i+1 < len(args) {
				i++
			}
			continue
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			result.ScriptIdx = i
			break
		}
	}

	return result
}
