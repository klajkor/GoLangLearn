// Package web embeds the built frontend (web/dist) into the Go binary.
// Run `npm run build` in this directory to refresh it.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built frontend with the "dist" prefix removed.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err) // only fails if the embed pattern is wrong
	}
	return sub
}
