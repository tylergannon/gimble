package gimble

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tylergannon/gimble/internal/webbridge"
	_ "github.com/tylergannon/gimble/internal/webembed"
)

func runDir(ctx context.Context) string {
	s, _ := ctx.Value(scopeKey{}).(*scope)
	if s == nil || s.run == nil {
		return ""
	}
	return s.run.dir
}

// Runtime owns a project's runs and its web application. Create one with
// NewRuntime, then use Run to execute workflows. It remains active until the
// context passed to NewRuntime is cancelled.
type Runtime struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	dir    string
	done   chan struct{}
	web    string
}

type runtimeOption func(*runtimeConfig) error

type runtimeConfig struct {
	listener string
	port     int
	uds      string
	explicit string
}

// WithPort serves the web application on the given loopback TCP port. It is
// mutually exclusive with WithUDS and WithNoWeb. Without a listener option,
// NewRuntime uses port 8080.
func WithPort(port int) runtimeOption {
	return func(c *runtimeConfig) error {
		if port < 0 || port > 65535 {
			return fmt.Errorf("gimble: invalid web port %d", port)
		}
		if err := c.selectListener("port"); err != nil {
			return err
		}
		c.listener, c.port = "tcp", port
		return nil
	}
}

// WithUDS serves the web application on a Unix-domain socket at path. It is
// mutually exclusive with WithPort and WithNoWeb. NewRuntime fails rather than
// replacing an existing filesystem entry at path.
func WithUDS(path string) runtimeOption {
	return func(c *runtimeConfig) error {
		if strings.TrimSpace(path) == "" {
			return errors.New("gimble: UDS path must not be blank")
		}
		if err := c.selectListener("UDS"); err != nil {
			return err
		}
		c.listener, c.uds = "unix", path
		return nil
	}
}

// WithNoWeb prevents the runtime from opening a web listener. Runs and their
// durable records still work normally. It is mutually exclusive with WithPort
// and WithUDS.
func WithNoWeb() runtimeOption {
	return func(c *runtimeConfig) error {
		if err := c.selectListener("no web"); err != nil {
			return err
		}
		c.listener = "none"
		return nil
	}
}

func (c *runtimeConfig) selectListener(name string) error {
	if c.explicit != "" {
		return fmt.Errorf("gimble: runtime listener options %q and %q conflict", c.explicit, name)
	}
	c.explicit = name
	return nil
}

// NewRuntime creates the runtime for projectDir and starts its web application
// before returning. The web application listens on loopback port 8080 unless
// an option selects another port, a Unix-domain socket, or no web listener.
// Cancelling ctx stops the web application and every run owned by the runtime.
func NewRuntime(ctx context.Context, projectDir string, opts ...runtimeOption) (*Runtime, error) {
	if ctx == nil {
		return nil, errors.New("gimble: runtime context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("gimble: project directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("gimble: project directory: %w", err)
	}
	config := runtimeConfig{listener: "tcp", port: 8080}
	for _, option := range opts {
		if option == nil {
			return nil, errors.New("gimble: nil runtime option")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	runtimeCtx, cancel := context.WithCancelCause(ctx)
	runtime := &Runtime{ctx: runtimeCtx, cancel: cancel, dir: dir, done: make(chan struct{})}
	if config.listener != "none" {
		if err := runtime.startWeb(config); err != nil {
			cancel(err)
			return nil, err
		}
	} else {
		context.AfterFunc(runtimeCtx, func() { close(runtime.done) })
	}
	return runtime, nil
}

func (r *Runtime) startWeb(config runtimeConfig) error {
	network, address := config.listener, config.uds
	if network == "tcp" {
		address = net.JoinHostPort("127.0.0.1", strconv.Itoa(config.port))
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("gimble: listen on %s: %w", address, err)
	}
	origin := ""
	if network == "tcp" {
		origin = "http://" + listener.Addr().String()
	}
	if configured := os.Getenv("GIMBLE_WEB_ORIGIN"); configured != "" {
		origin = configured
	}
	handler, mode, err := webbridge.New(os.Getenv("GIMBLE_WEB_PROXY"), origin)
	if err != nil {
		_ = listener.Close()
		return fmt.Errorf("gimble: assemble web application: %w", err)
	}
	server := &http.Server{
		Handler: handler,
		BaseContext: func(net.Listener) context.Context {
			return r.ctx
		},
	}
	r.web = listener.Addr().String()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.cancel(fmt.Errorf("gimble: serve web application: %w", err))
		}
	}()
	context.AfterFunc(r.ctx, func() {
		_ = server.Close()
		if network == "unix" {
			_ = os.Remove(address)
		}
		close(r.done)
	})
	log.Printf("gimble: web application listening on %s (%s)", listener.Addr(), mode)
	return nil
}

type run struct {
	dir      string // <project>/runs/<id>
	writer   *eventWriter
	project  *eventWriter
	mu       sync.Mutex
	sessions map[string]*eventWriter
}

// Run starts one workflow run and blocks until body returns. Its context ends
// when either ctx or the runtime context ends. The run is the root scope: its
// sessions close and its durable record completes before Run returns.
func (runtime *Runtime) Run(ctx context.Context, name string, body func(ctx context.Context) error) error {
	if runtime == nil {
		return errors.New("gimble: nil runtime")
	}
	if ctx == nil {
		return errors.New("gimble: run context is nil")
	}
	if body == nil {
		return errors.New("gimble: run body is nil")
	}
	runCtx, cancel := context.WithCancelCause(runtime.ctx)
	stop := context.AfterFunc(ctx, func() { cancel(context.Cause(ctx)) })
	defer stop()
	defer cancel(nil)
	err := runtime.run(runCtx, name, body)
	if err == nil && context.Cause(runCtx) != nil {
		return context.Cause(runCtx)
	}
	return err
}

func (runtime *Runtime) run(ctx context.Context, name string, body func(ctx context.Context) error) error {
	project := runtime.dir
	id := time.Now().Format("20060102-150405") + "." + name
	dir := filepath.Join(project, "runs", id)
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	w, err := newEventWriter(filepath.Join(dir, "run.jsonl"))
	if err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	pw, _ := newEventWriter(filepath.Join(project, "project.jsonl"))
	r := &run{dir: dir, writer: w, project: pw, sessions: make(map[string]*eventWriter)}
	if pw != nil {
		_ = pw.write(Event{Kind: "run_started", Name: name})
	}
	r.event(Event{Kind: "run_started", Name: name})
	logf("run %s started in %s", id, dir)
	err = (&scope{run: r}).do(ctx, body)
	if ctx.Err() != nil {
		r.event(Event{Kind: "run_cancelled", Error: ctx.Err().Error()})
		if pw != nil {
			_ = pw.write(Event{Kind: "run_cancelled", Name: name, Error: ctx.Err().Error()})
		}
	}
	r.event(Event{Kind: "run_ended", Name: name, Error: errString(err)})
	if pw != nil {
		_ = pw.write(Event{Kind: "run_ended", Name: name, Error: errString(err)})
	}
	r.event(Event{Kind: "complete"})
	_ = w.close()
	if pw != nil {
		_ = pw.close()
	}
	logf("run %s ended: %v", id, orNone(err))
	return err
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// logf traces what the runtime does on stderr until the run log exists.
func logf(format string, args ...any) {
	log.Printf("gimble: "+format, args...)
}

func orNone(err error) any {
	if err == nil {
		return "ok"
	}
	return err
}
