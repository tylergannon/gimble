// Package web embeds the SvelteKit build the skgo adapter writes.
//
// A bare checkout embeds only build/.gitkeep so Go can compile before
// `just build` has produced the frontend. Starting that binary still fails
// loudly when the adapter manifest is absent.
package web

import "embed"

// Build holds the adapter output rooted at build/. The `all:` prefix is
// mandatory — without it Go silently omits `_app/`, because its default
// patterns skip names starting with `_` or `.`.
//
//go:embed all:build
var Build embed.FS
