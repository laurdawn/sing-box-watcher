//go:build prod

package webfs

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var webDist embed.FS

func FS() http.FileSystem {
	sub, _ := fs.Sub(webDist, "dist")
	return http.FS(sub)
}
