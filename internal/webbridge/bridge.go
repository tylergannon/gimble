// Package webbridge connects the public runtime to its private embedded web
// application without making the assembly API public.
package webbridge

import (
	"errors"
	"net/http"
)

// Handler is the interface needed from the assembled web application.
type Handler = http.Handler

type Factory func(proxy, origin string) (Handler, string, error)

var factory Factory

// Register installs the embedded application factory during package init.
func Register(f Factory) {
	if factory != nil {
		panic("gimble: web application registered twice")
	}
	factory = f
}

// New assembles the registered application.
func New(proxy, origin string) (Handler, string, error) {
	if factory == nil {
		return nil, "", errors.New("gimble: web application is not registered")
	}
	return factory(proxy, origin)
}
