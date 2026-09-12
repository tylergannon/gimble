// Package routes holds the Go declarations for the guide.
package routes

import (
	"context"

	"github.com/tylergannon/skgo"
)

// guide identifies the product the page documents. Keeping this as a query
// proves the built-in guide is served through the same Go application as a
// future interactive run view.
func guide(context.Context) (string, error) {
	return "Gimble", nil
}

var _ = skgo.Query(guide)
