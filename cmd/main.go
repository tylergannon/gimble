// Command gimble runs Gimble's project runtime and web application.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/tylergannon/gimble"
)

func main() {
	port := flag.Int("port", 8080, "loopback TCP port for the web application")
	uds := flag.String("uds", "", "Unix-domain socket for the web application instead of TCP")
	noWeb := flag.Bool("no-web", false, "run without the web application")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var err error
	switch {
	case *noWeb:
		_, err = gimble.NewRuntime(ctx, ".gimble", gimble.WithNoWeb())
	case *uds != "":
		_, err = gimble.NewRuntime(ctx, ".gimble", gimble.WithUDS(*uds))
	default:
		_, err = gimble.NewRuntime(ctx, ".gimble", gimble.WithPort(*port))
	}
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
