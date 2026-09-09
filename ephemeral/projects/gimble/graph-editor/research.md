# Embedding a SvelteKit static build in a Go binary

## Sources

Docs (current):
- adapter-static: https://svelte.dev/docs/kit/adapter-static
- Single-page apps: https://svelte.dev/docs/kit/single-page-apps
- Page options (prerender, ssr, trailingSlash): https://svelte.dev/docs/kit/page-options
- Configuration (paths.relative, appDir, prerender): https://svelte.dev/docs/kit/configuration
- sv create: https://svelte.dev/docs/cli/sv-create
- Go embed rules: https://pkg.go.dev/embed
- Issue clarifying ssr + prerender: https://github.com/sveltejs/kit/issues/14471 (opened 2025-09, closed by docs PR #14560)

Projects and articles that do SvelteKit + adapter-static + embed.FS:
- https://github.com/mateo08c/GinSvelteEmbed (Svelte 5, Kit 2, Vite 5; the newest example found)
- https://github.com/saas-templates/go-sveltekit-spa (chi router; prerendered landing page plus SPA under /app)
- https://www.liip.ch/en/blog/embed-sveltekit-into-a-go-binary with code at https://github.com/munxar/goapi (fully prerendered, no fallback)
- https://dev.to/aryaprakasa/serving-single-page-application-in-a-single-binary-file-with-go-12ij with code at https://github.com/aprakasa/go-embed-spa

No repository from 2025 or 2026 was found that does exactly this; the newest is GinSvelteEmbed on Kit 2 / Svelte 5. The pattern has not changed since Kit 1, and the current docs confirm every piece below.

## Current versions (npm registry, latest tag)

svelte 5.57.0, @sveltejs/kit 2.70.3, @sveltejs/adapter-static 3.0.10, vite 8.2.2, sv 0.17.0. Local Go is 1.27.1; embed needs 1.16 or newer.

## Create command

```sh
npx sv create web/editor --template minimal --types ts --no-add-ons --install pnpm
pnpm add -D @sveltejs/adapter-static
```

Flags per the sv create docs: `--template minimal|demo|library`, `--types ts|jsdoc` or `--no-types`, `--add <addons...>` or `--no-add-ons`, `--install npm|pnpm|yarn|bun|deno` or `--no-install`, `--no-dir-check`. `sv create` scaffolds with adapter-auto; swap it for adapter-static in svelte.config.js.

## What the projects configure

GinSvelteEmbed, svelte.config.js:

```js
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({ pages: 'build', assets: 'build', fallback: 'index.html', precompress: false }),
    prerender: { handleHttpError: 'ignore' }
  }
};
```

Its root layout exports `ssr = false` and `prerender = false`, so nothing is prerendered and every URL is served by the fallback `index.html`. The Go side does `//go:embed build/*`, `fs.Sub(buildFS, "build")`, and a Gin middleware that skips `/api` and serves `index.html` for any path that is not a file.

go-sveltekit-spa uses `adapter({ fallback: 'app/index.html' })` with `prerender = true` in the root layout, so `/` is a prerendered `build/index.html` and everything under `/app/*` is the SPA. Go: `//go:embed all:build/**`, `fs.Sub`, `http.FileServer(http.FS(appFS))` at `/`, and a handler on `/app` that writes `build/app/index.html`.

The Liip article uses no fallback at all: `prerender = true`, `trailingSlash: 'always'`, `//go:embed all:build`, and a wrapper around `http.FileServer` that opens the path, and on `os.ErrNotExist` retries with `.html` appended. The dev.to article uses `fallback: "index.html"` and three `//go:embed` lines to pull in `build/_app/...` because a plain `build/*` pattern drops the underscore-prefixed `_app` directory.

## Recommended shape for tractor edit

svelte.config.js:

```js
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      pages: '../../internal/editor/dist',
      assets: '../../internal/editor/dist',
      fallback: 'index.html',
      precompress: false,
      strict: true
    })
  }
};
```

src/routes/+layout.ts:

