package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tylergannon/gimble/internal/editor"
	"github.com/tylergannon/gimble/internal/editor/generated"
	"github.com/tylergannon/gimble/lint"
	"github.com/tylergannon/polytype/devalue"
	"github.com/tylergannon/skgo"
)

// origin is the URL the handler under test is configured with. Nothing
// listens on it: every request below is served in-process.
const origin = "http://127.0.0.1:7331"

// newHandler is the production stack over the embedded build, the same
// function `gimble edit` calls, on a copy of one example pipeline.
func newHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	source, err := os.ReadFile("../../../examples/loops/bake-off.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pipeline := filepath.Join(t.TempDir(), "bake-off.yaml")
	if err := os.WriteFile(pipeline, source, 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := editor.Open(pipeline, lint.New(lint.Options{}))
	if err != nil {
		t.Fatal(err)
	}
	dist, err := Dist()
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(store, dist, "", origin)
	if err != nil {
		t.Fatalf("assembling the editor stack: %v", err)
	}
	return handler, pipeline
}

// remote returns the registered function with the given export name, which
// is how a test learns the `<hash>/<name>` id the built page calls.
func remote(t *testing.T, name string) *skgo.Remote {
	t.Helper()
	for _, r := range generated.Remotes() {
		if r.Name() == name {
			return r
		}
	}
	t.Fatalf("no remote function named %s", name)
	return nil
}

func remoteURL(t *testing.T, name string) string {
	return "/_app/remote/" + remote(t, name).ID()
}

// request builds a request as a browser on the printed URL would send it.
func request(method, target string, body string) *http.Request {
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	r := httptest.NewRequest(method, target, reader)
	r.Host = "127.0.0.1:7331"
	return r
}

func do(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// envelope decodes kit's remote-function response: a JSON wrapper around a
// devalue-encoded payload.
func envelope(t *testing.T, body []byte) (kind string, data *devalue.Object) {
	t.Helper()
	var resp struct {
		Type  string          `json:"type"`
		Data  string          `json:"data"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decoding envelope %q: %v", body, err)
	}
	if resp.Type != "result" {
		t.Fatalf("response type = %q: %s", resp.Type, body)
	}
	parsed, err := devalue.Parse(resp.Data, nil)
	if err != nil {
		t.Fatalf("parsing devalue payload %q: %v", resp.Data, err)
	}
	obj, ok := parsed.(*devalue.Object)
	if !ok {
		t.Fatalf("payload is %T, want an object", parsed)
	}
	return resp.Type, obj
}

func field(t *testing.T, o *devalue.Object, key string) any {
	t.Helper()
	v, ok := o.Get(key)
	if !ok {
		t.Fatalf("no %q in %v", key, o.Keys())
	}
	return v
}

func object(t *testing.T, v any) *devalue.Object {
	t.Helper()
	o, ok := v.(*devalue.Object)
	if !ok {
		t.Fatalf("value is %T, want an object", v)
	}
	return o
}

func TestThePipelineIsServerRenderedFromTheEmbeddedBuild(t *testing.T) {
	h, _ := newHandler(t)
	rec := do(h, request(http.MethodGet, "/", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "<title>Gimble editor</title>") {
		t.Fatalf("GET / is not the editor page:\n%s", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), ">bake-off</span>") ||
		!strings.Contains(rec.Body.String(), ">3 nodes</span>") {
		t.Fatalf("GET / did not render the pipeline before JavaScript ran:\n%s", rec.Body)
	}
}

func TestAnOrdinaryGoHandlerCanShareTheServer(t *testing.T) {
	app, _ := newHandler(t)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", app)

	health := do(mux, request(http.MethodGet, "/healthz", ""))
	if health.Code != http.StatusOK || health.Body.String() != "ok" {
		t.Fatalf("GET /healthz = %d %q", health.Code, health.Body.String())
	}
	page := do(mux, request(http.MethodGet, "/", ""))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), ">bake-off</span>") {
		t.Fatalf("GET / through the shared mux = %d:\n%s", page.Code, page.Body)
	}
}

func TestAnUnknownPathIsNotFound(t *testing.T) {
	h, _ := newHandler(t)
	if rec := do(h, request(http.MethodGet, "/api/doc", "")); rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/doc = %d, want 404: the old JSON API is gone", rec.Code)
	}
}

func TestAForeignHostIsRefused(t *testing.T) {
	h, _ := newHandler(t)
	r := request(http.MethodGet, "/", "")
	r.Host = "editor.example.com"
	if rec := do(h, r); rec.Code != http.StatusForbidden {
		t.Fatalf("GET / with Host example.com = %d, want 403", rec.Code)
	}
}

func TestGetDocAnswersOverTheWire(t *testing.T) {
	h, pipeline := newHandler(t)
	source, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	rec := do(h, request(http.MethodGet, remoteURL(t, "getDoc"), ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("getDoc = %d: %s", rec.Code, rec.Body)
	}
	_, data := envelope(t, rec.Body.Bytes())
	// Kit's client reads a query's value out of `q[<id>/<payload>].v`.
	node := object(t, field(t, object(t, field(t, data, "q")), remote(t, "getDoc").ID()+"/"))
	doc := object(t, field(t, node, "v"))
	if got := field(t, doc, "yaml"); got != string(source) {
		t.Fatalf("yaml over the wire = %q", got)
	}
	if got := field(t, doc, "path"); got != pipeline {
		t.Fatalf("path over the wire = %q", got)
	}
	if _, ok := field(t, doc, "diagnostics").([]any); !ok {
		t.Fatalf("diagnostics = %#v, want an array", field(t, doc, "diagnostics"))
	}
}

func saveBody(t *testing.T, yaml, version string) string {
	t.Helper()
	arg := devalue.NewObject(
		"yaml", yaml,
		"layout", []any{devalue.NewObject("id", "a", "x", 1.0, "y", 2.0, "w", 0.0, "h", 0.0)},
		"write_layout", true,
		"version", version,
	)
	// Kit's client sends the devalue form of the argument, base64url-encoded.
	encoded, err := devalue.Stringify(arg)
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString([]byte(encoded))
	body, err := json.Marshal(map[string]any{"payload": payload, "refreshes": []string{}})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestSaveDocRefusesACrossSiteCommand(t *testing.T) {
	h, _ := newHandler(t)
	r := request(http.MethodPost, remoteURL(t, "saveDoc"), saveBody(t, "name: x\n", "v"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "https://evil.example")
	if rec := do(h, r); rec.Code != http.StatusForbidden {
		t.Fatalf("a command from another origin = %d, want 403", rec.Code)
	}
	r = request(http.MethodPost, remoteURL(t, "saveDoc"), saveBody(t, "name: x\n", "v"))
	r.Header.Set("Content-Type", "application/json")
	if rec := do(h, r); rec.Code != http.StatusForbidden {
		t.Fatalf("a command with no Origin = %d, want 403", rec.Code)
	}
}

func TestSaveDocWritesTheFileAndRefusesAStaleVersion(t *testing.T) {
	h, pipeline := newHandler(t)
	rec := do(h, request(http.MethodGet, remoteURL(t, "getDoc"), ""))
	_, data := envelope(t, rec.Body.Bytes())
	current := object(t, field(t, data, "_"))
	version := field(t, current, "version").(string)

	updated := "name: saved\nstart: a\nnodes:\n  - id: a\n    type: command\n    command: \"true\"\n    edges:\n      success: success\n"
	r := request(http.MethodPost, remoteURL(t, "saveDoc"), saveBody(t, updated, version))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", origin)
	rec = do(h, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("saveDoc = %d: %s", rec.Code, rec.Body)
	}
	_, data = envelope(t, rec.Body.Bytes())
	result := object(t, field(t, data, "_"))
	if saved := field(t, result, "saved"); saved != true {
		t.Fatalf("saved = %#v", saved)
	}
	onDisk, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != updated {
		t.Fatalf("pipeline on disk:\n%s", onDisk)
	}
	if _, err := os.Stat(editor.SidecarPath(pipeline)); err != nil {
		t.Fatalf("sidecar not written: %v", err)
	}

	// The same version again is stale now; nothing is written.
	r = request(http.MethodPost, remoteURL(t, "saveDoc"), saveBody(t, "name: stale\n", version))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", origin)
	rec = do(h, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("stale saveDoc = %d: %s", rec.Code, rec.Body)
	}
	_, data = envelope(t, rec.Body.Bytes())
	result = object(t, field(t, data, "_"))
	if saved := field(t, result, "saved"); saved != false {
		t.Fatalf("stale save reported saved = %#v", saved)
	}
	if got := field(t, object(t, field(t, result, "document")), "yaml"); got != updated {
		t.Fatalf("the refusal carries %q, want what is on disk", got)
	}
	again, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != updated {
		t.Fatal("a stale save changed the file")
	}
}

func TestWatchDocStreamsTheVersion(t *testing.T) {
	h, pipeline := newHandler(t)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	r := request(http.MethodGet, remoteURL(t, "watchDoc"), "").WithContext(ctx)
	rec := do(h, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("watchDoc = %d: %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q", ct)
	}
	store, err := editor.Open(pipeline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body := rec.Body.String(); !strings.Contains(body, store.Version()) {
		t.Fatalf("the stream did not announce the version on disk:\n%s", body)
	}
}
