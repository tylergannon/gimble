# Serving a Go application's SvelteKit page with skgo

This is a field report from porting one real application, tractor's graph
editor (`tractor edit`), from a hand-written `net/http` JSON API plus
adapter-static onto skgo. It is written for the skgo user manual: what a Go
developer has to do, in order, what each step produces, and where the sharp
edges are. Everything here was done and run; nothing is inferred from
documentation.

The application before the port: a Go server embedding a kit 2 static build,
serving `GET /api/doc`, `PUT /api/doc` and a hand-rolled server-sent-events
stream on `/api/events`, with a TypeScript `api.ts` that called them with
`fetch` and `EventSource`. After the port: the same page, a kit 3 app built
by the skgo adapter, calling three remote functions that are ordinary Go
functions declared beside the route. The Go server is skgo's static handler
and remote registry behind one loopback guard.

## What skgo gives you, in one paragraph

You write a `*.remote.go` file next to a SvelteKit route. In it are plain Go
functions of the shape `func(ctx, arg) (result, error)`, marked with
`skgo.Query`, `skgo.Command`, or `skgo.LiveQuery`. `go generate` then writes
the `.remote.ts` module the page imports, the TypeScript types for every Go
struct that crosses the wire, and a Go registration your server mounts. The
page calls the functions exactly as it would call kit remote functions;
kit's own client makes the requests; Go answers them. Every body in the
generated `.remote.ts` throws, so a value on screen is proof that Go
answered.

## Installing

### Toolchain

skgo's `go.mod` says `go 1.27.1`, so the consuming module's `go` directive
becomes `1.27.1` when you require it. Have that toolchain installed before
you start. Node is a build-time dependency only; skgo's shape uses Node
24.16.0 and vite-plus (`vp`) 0.3.0, provisioned by mise from a `mise.toml`
in the web directory. `mise install` in that directory fetches both.

### Requiring a private module

skgo was private when this was done. A `replace` directive to a local
checkout works but breaks every other machine. What worked instead, with no
`replace`, was letting Go fetch the module over git:

```sh
export GOPRIVATE=github.com/tylergannon/skgo
go get github.com/tylergannon/skgo@<commit>
go get -tool github.com/tylergannon/skgo/cmd/skgo@<commit>
go get -tool github.com/tylergannon/polytype/polytype@v1.0.0-rc.10
go mod tidy
```

`GOPRIVATE` turns off the module proxy and checksum database for that path,
so Go clones from GitHub with your git credentials and computes the checksum
locally. The pseudo-version it resolves (`v0.0.0-20260907023726-c64663c68c74`)
and the `go.sum` lines are the same ones the public proxy will produce once
the repository is public, so nothing needs to change at that point except
that `GOPRIVATE` becomes unnecessary. The first `go get` failed once with
`fatal: shallow file has changed since we read it` from the module cache's
git clone; running it again succeeded.

Both `skgo` and `polytype` go in as `tool` directives. `go tool skgo` and
`go tool polytype` then build from the module cache, so neither has to be a
writable checkout and `go generate` works on a fresh clone.

### Pinning, and why the adapter must match the module

The one file skgo asks you to copy into your app is `skgo-adapter.js`, the
SvelteKit adapter. Its Go side reads the manifest that this JavaScript
writes, so the two must come from the same skgo version. Take it from the
module cache of the version you pinned:

```sh
cp "$(go env GOMODCACHE)/github.com/tylergannon/skgo@<version>/internal/newapp/template/web/skgo-adapter.js" web/
```

This bit me. I copied the adapter from the live `~/src/skgo` checkout, which
had moved to a newer commit while I worked. That adapter imported `esbuild`
and built an SSR bundle the pinned Go side knew nothing about, and `vp build`
failed with "Could not resolve 'esbuild'". The copy from the module cache
built at once.

skgo's `main` moved during this port from c64663c to 7c86073 (#43, in-process
SSR with an embedded goja engine). The editor stayed on c64663c on purpose:
the page runs with `ssr = false`, and the newer module adds goja and its
dependencies to every binary that imports skgo. Bumping is a deliberate
step: update the `require` and both `tool` lines, recopy the adapter from
the new module version, run `go generate`, rebuild the page, rebuild the
binary.

## The shape of an app

