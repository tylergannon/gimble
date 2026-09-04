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

	response, err := http.Get(url + "api/doc")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/doc = %d: %s", response.StatusCode, body)
	}
	var document struct {
		Path        string          `json:"path"`
		YAML        string          `json:"yaml"`
		Diagnostics json.RawMessage `json:"diagnostics"`
		ParseError  string          `json:"parse_error"`
	}
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatalf("decode: %v\n%s", err, body)
	}
	if document.Path != pipeline || document.YAML != string(source) || document.ParseError != "" {
		t.Fatalf("document = %+v", document)
	}
	if string(document.Diagnostics) == "null" {
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
	if _, _, err := executeCommand("edit", text, "--no-open"); err == nil || !strings.Contains(err.Error(), ".yaml, .yml, or .json") {
		t.Fatalf("extension error = %v", err)
	}
	if _, _, err := executeCommand("edit", "--no-open"); err == nil {
		t.Fatal("missing argument accepted")
	}
}
