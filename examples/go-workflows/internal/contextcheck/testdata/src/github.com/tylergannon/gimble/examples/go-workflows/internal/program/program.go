package program

import "context"

func NewContext(context.Context, string, any) (context.Context, error) { return nil, nil }
func Scope(context.Context, string) (context.Context, error)           { return nil, nil }
func SetContext(context.Context, string, any) error                    { return nil }