skgo's `skgo new` scaffold is the reference layout. Mirroring it inside an
existing repository looks like this for tractor:

```
web/editor/                       the vite root (package.json, src/, build/)
  mise.toml                       node + vp, and the build tasks
  package.json                    kit 3.0.0-next.25, vite aliased to vite-plus-core
  vite.config.ts                  sveltekit({ adapter: skgo(...) })
  skgo-adapter.js                 copied from the pinned skgo module
  skgo.remotes.json               written by go generate, read by the adapter
  dist.go                         package web; //go:embed all:build
  src/routes/editor.remote.go     the remote functions, package routes
  src/routes/editor.remote.ts     generated stub, every body throws
  src/routes/skgo_remotes_gen.go  generated registration for this package
  src/routes/go.mod               generated module boundary (see below)
  src/lib/skgo/editor/types.ts    generated TypeScript for the Go wire types
internal/editor/                  the Go package the remote functions call
  store.go                        the wire types and the file logic
  skgo_polytype_gen.go            generated, build-tagged jsonschema
  jsonschema_gen.go, jsonschema/  generated by polytype
  generated/config.go             //go:generate go tool skgo generate --web ../../../web/editor
  generated/skgo_bindings_gen.go  Remotes(), Loads(), Endpoints()
  generated/links/<base32>/       symlinks into src/routes (see below)
  generated/links.json            inventory of the link tree
  server/server.go                the one composition: NewHandler
```

Three things in that tree deserve explanation.

**The route tree is its own Go module.** SvelteKit names route directories
after URLs, so `[id]` and `(group)` are legal there and illegal in a Go
import path. `skgo generate` writes a `go.mod` into `src/routes` whose only
job is to stop `go build ./...` descending into the tree, and a "link tree"
of symlinks under `generated/links/` whose names Go can spell. The generated
bindings import the link, not the route. Commit the symlinks; they are
relative and git stores them fine. If your routes package has to be
imported by anything else, import it through its link path.

**The build directory is embedded from beside the page, not from the Go
package.** `web/editor/dist.go` is a Go package whose only content is
`//go:embed all:build`. Keeping the embed out of the package the remote
functions import matters on a fresh tree: `skgo generate` type-checks the
routes package with `packages.Load`, and an embed directive whose directory
does not exist yet is a type error. With the embed in its own package, a
tree that has never built the page still generates.

**The wire types live in the Go package that does the work**, not in the
routes package. The routes package imports `internal/editor`; the server
composition imports `generated`, which imports the routes package through
the link. If `internal/editor` also imported `generated` there would be a
cycle, so the composition sits in its own package, `internal/editor/server`.

## Writing the remote functions

```go
package routes

func getDoc(ctx context.Context) (editor.Document, error)
func saveDoc(ctx context.Context, request editor.SaveRequest) (editor.SaveResult, error)
func watchDoc(ctx context.Context, yield func(editor.Change) error) error

var (
	_ = skgo.Query(getDoc)
	_ = skgo.Command(saveDoc)
	_ = skgo.LiveQuery(watchDoc)
)
```

That is the whole declaration. The generator reads the real signatures with
`go/types`, so an argument is optional exactly as it is in kit, and a live
query is a function that pushes values through `yield` until its context is
cancelled, which happens when the browser disconnects.

### Reaching per-process state from a remote function

A remote function is a package-level function, so it needs a way to find
the thing it operates on. skgo's answer is kit's answer: the `handle` hook.
The server composition sets `LoadConfig.Handle` to a function that calls
`skgo.SetLocal(ctx, store)`, and every remote function reads it back with
`skgo.LocalOf[*editor.Store](ctx)`. The hook runs on every request the app
answers, remote calls included, and it needs the loads registry to be
mounted even when the app has no server loads at all. This is the "one
place to make a trust decision" the skgo docs describe; for an app with one
open file it is also simply where the file goes.

### What the generator refuses, and what to do instead

The type projection is polytype's, and it refuses rather than widens. These
came up in one small API:

- **No maps.** The layout sidecar is a JSON object keyed by node id. It
  travels as `[]Placement` with the id inside each element, and the Go side
  converts to and from the keyed map when it reads and writes the file. The
  page converts back to a record. The on-disk format did not change.
