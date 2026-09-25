package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distEmbedFS embed.FS

// DistFS returns the embedded Svelte 5 production bundle filesystem rooted at web/dist.
func DistFS() (fs.FS, error) {
	return fs.Sub(distEmbedFS, "dist")
}
