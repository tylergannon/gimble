# Gimble

Gimble is a Go library for writing agent workflows as ordinary Go. Its public
programming contract is the root package's Godoc and compiling examples:

```sh
go doc -all github.com/tylergannon/gimble
```

The web application is a SvelteKit app served by Go. One binary, one build
gesture; Node is a build-time dependency only.

## Build and run

This project uses Justfile for its build commands. It requires Node 24, pnpm 11, and just.

```sh
just build
./bin/gimble
```

Then open http://127.0.0.1:8080. The page's heading, the Go version below it, and the
greeting counter all come from `web/src/routes/hello.remote.go` — the
`hello.remote.ts` beside it is generated and every one of its bodies throws, so
anything that renders is proof that Go answered.

For development, build once and run these in separate terminals:

```sh
just dev-web
just dev-go
```

The public origin is set once in `Justfile`. Change it there and rebuild
before serving the application at a different origin.

## Where things are

| | |
| --- | --- |
| `web/` | an ordinary SvelteKit app: kit's tooling, kit's conventions |
| `@skgo/sveltekit-adapter` | kit's adapter, an ordinary devDependency paired with the skgo version `go.mod` requires |
| `web/src/routes/*.remote.go` | server logic, colocated with the routes that call it |
| `web/src/routes/**/server.go` | ordinary Go HTTP handlers for SvelteKit `+server.ts` routes |
| `generated/` | what `go generate ./...` writes; never edited by hand |
| `cmd/` | the binary |
| `server.go` | the one composition the binary and any test both use |

Write a remote function by adding a Go function to a `*.remote.go` file beside
the route that needs it and marking it with `skgo.Query` or `skgo.Command`, then
run `just build`. skgo writes the `.remote.ts` module kit compiles, the
TypeScript types its callers see, and the Go registration the server mounts.

Write a server route with an ordinary `net/http` handler in `server.go` beside
the route and mark it with `skgo.GET`, `skgo.POST` or another HTTP method. The
same build gesture writes the throwing `+server.ts` stub Kit compiles and the Go
registration the server answers in both production and development.

## The origin

The app's origin is fixed when the frontend is built, and the server checks it
on every non-GET remote call — that is kit's own rule, and a command is a POST.
It is written once, as `ORIGIN` in `Justfile`, and both halves read it from
there.

To serve the app somewhere else, change `ORIGIN` and run `just build`
again. Changing only one half makes every command fail with 403; the server
says so in its log and in the response body if it happens.

## Development

`just build` once, then two terminals:

```sh
just dev-web   # vite, on 127.0.0.1:5173
just dev-go    # the Go server, rendering from vite's modules
```

`#lib` is a Node subpath import (`package.json` → `imports`), and TypeScript
resolves those without probing for extensions: write `#lib/model.ts`, not
`#lib/model`. Vite accepts both, so only the type check would tell you.

Go answers loads, remote functions and server routes in both modes. In
production the binary renders pages from the embedded frontend build. In
development it still renders each document in Go, using modules transformed by
Vite; modules, styles, static assets and HMR continue through to Vite.

## Browser acceptance

The generated Playwright-BDD suite in `e2e/` is the starter application's
executable contract. Start either the production binary or both development
processes, then run:

```sh
just e2e
```

The same scenarios run in both modes. They prove that Go rendered the initial
document, a greeting visibly refreshes without reloading, and client navigation
and a direct deep link both reach the About route. Each outcome leaves a
screenshot under `e2e/screenshots/` so a successful run can be inspected.
