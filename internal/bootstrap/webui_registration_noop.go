//go:build !webui

package bootstrap

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

func registerWebUINativesIfAvailable(rootFrame core.Frame) {
	// WebUI not available, do nothing
}
