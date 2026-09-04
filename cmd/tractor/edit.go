package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tylergannon/tractor/internal/editor"
)

func newEditCommand() *cobra.Command {
	var addr string
	var noOpen bool
	command := &cobra.Command{
		Use:   "edit <pipeline>",
		Short: "Edit a pipeline in the browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runEditor(command, args[0], addr, noOpen)
		},
	}
	command.Flags().StringVar(&addr, "addr", "127.0.0.1:0", "listen address (loopback only)")
	command.Flags().BoolVar(&noOpen, "no-open", false, "print the URL without opening a browser")
	return command
}

func runEditor(command *cobra.Command, pipeline, addr string, noOpen bool) error {
	path, err := filepath.Abs(pipeline)
	if err != nil {
		return fmt.Errorf("resolve pipeline path: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("pipeline %q: %w", pipeline, err)
	}
	if info.IsDir() {
		return fmt.Errorf("pipeline %q is a directory", pipeline)
	}
	// The page always writes YAML back, so a JSON pipeline would be rewritten
	// as YAML on its first edit.
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
	default:
		return fmt.Errorf("pipeline %q must end in .yaml or .yml; the editor writes YAML", pipeline)
	}

	server, err := editor.New(path, cliValidator(), editor.Dist())
	if err != nil {
		return err
	}
	listener, err := editor.Listen(addr)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s/", listener.Addr())

	ctx, stop := signal.NotifyContext(command.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	httpServer := &http.Server{
		Handler:           server,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	served := make(chan error, 1)
	go func() { served <- httpServer.Serve(listener) }()

	if _, err := fmt.Fprintln(command.OutOrStdout(), url); err != nil {
		return err
	}
	if !noOpen {
		if err := openBrowser(url); err != nil {
			if _, err := fmt.Fprintf(command.ErrOrStderr(), "open browser: %v\n", err); err != nil {
				return err
			}
		}
	}

	select {
	case err := <-served:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_ = httpServer.Close()
	}
	<-served
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default:
		return fmt.Errorf("no browser opener for %s; visit %s", runtime.GOOS, url)
	}
}
