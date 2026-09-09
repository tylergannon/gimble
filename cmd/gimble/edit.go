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
	"github.com/tylergannon/gimble/internal/editor"
	"github.com/tylergannon/gimble/internal/editor/server"
)

func newEditCommand() *cobra.Command {
	var addr, proxy string
	var noOpen bool
	command := &cobra.Command{
		Use:   "edit <pipeline>",
		Short: "Edit a pipeline in the browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runEditor(command, args[0], addr, proxy, noOpen)
		},
	}
	command.Flags().StringVar(&addr, "addr", "127.0.0.1:7331", "listen address (loopback only)")
	command.Flags().BoolVar(&noOpen, "no-open", false, "print the URL without opening a browser")
	command.Flags().StringVar(&proxy, "proxy", "", "URL of a `vp dev` server to forward pages to while developing the editor page; empty serves the embedded build")
	_ = command.Flags().MarkHidden("proxy")
	return command
}

func runEditor(command *cobra.Command, pipeline, addr, proxy string, noOpen bool) error {
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

	store, err := editor.Open(path, cliValidator())
	if err != nil {
		return err
	}
	dist, err := server.Dist()
	if err != nil {
		return err
	}
	listener, err := editor.Listen(addr)
	if err != nil {
		return err
	}
	// The origin is the URL printed below, and it is the only origin a save
	// is accepted from: skgo refuses a command whose Origin header differs,
	// which is kit's own cross-site rule. Browse the editor at this URL.
	origin := "http://" + listener.Addr().String()
	url := origin + "/"
	handler, err := server.NewHandler(store, dist, proxy, origin)
	if err != nil {
		_ = listener.Close()
		return err
	}

	ctx, stop := signal.NotifyContext(command.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	httpServer := &http.Server{
		Handler:           handler,
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
