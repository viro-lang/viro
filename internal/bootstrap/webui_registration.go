//go:build webui

package bootstrap

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
)

func registerWebUINativesIfAvailable(rootFrame core.Frame) {
	native.RegisterWebUINatives(rootFrame)
}
