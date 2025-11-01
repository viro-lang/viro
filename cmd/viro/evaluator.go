package main

import (
	"io"

	"github.com/marcin-radoszewski/viro/internal/api"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
)

func setupEvaluator(cfg *api.Config) *eval.Evaluator {
	evaluator := eval.NewEvaluator()

	if cfg.Quiet {
		evaluator.OutputWriter = io.Discard
	}

	rootFrame := evaluator.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, evaluator)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	return evaluator
}
