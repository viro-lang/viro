package integration

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func mustConfigFromArgs(t *testing.T, args []string) *api.Config {
	t.Helper()
	cfg, err := api.ConfigFromArgs(args)
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}
	return cfg
}
