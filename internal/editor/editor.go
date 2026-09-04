// Package editor serves the graph editor: the embedded web bundle plus a
// small JSON API that reads, lints, writes, and watches one pipeline file.
package editor

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tylergannon/tractor/engine"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/lint"
)

//go:embed all:dist
var bundle embed.FS

const (
	defaultPollInterval      = 500 * time.Millisecond
	defaultKeepaliveInterval = 15 * time.Second
	maxBodyBytes             = 64 << 20
)

// Dist returns the embedded editor bundle rooted at its index.html.
func Dist() fs.FS {
	sub, err := fs.Sub(bundle, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// Document is the wire shape shared by GET and PUT /api/doc.
type Document struct {
	Path        string                   `json:"path"`
	YAML        string                   `json:"yaml"`
	Layout      json.RawMessage          `json:"layout"`
	Version     string                   `json:"version"`
	Diagnostics []lint.Diagnostic        `json:"diagnostics"`
	ParseError  string                   `json:"parse_error"`
	Models      []engine.ModelResolution `json:"models"`
}

type putRequest struct {
	YAML    *string         `json:"yaml"`
	Layout  json.RawMessage `json:"layout"`
	Version string          `json:"version"`
}

// SidecarPath returns the layout sidecar beside a pipeline file:
// examples/x.yaml -> examples/x.layout.json.
func SidecarPath(pipeline string) string {
	return strings.TrimSuffix(pipeline, filepath.Ext(pipeline)) + ".layout.json"
}

// Server serves one pipeline file to the editor page.
type Server struct {
	path      string
	sidecar   string
	validator *lint.Validator
	dist      fs.FS
	static    http.Handler

	pollInterval      time.Duration
	keepaliveInterval time.Duration

	mu sync.Mutex // serializes disk reads and writes of the pipeline and sidecar
}

// New builds a server for the pipeline at path. validator may be nil, in
// which case only parse errors are reported. dist is the site to serve for
// non-API paths, normally Dist().
func New(pipeline string, validator *lint.Validator, dist fs.FS) (*Server, error) {
	absolute, err := filepath.Abs(pipeline)
	if err != nil {
		return nil, fmt.Errorf("resolve pipeline path: %w", err)
	}
	return &Server{
		path:              absolute,
		sidecar:           SidecarPath(absolute),
		validator:         validator,
		dist:              dist,
		static:            http.FileServerFS(dist),
		pollInterval:      defaultPollInterval,
		keepaliveInterval: defaultKeepaliveInterval,
	}, nil
}

// Path returns the absolute pipeline path the server edits.
func (s *Server) Path() string { return s.path }

// Listen opens a TCP listener that accepts loopback addresses only. An empty
// host means 127.0.0.1.
func Listen(addr string) (net.Listener, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("listen address %q: %w", addr, err)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if host != "localhost" {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return nil, fmt.Errorf("the editor listens on loopback only; refusing %q", addr)
		}
	}
	return net.Listen("tcp", net.JoinHostPort(host, port))
}

// ServeHTTP routes the API and falls back to the embedded site. Requests
// whose Host is not loopback are refused so a DNS-rebinding page cannot
// reach the API through a browser.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !loopbackHost(r.Host) {
		writeError(w, http.StatusForbidden, fmt.Sprintf("host %q is not loopback", r.Host))
		return
	}
	switch {
	case r.URL.Path == "/api/doc":
		switch r.Method {
		case http.MethodGet:
			s.handleGet(w)
		case http.MethodPut:
			s.handlePut(w, r)
		default:
			w.Header().Set("Allow", "GET, PUT")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	case r.URL.Path == "/api/events":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		s.handleEvents(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		writeError(w, http.StatusNotFound, "not found")
	default:
		s.serveStatic(w, r)
	}
}

// loopbackHost reports whether a Host header names 127.0.0.1, [::1], or
// localhost, with or without a port.
func loopbackHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return host == "127.0.0.1" || host == "::1" || strings.EqualFold(host, "localhost")
}

func (s *Server) handleGet(w http.ResponseWriter) {
	document, err := s.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, document)
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("read body: %v", err))
		return
	}
	var request putRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode body: %v", err))
		return
	}
	if request.YAML == nil {
		writeError(w, http.StatusBadRequest, `"yaml" is required`)
		return
	}
	layout := normalizeLayout(request.Layout)
	if layout != nil && !isJSONObject(layout) {
		writeError(w, http.StatusBadRequest, `"layout" must be an object or null`)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.loadLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if request.Version != current.Version {
		writeJSON(w, http.StatusConflict, current)
		return
	}
	if err := writeFileAtomic(s.path, []byte(*request.YAML)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if layout != nil {
		var indented bytes.Buffer
		if err := json.Indent(&indented, layout, "", "  "); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("layout: %v", err))
			return
		}
		indented.WriteByte('\n')
		if err := writeFileAtomic(s.sidecar, indented.Bytes()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	updated, err := s.loadLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	controller := http.NewResponseController(w)
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if err := controller.Flush(); err != nil {
		return
	}

	// The current version goes out first so a client can catch a change that
	// landed between its GET and this subscription, or during a reconnect.
	last := s.version()
	if err := writeChange(w, last); err != nil {
		return
	}
	if err := controller.Flush(); err != nil {
		return
	}
	poll := time.NewTicker(s.pollInterval)
	defer poll.Stop()
	keepalive := time.NewTicker(s.keepaliveInterval)
	defer keepalive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepalive.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
		case <-poll.C:
			version := s.version()
			if version == last {
				continue
			}
			last = version
			if err := writeChange(w, version); err != nil {
				return
			}
		}
		if err := controller.Flush(); err != nil {
			return
		}
	}
}

