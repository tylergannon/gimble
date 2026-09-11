// Package routes holds the server logic for the app's root routes, written in
// Go and colocated with the pages that use it.
//
// `skgo generate` reads the declarations at the bottom of this file and writes
// hello.remote.ts beside it — a module whose every body throws. Anything this
// page renders is therefore proof that Go answered.
package routes

import (
	"context"
	"runtime"
	"sync"

	"github.com/tylergannon/skgo"
)

// Status is what the home page shows. Its fields cross to TypeScript, so it
// is the type the generated `types.ts` beside this file declares.
type Status struct {
	// Name is the application's name.
	Name string `json:"name"`
	// GoVersion is the Go runtime that answered.
	GoVersion string `json:"goVersion"`
	// Greetings counts the greetings this process has served.
	Greetings int `json:"greetings"`
	// LastGreeting is who was greeted last.
	LastGreeting string `json:"lastGreeting"`
}

// The demo's whole state. A real app reaches for its own store here; the point
// of the mutex is only that a command changes something a query can see.
var (
	mu           sync.Mutex
	greetings    int
	lastGreeting string
)

// status answers the home page's query. Queries are GET requests.
func status(ctx context.Context) (Status, error) {
	mu.Lock()
	defer mu.Unlock()
	return Status{
		Name:         "gimble",
		GoVersion:    runtime.Version(),
		Greetings:    greetings,
		LastGreeting: lastGreeting,
	}, nil
}

// greet is a command, so it arrives as a POST — which is the request the app's
// origin has to be right for. If this one works, the origin is right.
func greet(ctx context.Context, name string) (Status, error) {
	if name == "" {
		return Status{}, skgo.Errorf(400, "Tell me who to greet")
	}
	mu.Lock()
	greetings++
	lastGreeting = name
	mu.Unlock()

	result, err := status(ctx)
	if err != nil {
		return Status{}, err
	}
	// `.updates(status())` is a request from the browser, not permission to
	// run arbitrary queries. Accept this starter's one no-argument status
	// instance so its refreshed value rides back on the command response.
	return result, skgo.RefreshRequestedNoArg(ctx, status)
}

var (
	_ = skgo.Query(status)
	_ = skgo.Command(greet)
)
