// Command gimble runs Gimble's project runtime and web application.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/tylergannon/gimble/web"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		_, _ = fmt.Fprintln(os.Stderr, "gimble:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer, getenv func(string) string) error {
	if len(args) > 0 && args[0] == "run-prompt" {
		return runPrompt(args[1:], stdout, stderr, getenv)
	}
	return runServer(args, stderr)
}

func runServer(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("gimble", flag.ContinueOnError)
	flags.SetOutput(stderr)
	port := flags.Int("port", 8080, "loopback TCP port for the web application")
	uds := flags.String("uds", "", "Unix-domain socket for the web application instead of TCP")
	noWeb := flags.Bool("no-web", false, "run without the web application")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

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
		return err
	}
	<-ctx.Done()
	return nil
}
