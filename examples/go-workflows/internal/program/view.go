package program

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// materializeView gives native tools meaningful filenames for a revision.
// Only directories and symlinks are new; inherited JSON bytes are not copied.
// The caller holds the store lock and publishes the view after it is complete.
// Writing directly through a symlink is not copy-on-write: callers must use
// SetContext to create an override without changing earlier scope revisions.
func (store *contextStore) materializeView(ctx context.Context, index string, keys []string) (view string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	view, err = os.MkdirTemp(store.dir, fmt.Sprintf("view-%06d-", store.revision))
	if err != nil {
		return "", err
	}
	created := view
	defer func() {
		if err != nil {
			_ = os.RemoveAll(created)
		}
	}()
	if err = os.Mkdir(filepath.Join(view, "values"), 0o755); err != nil {
		return "", err
	}
	for _, key := range keys {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		if err = os.Symlink(store.values[key].path, filepath.Join(view, "values", key+".json")); err != nil {
			return "", err
		}
	}
	if err = os.Symlink(index, filepath.Join(view, "index.json")); err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	return view, nil
}
