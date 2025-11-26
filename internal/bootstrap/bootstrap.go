package bootstrap

import (
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"

	"github.com/marcin-radoszewski/viro/bootstrap"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/debug"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func wrapBootstrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*verror.Error); ok {
		return verr
	}
	return verror.NewBootstrapError(verror.ErrIDBootstrapFailure, [3]string{context, err.Error(), ""})
}

func LoadAndExecuteBootstrapScripts(evaluator *eval.Evaluator) error {
	return LoadAndExecuteBootstrapScriptsFromFS(evaluator, bootstrap.Files())
}

func LoadAndExecuteBootstrapScriptsFromFS(evaluator *eval.Evaluator, bootstrapFS fs.FS) error {
	var scripts []string

	err := fs.WalkDir(bootstrapFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".viro") {
			scripts = append(scripts, path)
		}
		return nil
	})
	if err != nil {
		return wrapBootstrapError(err, "walk bootstrap filesystem")
	}

	sort.Strings(scripts)

	if len(scripts) > 0 && scripts[0] != "init.viro" {
		for i, script := range scripts {
			if script == "init.viro" {
				scripts[0], scripts[i] = scripts[i], scripts[0]
				break
			}
		}
	}

	for _, script := range scripts {
		content, err := fs.ReadFile(bootstrapFS, script)
		if err != nil {
			return wrapBootstrapError(err, fmt.Sprintf("read bootstrap script %s", script))
		}

		sourceName := fmt.Sprintf("bootstrap/%s", script)
		values, locations, err := parse.ParseWithSource(string(content), sourceName)
		if err != nil {
			return wrapBootstrapError(err, fmt.Sprintf("parse bootstrap script %s", script))
		}

		_, err = evaluator.DoBlock(values, locations)
		if err != nil {
			return wrapBootstrapError(err, fmt.Sprintf("execute bootstrap script %s", script))
		}
	}

	return nil
}

func InitTrace(profile bool) error {
	return InitTraceWithOutput(profile, "")
}

func InitTraceWithOutput(profile bool, output string) error {
	if profile {
		return trace.InitTraceSilent()
	}
	return trace.InitTrace(output, 50) // default 50MB max size
}

func InitDebugger() {
	debug.InitDebugger()
}

func NewEvaluatorWithNatives(stdout, stderr io.Writer, stdin io.Reader, quiet bool) (*eval.Evaluator, error) {
	evaluator := eval.NewEvaluator()

	if quiet {
		evaluator.SetOutputWriter(io.Discard)
	} else if stdout != nil {
		evaluator.SetOutputWriter(stdout)
	}
	if stderr != nil {
		evaluator.SetErrorWriter(stderr)
	}
	if stdin != nil {
		evaluator.SetInputReader(stdin)
	}

	rootFrame := evaluator.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, evaluator)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)
	native.RegisterBitwiseNatives(rootFrame)

	if err := LoadAndExecuteBootstrapScripts(evaluator); err != nil {
		return nil, err
	}

	return evaluator, nil
}

func InjectSystemArgs(evaluator core.Evaluator, args []string) {
	viroArgs := make([]core.Value, len(args))
	for i, arg := range args {
		viroArgs[i] = value.NewStringValue(arg)
	}

	argsBlock := value.NewBlockValue(viroArgs)

	ownedFrame := frame.NewFrame(frame.FrameObject, -1)
	ownedFrame.Bind("args", argsBlock)

	systemObj := value.NewObject(ownedFrame)

	rootFrame := evaluator.GetFrameByIndex(0)
	rootFrame.Bind("system", systemObj)
}
