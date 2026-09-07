// Package server assembles the editor's HTTP stack over skgo. It exists so
// that there is exactly one composition: `tractor edit` and the tests beside
// this file both call NewHandler, and neither can be green over a stack the
// other does not use.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/tylergannon/skgo"

	"github.com/tylergannon/tractor/internal/editor"
	"github.com/tylergannon/tractor/internal/editor/generated"
	web "github.com/tylergannon/tractor/web/editor"
)

// Dist returns the embedded editor build, rooted where the manifest is.
func Dist() (fs.FS, error) {
	dist, err := fs.Sub(web.Build, "build")
	if err != nil {
		return nil, fmt.Errorf("open the embedded editor build: %w", err)
	}
	return dist, nil
}

// NewHandler builds the editor's server for one pipeline over the build in
// dist. origin is the URL browsers reach the editor at; skgo refuses a
// command (a POST) from any other origin, which is kit's own cross-site rule.
// With a non-empty proxy, pages are forwarded to a running `vp dev` server
// instead of being served from dist; remote functions are answered here in
// both modes.
//
// The order is kit's own dispatch order, turned inside out into middleware:
// the handle hook runs first, on every request, and puts the store where the
// remote functions can read it; then the remote functions; then pages.
func NewHandler(store *editor.Store, dist fs.FS, proxy, origin string) (http.Handler, error) {
	manifest, err := skgo.ReadManifest(dist)
	if err != nil {
		return nil, err
	}

	remoteCfg := manifest.RemoteConfig(origin)
	loadCfg := manifest.LoadConfig(origin)

	var pages http.Handler
	if proxy != "" {
		target, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		// In dev the client is served by vite, not by this build, so its
		// baked version differs from the manifest's; sending
		// x-sveltekit-version would make it reload in a loop. Kit skips the
		// cross-site check in dev too.
		remoteCfg = manifest.RemoteConfig("")
		remoteCfg.Version = ""
		remoteCfg.Dev = true
		remoteCfg.CookieOrigin = origin
		loadCfg.Version = ""
		loadCfg.Dev = true
		pages = skgo.NewDevProxy(target, log.Printf)
	} else {
		static, err := skgo.NewStaticHandler(dist)
		if err != nil {
			return nil, err
		}
		pages = static
	}

	// The handle hook is the one place the request learns which file the
	// editor is open on.
	loadCfg.Handle = func(ctx context.Context) error {
		return skgo.SetLocal(ctx, store)
	}

	remotes, err := skgo.NewRemotes(remoteCfg, generated.Remotes()...)
	if err != nil {
		return nil, err
	}
	loads, err := skgo.NewLoads(loadCfg, generated.Loads()...)
	if err != nil {
		return nil, err
	}
	return editor.LoopbackOnly(loads.Intercept(remotes.Intercept(pages))), nil
}
