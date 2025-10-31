package main

import "strings"

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

		if arg == "-c" {
			if i+1 < len(args) {
				i++
			}
			continue
		}

		if arg == "--sandbox-root" || arg == "--history-file" || arg == "--prompt" {
			if i+1 < len(args) {
				i++
			}
			continue
		}

		if !strings.HasPrefix(arg, "-") {
			result.ScriptIdx = i
			break
		}
	}

	return result
}
