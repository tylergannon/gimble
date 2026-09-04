package editor

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/tylergannon/tractor/lint"
)

const unreachableYAML = `name: t
start: a
nodes:
  - id: a
    type: command
    command: "true"
    edges:
      success: success
  - id: b
    type: command
    command: "true"
    edges:
      success: success
`

func newTestServer(t *testing.T, source string) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	pipeline := filepath.Join(dir, filepath.Base(source))
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pipeline, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	site := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<h1>editor</h1>")}}
	server, err := New(pipeline, lint.New(lint.Options{}), site)
	if err != nil {
		t.Fatal(err)
	}
	server.pollInterval = 10 * time.Millisecond
	server.keepaliveInterval = 20 * time.Millisecond
	return server, pipeline
}

// newRequest builds a request whose Host is loopback, as a browser on the
// printed URL would send.
func newRequest(method, target string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	request.Host = "127.0.0.1:7331"
	return request
}

func getDocument(t *testing.T, server http.Handler) Document {
	t.Helper()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodGet, "/api/doc", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /api/doc = %d: %s", recorder.Code, recorder.Body)
	}
	var document Document
	if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode: %v\n%s", err, recorder.Body)
	}
	return document
}

func putDocument(t *testing.T, server http.Handler, body any) (int, Document) {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodPut, "/api/doc", bytes.NewReader(payload)))
	var document Document
	if recorder.Code == http.StatusOK || recorder.Code == http.StatusConflict {
		if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
			t.Fatalf("decode %d: %v\n%s", recorder.Code, err, recorder.Body)
		}
	}
	return recorder.Code, document
}

func TestGetReturnsFileAndDiagnostics(t *testing.T) {
	server, pipeline := newTestServer(t, "../../examples/loops/bake-off.yaml")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodGet, "/api/doc", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type %q", got)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"path", "yaml", "layout", "version", "diagnostics", "parse_error"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("response lacks %q:\n%s", key, recorder.Body)
		}
	}
	if string(raw["layout"]) != "null" {
		t.Fatalf("layout = %s, want null", raw["layout"])
	}
	if string(raw["diagnostics"]) == "null" {
		t.Fatalf("diagnostics must be an array, got null")
	}

	document := getDocument(t, server)
	want, _ := os.ReadFile(pipeline)
	if document.YAML != string(want) {
		t.Fatalf("yaml mismatch:\n%s", document.YAML)
	}
	if document.Path != pipeline {
		t.Fatalf("path = %q, want %q", document.Path, pipeline)
	}
	if document.ParseError != "" {
		t.Fatalf("parse_error = %q", document.ParseError)
	}
	if document.Version == "" {
		t.Fatal("version is empty")
	}
	if lint.HasErrors(document.Diagnostics) {
		t.Fatalf("bake-off has lint errors: %+v", document.Diagnostics)
	}
}

