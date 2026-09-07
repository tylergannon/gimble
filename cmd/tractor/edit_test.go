package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tylergannon/polytype/devalue"
	"github.com/tylergannon/tractor/internal/editor/generated"
)

// lineWriter delivers the first complete line written to it over a channel.
type lineWriter struct {
	mu    sync.Mutex
	buf   bytes.Buffer
	first chan string
	sent  bool
}

func newLineWriter() *lineWriter { return &lineWriter{first: make(chan string, 1)} }

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	if !w.sent {
		if line, _, ok := strings.Cut(w.buf.String(), "\n"); ok {
			w.sent = true
			w.first <- line
		}
	}
	return len(p), nil
}

func TestEditPrintsURLAndServesDocument(t *testing.T) {
	source, err := os.ReadFile("../../examples/loops/bake-off.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pipeline := filepath.Join(t.TempDir(), "bake-off.yaml")
	if err := os.WriteFile(pipeline, source, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stdout := newLineWriter()
	var stderr bytes.Buffer
	command := newRootCommand()
	command.SetOut(stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"edit", pipeline, "--no-open", "--addr", "127.0.0.1:0"})
	done := make(chan error, 1)
	go func() { done <- command.ExecuteContext(ctx) }()

	var url string
	select {
	case url = <-stdout.first:
	case err := <-done:
		t.Fatalf("command exited before printing a URL: %v\n%s", err, stderr.String())
	case <-time.After(10 * time.Second):
		t.Fatal("no URL printed")
	}
	if !strings.HasPrefix(url, "http://127.0.0.1:") || !strings.HasSuffix(url, "/") {
		t.Fatalf("url = %q", url)
	}

	// The page reads the document through the getDoc remote function, which
	// kit's client addresses by the id `skgo generate` gave it.
	var id string
	for _, r := range generated.Remotes() {
		if r.Name() == "getDoc" {
			id = r.ID()
		}
	}
	if id == "" {
		t.Fatal("no getDoc remote function is registered")
	}
	response, err := http.Get(url + "_app/remote/" + id)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("getDoc = %d: %s", response.StatusCode, body)
	}
	var envelope struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode: %v\n%s", err, body)
	}
	if envelope.Type != "result" {
		t.Fatalf("response = %s", body)
	}
	parsed, err := devalue.Parse(envelope.Data, nil)
	if err != nil {
		t.Fatalf("devalue: %v", err)
	}
	value, _ := parsed.(*devalue.Object).Get("_")
	document, ok := value.(*devalue.Object)
	if !ok {
		t.Fatalf("document = %#v", value)
	}
	if path, _ := document.Get("path"); path != pipeline {
		t.Fatalf("path = %#v", path)
	}
	if yaml, _ := document.Get("yaml"); yaml != string(source) {
		t.Fatalf("yaml = %#v", yaml)
	}
	if parseError, _ := document.Get("parse_error"); parseError != "" {
		t.Fatalf("parse_error = %#v", parseError)
	}
	if diagnostics, _ := document.Get("diagnostics"); diagnostics == nil {
		t.Fatal("diagnostics is null")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("edit returned %v\n%s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("edit did not exit after cancellation")
	}
}

func TestEditRefusesMissingOrOddPipelines(t *testing.T) {
	if _, _, err := executeCommand("edit", filepath.Join(t.TempDir(), "missing.yaml"), "--no-open"); err == nil {
		t.Fatal("missing pipeline accepted")
	}
	text := filepath.Join(t.TempDir(), "pipeline.txt")
	if err := os.WriteFile(text, []byte("name: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := executeCommand("edit", text, "--no-open"); err == nil || !strings.Contains(err.Error(), ".yaml or .yml") {
		t.Fatalf("extension error = %v", err)
	}
	asJSON := filepath.Join(t.TempDir(), "pipeline.json")
	if err := os.WriteFile(asJSON, []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := executeCommand("edit", asJSON, "--no-open"); err == nil || !strings.Contains(err.Error(), ".yaml or .yml") {
		t.Fatalf("json error = %v", err)
	}
	if _, _, err := executeCommand("edit", "--no-open"); err == nil {
		t.Fatal("missing argument accepted")
	}
}
