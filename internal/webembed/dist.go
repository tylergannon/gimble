// Package webembed installs Gimble's embedded application into the runtime.
package webembed

import (
	"embed"
	"io/fs"

	"github.com/tylergannon/gimble/internal/webapp"
	"github.com/tylergannon/gimble/internal/webbridge"
)

//go:embed all:build
var build embed.FS

func init() {
	webbridge.Register(func(proxy, origin string) (handler webbridge.Handler, mode string, err error) {
		dist, err := fs.Sub(build, "build")
		if err != nil {
			return nil, "", err
		}
		return webapp.NewHandler(dist, proxy, origin)
	})
}