func TestGetReportsLintDiagnosticsAndParseErrors(t *testing.T) {
	server, pipeline := newTestServer(t, "../../examples/loops/bake-off.yaml")
	if err := os.WriteFile(pipeline, []byte(unreachableYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	document := getDocument(t, server)
	if document.ParseError != "" {
		t.Fatalf("parse_error = %q", document.ParseError)
	}
	found := false
	for _, diagnostic := range document.Diagnostics {
		if diagnostic.Rule == "reachability" && diagnostic.NodeID == "b" && diagnostic.Severity == lint.SeverityError {
			found = true
		}
	}
	if !found {
		t.Fatalf("no reachability error for b: %+v", document.Diagnostics)
	}

	if err := os.WriteFile(pipeline, []byte("name: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	document = getDocument(t, server)
	if document.ParseError == "" {
		t.Fatal("parse_error is empty for broken YAML")
	}
	if len(document.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v, want none on parse failure", document.Diagnostics)
	}
}

func TestPutWithStaleVersionConflictsWithoutWriting(t *testing.T) {
	server, pipeline := newTestServer(t, "../../examples/loops/bake-off.yaml")
	before, _ := os.ReadFile(pipeline)
	current := getDocument(t, server)

	status, document := putDocument(t, server, map[string]any{
		"yaml":    "name: replaced\n",
		"layout":  nil,
		"version": "stale",
	})
	if status != http.StatusConflict {
		t.Fatalf("status = %d, want 409", status)
	}
	if document.Version != current.Version || document.YAML != string(before) {
		t.Fatalf("409 body is not the current document: %+v", document)
	}
	after, _ := os.ReadFile(pipeline)
	if !bytes.Equal(before, after) {
		t.Fatal("stale PUT modified the file")
	}
	if _, err := os.Stat(SidecarPath(pipeline)); !os.IsNotExist(err) {
		t.Fatalf("stale PUT touched the sidecar: %v", err)
	}
}

func TestPutWritesFileAndSidecar(t *testing.T) {
	server, pipeline := newTestServer(t, "../../examples/loops/bake-off.yaml")
	current := getDocument(t, server)

	updated := strings.Replace(current.YAML, "start: attempts", "start: attempts # edited", 1)
	layout := map[string]any{"nodes": map[string]any{"attempts": map[string]int{"x": 10, "y": 20}}}
	status, document := putDocument(t, server, map[string]any{
		"yaml":    updated,
		"layout":  layout,
		"version": current.Version,
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	onDisk, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != updated {
		t.Fatalf("file not rewritten byte-for-byte:\n%s", onDisk)
	}
	if document.YAML != updated || document.Version == current.Version || document.ParseError != "" {
		t.Fatalf("response document stale: %+v", document)
	}
	sidecar, err := os.ReadFile(SidecarPath(pipeline))
	if err != nil {
		t.Fatalf("sidecar not written: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(sidecar, &got); err != nil {
		t.Fatalf("sidecar is not JSON: %v\n%s", err, sidecar)
	}
	var fromResponse map[string]any
	if err := json.Unmarshal(document.Layout, &fromResponse); err != nil {
		t.Fatalf("response layout: %v", err)
	}
	if fromResponse["nodes"] == nil || got["nodes"] == nil {
		t.Fatalf("layout lost: disk %s response %s", sidecar, document.Layout)
	}
	if entries, _ := filepath.Glob(filepath.Join(filepath.Dir(pipeline), ".*.tmp")); len(entries) != 0 {
		t.Fatalf("temp files left behind: %v", entries)
	}

	// A second PUT with a null layout leaves the sidecar alone.
	status, next := putDocument(t, server, map[string]any{
		"yaml":    updated + "# more\n",
		"layout":  nil,
		"version": document.Version,
	})
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	again, _ := os.ReadFile(SidecarPath(pipeline))
	if !bytes.Equal(again, sidecar) {
		t.Fatal("null layout rewrote the sidecar")
	}
	if string(next.Layout) == "null" {
		t.Fatal("response dropped the on-disk layout")
	}
}

func TestPutRejectsBadBodies(t *testing.T) {
	server, _ := newTestServer(t, "../../examples/loops/bake-off.yaml")
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodPut, "/api/doc", strings.NewReader("{")))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid JSON status = %d", recorder.Code)
	}
	status, _ := putDocument(t, server, map[string]any{"version": "x"})
	if status != http.StatusBadRequest {
		t.Fatalf("missing yaml status = %d", status)
	}
	status, _ = putDocument(t, server, map[string]any{"yaml": "", "layout": []int{1}, "version": "x"})
	if status != http.StatusBadRequest {
		t.Fatalf("array layout status = %d", status)
	}
}

func TestEventsEmitChangeWhenFileChangesOnDisk(t *testing.T) {
	server, pipeline := newTestServer(t, "../../examples/loops/bake-off.yaml")
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, httpServer.URL+"/api/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if got := response.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("content type %q", got)
	}

	reader := bufio.NewReader(response.Body)
	readChange := func() string {
		t.Helper()
		var event, data string
		for event == "" || data == "" {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case strings.HasPrefix(line, "event: "):
				event = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
			case strings.HasPrefix(line, "data: "):
				data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			}
		}
		if event != "change" {
			t.Fatalf("event = %q", event)
		}
		if blank, err := reader.ReadString('\n'); err != nil || blank != "\n" {
			t.Fatalf("event terminator %q, %v", blank, err)
		}
		var payload struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			t.Fatalf("data %q: %v", data, err)
		}
		return payload.Version
	}

	// The current version is announced on connect.
	initial := getDocument(t, server).Version
	if got := readChange(); got != initial {
		t.Fatalf("initial event version %q, want %q", got, initial)
	}
	// The keepalive comment arrives before any change.
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(line, ":") {
		t.Fatalf("line after the initial event %q is not a comment", line)
	}

	if err := os.WriteFile(pipeline, []byte(unreachableYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := readChange(), getDocument(t, server).Version; got != want || got == initial {
		t.Fatalf("event version %q, want %q (initial %q)", got, want, initial)
	}
}

func TestRejectsForeignHost(t *testing.T) {
	server, _ := newTestServer(t, "../../examples/loops/bake-off.yaml")
	for _, target := range []string{"/api/doc", "/api/events", "/", "/api/nothing"} {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		request.Host = "evil.example:7331"
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("%s with foreign host = %d, want 403", target, recorder.Code)
		}
	}
	for _, host := range []string{"127.0.0.1", "127.0.0.1:7331", "localhost", "localhost:7331", "LOCALHOST:80"} {
		request := httptest.NewRequest(http.MethodGet, "/api/doc", nil)
		request.Host = host
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("host %q = %d, want 200", host, recorder.Code)
		}
	}
	for _, host := range []string{"127.0.0.1.evil.example", "127.0.0.2", "[::1]:7331", ""} {
		request := httptest.NewRequest(http.MethodGet, "/api/doc", nil)
		request.Host = host
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("host %q = %d, want 403", host, recorder.Code)
		}
	}
}