- **No pointers.** `*EdgeRef` became `[]string`, empty when absent. The
  reason given is exact: polytype would declare a non-nullable field while
  `encoding/json` writes `null` for a nil pointer.
- **No `json.RawMessage`** for the same reason as maps; it is bytes.
- **Nil slices cross the wire as `null`.** Initialise every slice field to
  an empty slice before returning, or the TypeScript type `Array<T>` is a
  lie at runtime. The store tests assert this.
- **A named type from another package** is projected into a declaration
  package skgo writes for it. Copying `lint.Diagnostic` and
  `engine.ModelResolution` into editor-owned wire structs kept the generated
  files, including a `Schema()` method on the type, out of tractor's core
  packages. That was a choice, not a requirement.

Doc comments on the wire structs become the JSDoc on the TypeScript types
and the descriptions in the JSON schema, so write them for the page's
reader.

### What `go generate` writes, and its permissions

One run wrote everything listed in the tree above. Two observations: the
`jsonschema/*.json` files come out mode 0600, which git does not record, so
it is harmless but surprising in `ls -l`; and a second run over an unchanged
tree changed nothing, which is the property that lets the output be
committed and checked in CI.

## Porting the page

### package.json and tsconfig

Copy the scaffold's devDependencies exactly. The important ones:
`@sveltejs/kit` 3.0.0-next.25, `typescript` 6.0.3 (not 7: kit's sync needs
`ts.sys`), `vite` aliased to `npm:@voidzero-dev/vite-plus-core@0.3.0`, and
`vite-plus` 0.3.0. `devEngines` pins pnpm 11.25.0 and `vp install` downloads
it. The tsconfig is `{ "extends": "$app/tsconfig" }`; the kit 2 form that
extends `./.svelte-kit/tsconfig.json` fails with "Tsconfig not found".

### `$lib` is `#lib`, and it wants extensions

Kit 3 removed `$lib`. The replacement is a Node subpath import,
`"imports": { "#lib/*": "./src/lib/*" }` in package.json. A blanket
`sed 's#\$lib/#\#lib/#'` got the page building, but `svelte-check` then
failed on every `#lib/model` with "Cannot find module". Subpath imports
resolve like Node, without extension probing: `#lib/model.ts` and
`#lib/store.svelte.ts` resolve, `#lib/model` does not. Vite is fine with the
`.ts` extension because kit's tsconfig sets `allowImportingTsExtensions`.
Imports of `.svelte`, `.css` and `.svg` files already carried extensions and
never broke.

### vite.config.ts

There is no `svelte.config.js`; everything goes to the `sveltekit()` plugin
flat. For a page that is committed and embedded:

```ts
sveltekit({
	adapter: skgo({ precompress: false }),
	version: { name: 'tractor' },
	experimental: { remoteFunctions: true },
	compilerOptions: { experimental: { async: true } }
})
```

`precompress: false` because the scaffold's default writes `.br` and `.gz`
beside every asset and the committed tree would triple. A constant
`version.name` because the default is a build timestamp, which would change
the boot document on every build. With both, `vp build` over unchanged
source is byte-identical, and CI can `git diff --exit-code` the build.

`paths.origin` was left unset. The scaffold sets it from `ORIGIN` because
kit 3 fixes the app's origin at build time, but nothing in a client-rendered
page reads it: kit's client builds remote URLs relative to `location`, and
skgo's cross-site check reads the origin you hand `RemoteConfig` at run
time. That is what lets `tractor edit` listen on a random port.

### The page's seam

`api.ts` used to be the HTTP contract. It is now a thin adapter over the
generated remote functions, and the rest of the page did not change. The
three shapes worth knowing:

- **A query is cached by its argument.** `await getDoc()` twice returns the
  cached value the second time, which is wrong for a file that changes on
  disk. `const q = getDoc(); await q.refresh(); return await q;` makes one
  request each time.
- **A command's result is the response.** `saveDoc` returns
  `{ saved, document }`; a stale version comes back as `saved: false` with
  the document on disk, which the page turns into its existing conflict
  path. A command may also name queries to refresh in the same flight, but
  this page gets everything it needs from the result.
