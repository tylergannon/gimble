// Package routes holds the Go implementation of the editor page's server load.
package routes

import (
	"context"

	"github.com/tylergannon/gimble/internal/editor"
	"github.com/tylergannon/skgo"
)

// pageLoad puts the pipeline itself in the initial SSR document. The browser
// hydrates this value and only uses the query for later on-disk changes.
func pageLoad(ctx context.Context) (editor.Document, error) {
	s, err := store(ctx)
	if err != nil {
		return editor.Document{}, err
	}
	return s.Load()
}

var _ = skgo.Load(pageLoad)
