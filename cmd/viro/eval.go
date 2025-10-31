package main

func runEval(cfg *Config) int {
	ctx := &ExecutionContext{
		Config:      cfg,
		Input:       &ExprInput{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin},
		Args:        []string{},
		PrintResult: !cfg.NoPrint,
		ParseOnly:   false,
	}

	_, exitCode := executeViroCode(ctx)
	return exitCode
}
