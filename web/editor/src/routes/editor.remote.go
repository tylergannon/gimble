// Package routes holds the server side of the editor page, written in Go and
// colocated with the page that calls it.
//
// `skgo generate` reads the declarations at the bottom of this file and writes
// editor.remote.ts beside it, a module whose every body throws. Anything the
// page renders is therefore proof that Go answered.
package routes

import (
	"context"

	"github.com/tylergannon/skgo"
	"github.com/tylergannon/tractor/internal/editor"
)

// store is the file the server was started on. The server's handle hook puts
// it on every request; see internal/editor/server.
func store(ctx context.Context) (*editor.Store, error) {
	s, ok := skgo.LocalOf[*editor.Store](ctx)
	if !ok || s == nil {
		return nil, skgo.Errorf(500, "the editor has no pipeline open")
	}
	return s, nil
}

// getDoc reads the pipeline, its layout sidecar, and what lint makes of it.
func getDoc(ctx context.Context) (editor.Document, error) {
	s, err := store(ctx)
	if err != nil {
		return editor.Document{}, err
	}
	return s.Load()
}

// saveDoc writes the pipeline back, provided it still has the version the
// page last read. A stale save writes nothing and returns what is on disk.
func saveDoc(ctx context.Context, request editor.SaveRequest) (editor.SaveResult, error) {
	s, err := store(ctx)
	if err != nil {
		return editor.SaveResult{}, err
	}
	return s.Save(request)
}

// watchDoc announces the current version at once and then every change to
// the files on disk, until the page disconnects.
func watchDoc(ctx context.Context, yield func(editor.Change) error) error {
	s, err := store(ctx)
	if err != nil {
		return err
	}
	return s.Watch(ctx, yield)
}

var (
	_ = skgo.Query(getDoc)
	_ = skgo.Command(saveDoc)
	_ = skgo.LiveQuery(watchDoc)
)
