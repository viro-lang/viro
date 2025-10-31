package main

func runCheck(cfg *Config) int {
	ctx := &ExecutionContext{
		Config:      cfg,
		Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
		Args:        nil,
		PrintResult: false,
		ParseOnly:   true,
	}

	_, exitCode := executeViroCode(ctx)
	return exitCode
}