```ts
export const ssr = false;
export const prerender = true;
```

Go, in internal/editor:

```go
package editor

import ("embed"; "io/fs"; "net/http"; "path"; "strings")

//go:embed all:dist
var dist embed.FS

func Handler() http.Handler {
    sub, _ := fs.Sub(dist, "dist")
    files := http.FS(sub)
    server := http.FileServer(files)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
        if p == "" { p = "index.html" }
        if f, err := files.Open("/" + p); err == nil {
            f.Close()
            if strings.HasPrefix(p, "_app/immutable/") {
                w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
            }
            server.ServeHTTP(w, r)
            return
        }
        w.Header().Set("Cache-Control", "no-cache")
        r.URL.Path = "/"
        server.ServeHTTP(w, r)
    })
}
```

Mount with `mux.Handle("/api/", api)` first and `mux.Handle("/", editor.Handler())` last.

## Option reference (adapter-static 3.x)

- `pages` (default `build`): directory for prerendered HTML.
- `assets` (default = pages): directory for `_app`, static files. Keep equal to pages.
- `fallback` (default none): file name to generate from app.html for SPA routing; `index.html`, `200.html`, or `404.html`. Any name works; the Go handler decides what to serve.
- `precompress` (default false): also emit `.br` and `.gz` beside each file.
- `strict` (default true): fail the build if any route is neither prerendered nor covered by a fallback.

## Pitfalls

- `ssr = false` with `prerender = true` makes the prerendered HTML files empty shells (the fallback and every prerendered page are the same: app.html plus the script tags). Issue #14471 got this written into the docs. For a browser-only tool that is fine and is what the brief asks for; there is nothing on those pages a crawler needs. Keep `prerender = true` so `strict` still validates the route list and `index.html` is emitted even if `fallback` were dropped.
- `fallback` vs full prerender: with only one route (`/`) there is no difference in output. With deep links such as `/nodes/foo` the Go handler must serve index.html for unknown paths, so set `fallback: 'index.html'` and treat it as the SPA entry. Without a fallback, adapter-static's `strict` fails on any dynamic route, and the Liip `.html`-suffix trick is required instead.
- `paths.relative` defaults to true, but the docs state the fallback page always uses absolute asset paths (`/_app/...`). So the editor must be served at `/`, not under a prefix, unless `kit.paths.base` is set and the Go mux strips it.
- `trailingSlash` defaults to `'never'`, so `/about` becomes `about.html`; `'always'` makes `about/index.html`. Irrelevant for a single-route SPA served by fallback, but if more routes are prerendered pick `'always'` so `http.FileServer` finds `index.html` inside directories without a suffix rewrite. `http.FileServer` also redirects `/index.html` to `/`; do not point the fallback path at `/index.html` or you get a redirect loop, hence `r.URL.Path = "/"` above.
- `//go:embed dist` skips files starting with `.` or `_`, so `_app` disappears silently. Use the `all:` prefix. Every pattern must match at least one file or non-empty directory or the build fails, so `internal/editor/dist` must be committed with real contents (the brief already says the build output is committed); an empty `.gitkeep`-only directory does not satisfy `all:dist` either because `.gitkeep` is hidden and the directory counts as empty for the non-`all:` rule; with `all:` it counts but then `index.html` is missing and the server 404s. Commit the real build.
- `precompress: true` produces `.br`/`.gz` siblings that `http.FileServer` does not negotiate; it would serve them only when requested by name. Leave it false and, if compression matters, wrap the handler with a gzip middleware or skip it entirely for 127.0.0.1.
- Cache headers: `http.FileServer` sets `Last-Modified` from the embed FS mod time, which is zero, so it emits no `Last-Modified` and no `ETag`. Set `Cache-Control: immutable` for `_app/immutable/*` (hashed names) and `no-cache` for `index.html`, as in the handler above.
- Dev workflow: `vite dev` serves the app itself; proxy `/api` to the Go server in vite.config.ts (`server.proxy['/api'] = 'http://127.0.0.1:PORT'`) as GinSvelteEmbed does.