func TestUnknownPathServesIndex(t *testing.T) {
	server, _ := newTestServer(t, "../../examples/loops/bake-off.yaml")
	for _, target := range []string{"/", "/nodes/judge", "/deep/route?x=1"} {
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, newRequest(http.MethodGet, target, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "<h1>editor</h1>") {
			t.Fatalf("%s: %d %q", target, recorder.Code, recorder.Body.String())
		}
		if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("%s: Cache-Control %q", target, got)
		}
	}
	server.dist = fstest.MapFS{
		"index.html":                 &fstest.MapFile{Data: []byte("<h1>editor</h1>")},
		"_app/immutable/chunks/a.js": &fstest.MapFile{Data: []byte("export {}")},
		"_app/version.json":          &fstest.MapFile{Data: []byte("{}")},
	}
	server.static = http.FileServerFS(server.dist)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodGet, "/_app/immutable/chunks/a.js", nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("immutable asset: %d %q", recorder.Code, recorder.Header().Get("Cache-Control"))
	}
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodGet, "/_app/version.json", nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("version.json: %d %q", recorder.Code, recorder.Header().Get("Cache-Control"))
	}
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, newRequest(http.MethodGet, "/api/nothing", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("/api/nothing = %d", recorder.Code)
	}
}

func TestEmbeddedDistHasIndex(t *testing.T) {
	recorder := httptest.NewRecorder()
	http.FileServerFS(Dist()).ServeHTTP(recorder, newRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "<html") {
		t.Fatalf("embedded index: %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestListenRefusesNonLoopback(t *testing.T) {
	if _, err := Listen("0.0.0.0:0"); err == nil {
		t.Fatal("0.0.0.0 accepted")
	}
	listener, err := Listen(":0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	if !strings.HasPrefix(listener.Addr().String(), "127.0.0.1:") {
		t.Fatalf("addr = %s", listener.Addr())
	}
}
