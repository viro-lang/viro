package main

func runScript(cfg *Config) int {
	ctx := &ExecutionContext{
		Config:      cfg,
		Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
		Args:        cfg.Args,
		PrintResult: false,
		ParseOnly:   false,
	}

	_, exitCode := executeViroCode(ctx)
	return exitCode
}