func writeChange(w io.Writer, version string) error {
	payload, _ := json.Marshal(struct {
		Version string `json:"version"`
	}{version})
	_, err := fmt.Fprintf(w, "event: change\ndata: %s\n\n", payload)
	return err
}

// serveStatic serves the bundle; unknown paths are rewritten to "/" so the
// single-page app boots and routes client-side.
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	clean := path.Clean("/" + r.URL.Path)
	if strings.HasPrefix(clean, "/_app/immutable/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if s.exists(clean) {
		s.static.ServeHTTP(w, r)
		return
	}
	rewritten := r.Clone(r.Context())
	rewritten.URL.Path = "/"
	rewritten.URL.RawPath = ""
	s.static.ServeHTTP(w, rewritten)
}

// exists reports whether clean names a file, or a directory with an
// index.html, inside the bundle.
func (s *Server) exists(clean string) bool {
	name := strings.TrimPrefix(clean, "/")
	if name == "" {
		name = "."
	}
	info, err := fs.Stat(s.dist, name)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	_, err = fs.Stat(s.dist, path.Join(name, "index.html"))
	return err == nil
}

// load reads the pipeline and sidecar, lints, and assembles the document.
func (s *Server) load() (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *Server) loadLocked() (*Document, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read pipeline: %w", err)
	}
	sidecar, err := readOptional(s.sidecar)
	if err != nil {
		return nil, fmt.Errorf("read layout: %w", err)
	}
	document := &Document{
		Path:        s.path,
		YAML:        string(raw),
		Version:     hashVersion(raw, sidecar),
		Diagnostics: []lint.Diagnostic{},
		Models:      []engine.ModelResolution{},
	}
	if json.Valid(sidecar) && isJSONObject(sidecar) {
		document.Layout = json.RawMessage(sidecar)
	}
	document.Diagnostics, document.ParseError, document.Models = s.inspect(raw)
	return document, nil
}

func (s *Server) inspect(raw []byte) ([]lint.Diagnostic, string, []engine.ModelResolution) {
	var (
		pipeline *graph.Graph
		err      error
	)
	if strings.EqualFold(filepath.Ext(s.path), ".json") {
		pipeline, err = graph.Parse(raw)
	} else {
		pipeline, err = graph.ParseYAML(raw)
	}
	if err != nil {
		return []lint.Diagnostic{}, err.Error(), []engine.ModelResolution{}
	}
	models, modelErr := engine.ResolveGraphModels(*pipeline, engine.SystemModelSelection{})
	if modelErr != nil {
		return []lint.Diagnostic{}, modelErr.Error(), []engine.ModelResolution{}
	}
	if s.validator == nil {
		return []lint.Diagnostic{}, "", models
	}
	diagnostics := s.validator.Validate(*pipeline)
	if diagnostics == nil {
		diagnostics = []lint.Diagnostic{}
	}
	return diagnostics, "", models
}

// version hashes the on-disk content without lint; missing files hash as empty.
func (s *Server) version() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, _ := readOptional(s.path)
	sidecar, _ := readOptional(s.sidecar)
	return hashVersion(raw, sidecar)
}

func hashVersion(yaml, sidecar []byte) string {
	digest := sha256.New()
	_, _ = fmt.Fprintf(digest, "%d:", len(yaml))
	digest.Write(yaml)
	_, _ = fmt.Fprintf(digest, "%d:", len(sidecar))
	digest.Write(sidecar)
	return hex.EncodeToString(digest.Sum(nil))
}

func readOptional(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

// writeFileAtomic writes data to a temp file in the same directory, then
// renames it over name, preserving the existing file mode.
func writeFileAtomic(name string, data []byte) error {
	mode := fs.FileMode(0o644)
	if info, err := os.Stat(name); err == nil {
		mode = info.Mode().Perm()
	}
	temp, err := os.CreateTemp(filepath.Dir(name), "."+filepath.Base(name)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempName := temp.Name()
	cleanup := func(err error) error {
		_ = temp.Close()
		_ = os.Remove(tempName)
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return cleanup(fmt.Errorf("write %s: %w", name, err))
	}
	if err := temp.Sync(); err != nil {
		return cleanup(fmt.Errorf("sync %s: %w", name, err))
	}
	if err := temp.Chmod(mode); err != nil {
		return cleanup(fmt.Errorf("chmod %s: %w", name, err))
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempName)
		return fmt.Errorf("close %s: %w", name, err)
	}
	if err := os.Rename(tempName, name); err != nil {
		_ = os.Remove(tempName)
		return fmt.Errorf("rename into %s: %w", name, err)
	}
	return nil
}

func normalizeLayout(layout json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(layout)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	return trimmed
}

func isJSONObject(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{message})
}
