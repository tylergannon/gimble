// Command gimble runs Gimble's project runtime and web application.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/tylergannon/gimble/web"
)

func main() {
	port := flag.Int("port", 8080, "loopback TCP port for the web application")
	uds := flag.String("uds", "", "Unix-domain socket for the web application instead of TCP")
	noWeb := flag.Bool("no-web", false, "run without the web application")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var options []web.Option
	switch {
	case *noWeb:
		options = append(options, web.WithNoWeb())
	case *uds != "":
		options = append(options, web.WithUDS(*uds))
	default:
		options = append(options, web.WithPort(*port))
	}
	if _, err := web.NewRuntime(ctx, ".gimble", options...); err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
