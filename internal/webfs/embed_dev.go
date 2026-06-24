//go:build !prod

package webfs

import "net/http"

func FS() http.FileSystem {
	return nil
}
