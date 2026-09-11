// Command gimble serves the SvelteKit app in web/ from Go.
//
// With --proxy it renders pages in Go from modules transformed by a running
// `vp dev` server; without it, it serves the build embedded at compile time and
// no Node process is involved at all. Either way loads, remote functions and
// server routes are answered by Go before a page is rendered.
package main

import (
	"flag"
	"io/fs"
	"log"
	"net/http"
	"net/url"

	"github.com/tylergannon/gimble/web"
)

// origin is the app's origin: the URL a browser reaches it at. It is baked in
// at link time by `just build`, from the single ORIGIN in Justfile that
// the frontend build also reads, so the two halves cannot disagree by
// accident. The value here is only the fallback for `go run ./cmd`.
//
// It is not derived from --listen. Behind a reverse proxy the two are
// legitimately different, and guessing would replace a clear refusal with a
// silent one.
var origin = "http://127.0.0.1:8080"

func main() {
	listen := flag.String("listen", hostPort(origin), "address to listen on")
	trusted := flag.String("origin", origin, "the origin browsers reach this app at; non-GET remote calls from anywhere else are refused")
	proxy := flag.String("proxy", "", "URL of a `vp dev` server supplying development modules and assets; empty serves the embedded build")
	flag.Parse()

	dist, err := fs.Sub(web.Build, "build")
	if err != nil {
		log.Fatalf("gimble: %v", err)
	}
	handler, mode, err := web.NewHandler(dist, *proxy, *trusted)
	if err != nil {
		log.Fatalf("gimble: %v", err)
	}

	// In dev the client is served by vite and kit itself skips the origin
	// check, so saying which origin is trusted would be a lie about prod.
	if mode == "prod" {
		log.Printf("gimble: listening on http://%s; commands are accepted from %s", *listen, *trusted)
	} else {
		log.Printf("gimble: listening on http://%s in dev mode, rendering from %s; the origin check is off, as it is in kit", *listen, *proxy)
	}
	if err := http.ListenAndServe(*listen, handler); err != nil {
		log.Fatalf("gimble: %v", err)
	}
}

// hostPort is the listen address an origin implies.
func hostPort(origin string) string {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return "127.0.0.1:8080"
	}
	return u.Host
}
