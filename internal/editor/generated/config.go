// Package generated holds the wiring skgo writes for the editor page: one
// function returning every remote function the page's routes declare, and
// the link tree that makes the Go files colocated with those routes
// importable. Regenerate it with `go generate ./...`.
//
// `go tool skgo` builds the generator from the module cache, so nothing here
// needs skgo to be a writable checkout.
package generated

//go:generate go tool skgo generate --web ../../../web/editor
