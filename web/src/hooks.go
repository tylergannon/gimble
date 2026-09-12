// Package hooks is the Go half of the app's universal hooks. It sits beside
// src/hooks.ts because kit's `transport` is one declaration with two halves:
// the browser's decode lives there, and the type it decodes lives here.
package hooks

import "github.com/tylergannon/skgo"

// RunSnapshot carries one run's complete observation across the SSR load
// boundary.
//
// The observation is a deliberately open native schema: its sessions,
// invocations and projection rows are keyed by ids the run invents as it
// goes, and its message rows are whatever the provider sent. skgo's load
// projection is polytype's static grammar, which has no map and no opaque
// value, so the snapshot cannot be spelled as a load result directly.
//
// It does not have to be. Kit's `transport` hook exists so an app can put a
// type on the wire that its own decode reconstitutes, and that is all this
// is: the encoded form is the snapshot's own JSON, and src/hooks.ts parses
// it back into the same object the JSON endpoint and the event stream
// already serve. Nothing downstream sees this struct -- the browser and the
// SSR render both receive the RunSnapshot itself -- so it is an encoding at
// one boundary and not a second state model.
type RunSnapshot struct {
	// JSON is the observation, encoded exactly as GET /api/runs/{runID}
	// encodes it.
	JSON string `json:"json"`
}

// The key is spelled identically here and in src/hooks.ts: it is the name
// the value travels under, and a client with no decoder for it cannot read
// the response at all.
var _ = skgo.Transported[RunSnapshot]("RunSnapshot")
