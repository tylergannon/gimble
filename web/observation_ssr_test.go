package web

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tylergannon/gimble/internal/observation"
)

// TestRunPageIsRenderedFromTheRunsObservation is the SSR boundary end to end:
// the Go load reads the registry the request carries, the observation crosses
// as the app's transported type, and the document the browser would receive
// with JavaScript disabled already holds it.
func TestRunPageIsRenderedFromTheRunsObservation(t *testing.T) {
	dist, err := fs.Sub(Build, "build")
	if err != nil {
		t.Fatal(err)
	}
	handler, mode, err := NewHandler(dist, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "prod" {
		t.Fatalf("this test renders the embedded build, but the handler is in %s mode", mode)
	}

	project := t.TempDir()
	registry := observation.NewRegistry(project)
	store := observation.Open(registry, "run-1", "demo", filepath.Join(project, "runs", "run-1"))
	store.Lifecycle(observation.Lifecycle{
		Placement: observation.Placement{Session: "writer"},
		Session:   &observation.SessionInfo{Name: "writer", Adapter: "fixture", Model: "test-model"},
		Record:    json.RawMessage(`{"event":{"kind":"SessionCreated","name":"writer"}}`),
	})
	at := observation.Placement{Session: "writer", Turn: "turn-1"}
	for _, event := range []json.RawMessage{
		json.RawMessage(`{"id":"evt1","type":"session.step.started","created":10,"data":{"sessionID":"ses_writer","assistantMessageID":"msg_writer","agent":"fixture","model":{"providerID":"fixture","id":"test-model"}}}`),
		json.RawMessage(`{"id":"evt2","type":"session.text.started","created":11,"data":{"sessionID":"ses_writer","assistantMessageID":"msg_writer","ordinal":0}}`),
		json.RawMessage(`{"id":"evt3","type":"session.text.ended","created":12,"data":{"sessionID":"ses_writer","assistantMessageID":"msg_writer","ordinal":0,"text":"hello from Gimble"}}`),
	} {
		if err := store.Event(at, event, nil); err != nil {
			t.Fatalf("observe event: %v", err)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/runs/run-1", nil)
	request = request.WithContext(observation.WithRegistry(request.Context(), registry))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /runs/run-1: status %d, body %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	// The value travels under the key src/hooks.go and src/hooks.ts both
	// spell; a document without it is one the client could not decode.
	if !strings.Contains(body, "RunSnapshot") {
		t.Fatalf("the rendered document carries no transported snapshot:\n%s", body)
	}
	// And what it carries is this run's actual observation, not an empty one.
	for _, want := range []string{"run-1", "turn-1", "hello from Gimble", "test-model"} {
		if !strings.Contains(body, want) {
			t.Fatalf("the rendered document does not mention %q:\n%s", want, body)
		}
	}

	// An unknown run is an ordinary 404, not an empty page.
	missing := httptest.NewRequest(http.MethodGet, "/runs/nope", nil)
	missing = missing.WithContext(observation.WithRegistry(missing.Context(), registry))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, missing)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GET /runs/nope: status %d, want 404", recorder.Code)
	}
}
