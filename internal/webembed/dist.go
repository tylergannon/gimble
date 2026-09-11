// Package webembed installs Gimble's embedded application into the runtime.
package webembed

import (
	"embed"
	"io/fs"

	"github.com/tylergannon/gimble/internal/webapp"
	"github.com/tylergannon/gimble/internal/webbridge"
)

// b* matches bare.txt in a bare checkout and also the ignored build directory
// after the frontend is built. The placeholder keeps this package compilable
// without checking generated frontend output into Git.
//go:embed all:b*
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
