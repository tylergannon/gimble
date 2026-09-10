package program

import "context"

// Scope models a child scope in the API; this stub does not create or enforce one.
func Scope(ctx context.Context, _ string) (context.Context, error) { return ctx, nil }
