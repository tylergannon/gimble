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
// With a non-empty proxy, Go renders pages from modules supplied by a running
// `vp dev` server; otherwise it renders from the embedded production bundle.
//
// The order is kit's own dispatch order, turned inside out into middleware:
// the handle hook runs first, on every request, and puts the store where the
// loads and remote functions can read it; then loads, remotes and endpoints;
// then Go-rendered pages and static assets.
func NewHandler(store *editor.Store, dist fs.FS, proxy, origin string) (http.Handler, error) {
	manifest, err := skgo.ReadManifest(dist)
	if err != nil {
		return nil, err
	}
	if proxy != "" {
		manifest, err = skgo.ReadDevManifest(dist, proxy)
		if err != nil {
			return nil, err
		}
	}

	remoteCfg := manifest.RemoteConfig(origin)
	loadCfg := manifest.LoadConfig(origin)
	endpointCfg := manifest.EndpointConfig(origin)
	handleCfg := manifest.HandleConfig()

	var pages http.Handler
	var build func(*skgo.Loads, *skgo.Remotes) (http.Handler, error)
	var endpoints *skgo.Endpoints
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
		endpointCfg.Dev = true
		handleCfg.Version = ""
		build = func(loads *skgo.Loads, remotes *skgo.Remotes) (http.Handler, error) {
			ssr, err := skgo.NewDevSSR(dist, manifest, loads, remotes, proxy, skgo.SSROptions{
				Fetch: endpoints.Intercept(http.NotFoundHandler()),
			})
			if err != nil {
				return nil, err
			}
			return skgo.NewDevPages(target, manifest, ssr, log.Printf, endpoints), nil
		}
	} else {
		build = func(loads *skgo.Loads, remotes *skgo.Remotes) (http.Handler, error) {
			ssr, err := skgo.NewSSR(dist, manifest, loads, remotes, skgo.SSROptions{
				Fetch: endpoints.Intercept(http.NotFoundHandler()),
			})
			if err != nil {
				return nil, err
			}
			return skgo.NewStaticHandler(dist, skgo.WithSSR(ssr))
		}
	}

	// The handle hook is the one place the request learns which file the
	// editor is open on.
	handle := skgo.Handle(func(ctx context.Context) error {
		return skgo.SetLocal(ctx, store)
	})

	remotes, err := skgo.NewRemotes(remoteCfg, generated.Remotes()...)
	if err != nil {
		return nil, err
	}
	loads, err := skgo.NewLoads(loadCfg, generated.Loads()...)
	if err != nil {
		return nil, err
	}
	endpoints, err = skgo.NewEndpoints(endpointCfg, generated.Endpoints()...)
	if err != nil {
		return nil, err
	}
	pages, err = build(loads, remotes)
	if err != nil {
		return nil, err
	}
	app := handle.Intercept(handleCfg,
		loads.Intercept(remotes.Intercept(endpoints.Intercept(pages))))
	return editor.LoopbackOnly(app), nil
}
