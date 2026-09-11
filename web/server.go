// NewHandler assembles the server. It is here, beside the embedded build and
// not in package gimble, because the page's remote functions import gimble
// and gimble cannot import the page. It exists so that there is exactly one
// production stack: the binary in cmd and any test beside this file both call
// NewHandler, and neither can be green over a composition the other does not
// use.
package web

import (
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tylergannon/skgo"

	"github.com/tylergannon/gimble/generated"
)

// NewHandler builds the server over the frontend build in dist. With a
// non-empty proxy it renders pages from a running `vp dev` server; otherwise it
// serves the embedded build. Loads, remote functions and server routes sit in
// front of pages in both modes, so Kit's generated throwing stubs never answer
// an application request.
func NewHandler(dist fs.FS, proxy, origin string) (http.Handler, string, error) {
	// The manifest is read in both modes: it is where appDir and base come
	// from, and those decide the URL prefix remote calls arrive on.
	manifest, err := skgo.ReadManifest(dist)
	if err != nil {
		return nil, "", err
	}
	if proxy != "" {
		manifest, err = skgo.ReadDevManifest(dist, proxy)
		if err != nil {
			return nil, "", err
		}
	}

	remoteCfg := manifest.RemoteConfig(origin)
	loadCfg := manifest.LoadConfig(origin)
	endpointCfg := manifest.EndpointConfig(origin)

	mode := "prod"
	var pages http.Handler
	// build makes the page handler once all three registries exist. The
	// renderer needs the loads and remotes directly, and its event.fetch host
	// needs the endpoint registry.
	var build func(*skgo.Loads, *skgo.Remotes) (http.Handler, error)
	var endpoints *skgo.Endpoints
	if proxy != "" {
		target, err := url.Parse(proxy)
		if err != nil {
			return nil, "", err
		}
		// In dev the client is served by vite, not by this build, so its baked
		// version differs from the manifest's — sending x-sveltekit-version
		// would make it reload in a loop. Kit skips the remote origin check in
		// dev too.
		remoteCfg = manifest.RemoteConfig("")
		remoteCfg.Version = ""
		remoteCfg.Dev = true
		remoteCfg.CookieOrigin = origin
		loadCfg.Version = ""
		loadCfg.Dev = true
		endpointCfg.Dev = true
		mode = "dev"
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
		// Pages are rendered here, in this process, by the SSR bundle the
		// adapter built. The renderer needs the loads and the remote functions
		// it will answer during a render, so the page handler is built after
		// both of them exist.
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

	remotes, err := skgo.NewRemotes(remoteCfg, generated.Remotes()...)
	if err != nil {
		return nil, "", err
	}
	loads, err := skgo.NewLoads(loadCfg, generated.Loads()...)
	if err != nil {
		return nil, "", err
	}
	endpoints, err = skgo.NewEndpoints(endpointCfg, generated.Endpoints()...)
	if err != nil {
		return nil, "", err
	}
	if pages, err = build(loads, remotes); err != nil {
		return nil, "", err
	}
	return explainOriginRefusals(remoteCfg,
		loads.Intercept(remotes.Intercept(endpoints.Intercept(pages)))), mode, nil
}

// explainOriginRefusals turns the one failure a new app is most likely to hit
// into a message that says what to do about it.
//
// A command is a POST, and skgo refuses a non-GET remote call whose `Origin`
// header is not the origin the server was configured with — the same check kit
// makes. That is a bare 403 with no body: in the browser every command fails
// and nothing says why. The origin is decided in two places that have to
// agree — the frontend is built with it, the binary is run with it — so this
// wrapper watches for the mismatch and writes down both values and the fix.
func explainOriginRefusals(cfg skgo.RemoteConfig, next http.Handler) http.Handler {
	if cfg.Origin == "" {
		return next
	}
	appDir := cfg.AppDir
	if appDir == "" {
		appDir = "_app"
	}
	prefix := cfg.Base + "/" + appDir + "/remote/"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Origin")
		if r.Method == http.MethodGet || got == cfg.Origin || !strings.HasPrefix(r.URL.Path, prefix) {
			next.ServeHTTP(w, r)
			return
		}
		msg := "This app's origin is " + cfg.Origin + ", but the request came from " +
			quoteOrigin(got) + ", so it was refused.\n\n" +
			"The origin is fixed when the frontend is built and checked again when the\n" +
			"binary runs. Either browse the app at " + cfg.Origin + ", or set ORIGIN in\n" +
			"Justfile to the origin you are using and run `just build` again.\n"
		log.Printf("skgo: refused %s %s: Origin %s, want %s. Rebuild with ORIGIN=%s or browse the app at %s",
			r.Method, r.URL.Path, quoteOrigin(got), cfg.Origin, quoteOrigin(got), cfg.Origin)
		http.Error(w, msg, http.StatusForbidden)
	})
}

func quoteOrigin(origin string) string {
	if origin == "" {
		return "(no Origin header)"
	}
	return origin
}
