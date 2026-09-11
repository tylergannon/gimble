package gimble

import (
	"context"
	"sync"
)

type group struct {
	scope  *scope
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
	err    error
}

// Group opens a concurrent scope named name. Go starts goroutines in it,
// each in a child scope of its own, and Wait joins them and ends it. The
// first error a goroutine returns cancels the group's ctx, interrupting
// the others' turns, and comes back from Wait.
func Group(ctx context.Context, name string) *group {
	parent, err := current(ctx)
	if err != nil {
		return &group{err: err}
	}
	g := &group{scope: parent.child(name)}
	g.ctx, g.cancel = context.WithCancel(context.WithValue(ctx, scopeKey{}, g.scope))
	g.scope.run.event(Event{Kind: "scope_began", Scope: g.scope.key, Name: name})
	return g
}

// Go runs fn in a goroutine, in a child scope of the group named name.
func (g *group) Go(name string, fn func(ctx context.Context) error) {
	if g.scope == nil {
		return
	}
	child := g.scope.child(name)
	g.wg.Go(func() {
		if err := child.do(g.ctx, fn); err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	})
}

// Wait joins the group's goroutines, ends its scope, and returns the first
// error one of them returned.
func (g *group) Wait() error {
	if g.scope == nil {
		return g.err
	}
	g.wg.Wait()
	g.scope.run.event(Event{Kind: "scope_ended", Scope: g.scope.key, Name: g.scope.key, Error: errString(g.err)})
	g.scope.end(g.cancel)
	return g.err
}
