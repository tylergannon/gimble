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

// Group opens an errgroup-shaped concurrent scope named name. Call its Go
// method for each child, then return its Wait method's result:
//
//	group := gimble.Group(ctx, "candidates")
//	group.Go("candidate", first)
//	group.Go("candidate", second)
//	return group.Wait()
//
// Each child receives its own named scope. The first error cancels the group,
// interrupting the other children's turns; Wait joins every child, ends the
// group scope, and returns that first error.
// The caller must Wait on every exit path. To stop children early, cancel
// their parent context before waiting; cancellation alone does not join them.
func Group(ctx context.Context, name string) *group {
	parent, err := current(ctx)
	if err != nil {
		return &group{err: err}
	}
	g := &group{scope: parent.child(name)}
	g.ctx, g.cancel = context.WithCancel(context.WithValue(ctx, scopeKey{}, g.scope))
	g.scope.run.event(g.scope.key, "", "", ScopeBegan{Name: name})
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
// error one of them returned. Call Wait once, after submitting all children.
// Child results may be read after Wait returns.
func (g *group) Wait() error {
	if g.scope == nil {
		return g.err
	}
	g.wg.Wait()
	g.scope.run.event(g.scope.key, "", "", ScopeEnded{Error: errString(g.err)})
	g.scope.end(g.cancel)
	return g.err
}
