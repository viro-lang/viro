package bootstrap

import (
	"embed"
	"io/fs"
)

//go:embed *.viro
var bootstrapFS embed.FS

func Files() fs.FS {
	return bootstrapFS
}
