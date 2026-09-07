// Package web embeds the SvelteKit build the skgo adapter writes.
//
// The tree does not compile until the frontend has been built at least once
// (`mise run build` in this directory); a missing frontend fails loudly here
// rather than producing a binary that serves nothing. The build is committed
// so `go install` needs no Node.
package web

import "embed"

// Build holds the adapter output rooted at build/. The `all:` prefix is
// mandatory: without it Go silently omits `_app/`, because its default
// patterns skip names starting with `_` or `.`.
//
//go:embed all:build
var Build embed.FS
