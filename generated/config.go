// Package generated holds the wiring skgo writes: one function returning every
// remote function the app declares, and the link tree that makes the Go files
// colocated with the routes importable. Regenerate it with `go generate ./...`.
//
// `go tool skgo` builds the generator from the module cache, so nothing here
// needs skgo to be a writable checkout.
package generated

//go:generate go tool skgo generate --web ../web
