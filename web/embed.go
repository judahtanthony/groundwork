// Package webui exposes the production web UI assets embedded in the gw binary.
package webui

import (
	"embed"
	"io/fs"
)

// embeddedDist contains the checked-in Vite production build. Keeping dist in
// the repository lets ordinary Go builds remain Node-free.
//
//go:embed dist
var embeddedDist embed.FS

// Dist is rooted at dist/ so callers can serve index.html directly.
var Dist = mustSub(embeddedDist, "dist")

func mustSub(root fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