- **A live query is an async iterable.** `watchDoc()[Symbol.asyncIterator]()`
  gives the imperative stream; `for await` or `iterator.next()` yields each
  value, kit reconnects with backoff if the connection drops, and
  `iterator.return()` releases it. This replaced a hand-rolled EventSource
  and its keepalive; skgo sends the SSE keepalive itself.

## The server

```go
manifest, _ := skgo.ReadManifest(dist)
remoteCfg := manifest.RemoteConfig(origin)
loadCfg := manifest.LoadConfig(origin)
loadCfg.Handle = func(ctx context.Context) error { return skgo.SetLocal(ctx, store) }
static, _ := skgo.NewStaticHandler(dist)
remotes, _ := skgo.NewRemotes(remoteCfg, generated.Remotes()...)
loads, _ := skgo.NewLoads(loadCfg, generated.Loads()...)
handler := loads.Intercept(remotes.Intercept(static))
```

`NewRemotes` refuses to start if the ids in the manifest are not exactly the
functions registered, so a binary built against a stale page fails at
startup rather than 404 in the browser. The static handler answers the boot
document for the routes in the manifest and a 404 with the boot document for
anything else, which is what a request for the old `/api/doc` now gets.

Two rules from kit that arrive with skgo and change how you run the server:

- **Non-GET remote calls are checked against one origin.** A command from
  any other origin is a 403 with no body. `tractor edit` picks a free port,
  so the origin is computed from the listener after `Listen` and is the URL
  printed for the user. Opening the same port as `localhost` shows the graph
  and fails every save. The page's own DNS-rebinding guard, a loopback Host
  check, stays as an outer handler.
- **Dev mode is Go proxying to vite, not vite proxying to Go.** With
  `--proxy http://127.0.0.1:5173`, pages come from `vp dev` and remote calls
  are still answered here; `RemoteConfig.Dev` turns off the version header
  and the origin check the way kit does in dev. The previous setup had vite
  proxying `/api` to Go; that direction is not supported, because kit's dev
  server would run the throwing stub for any remote call it received.

## Testing

Unit tests for the store call `Load`, `Save` and `Watch` directly. The
composition tests build the real handler over the embedded build and speak
kit's wire format, which is worth writing down because it is not JSON in
the usual sense:

- A query is `GET /_app/remote/<hash>/<name>`; with no argument there is no
  `payload` parameter. The response is `{"type":"result","data":"..."}` and
  `data` is a devalue string. `devalue.Parse` from polytype turns it into
  `*devalue.Object`; a query's value is at `q["<id>/"].v`, and also at `_`.
- A command is `POST` with a JSON body `{"payload": <base64url of the
  devalue string of the argument>, "refreshes": []}`. The first attempt sent
  the devalue string unencoded and got a 400. Build the argument with
  `devalue.NewObject(k, v, ...)` because a command payload preserves key
  order and refuses a Go map.
- A live query is `GET` with `Content-Type: text/event-stream`; the handler
  returns when the request context ends, so a test gives it a context with
  a short timeout and reads the recorder afterwards.
- The `<hash>/<name>` id comes from `generated.Remotes()` by `Name()`; do
  not hardcode the hash.

The test that mattered most was the browser. A Playwright script
(`proof/drive.mjs`) loaded the page, dragged a node and checked the sidecar
appeared, edited a field and checked the file, changed the file on disk and
watched the header follow, broke the YAML and saw the read-only banner, and
recorded every remote call and console error. The screenshots are in
`proof/`. The same script against `--proxy` mode proved the dev path.

## Rough edges to know about

- `go build ./...` and `go test ./...` walk `web/editor/node_modules`
  looking for Go packages. It costs a little time and finds nothing; the
  scaffold lives with it too.
- `vp build` prints a note suggesting `vpr build` when there is a `build`
  npm script. Both work; `vp build` is the built-in.
- The modernize linter in tractor's pre-commit hook rewrote unrelated test
  files once the `go` directive rose to 1.27.1, because new modernizers
  apply. Expect that in any module whose Go version rises with skgo's.
- `svelte-check` needs the extensions on `#lib` imports, but `vp build`
  does not, so a page can build and ship while its type check is red. Run
  the check in CI.
- The adapter's checks are strict and their messages say what to run: a
  `.remote.ts` edited by hand, a remote function nothing imports, or a
  missing `skgo.remotes.json` each fail the build with the fix in the text.
