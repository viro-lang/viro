//go:build !webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

func RegisterWebUINatives(rootFrame core.Frame) {
	// WebUI not available, do nothing
}
