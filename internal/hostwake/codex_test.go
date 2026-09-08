package hostwake

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCodexQueueDeliversToExactThread(t *testing.T) {
	channel := &CodexQueue{
		Binary: "codex-test",
		Run: func(_ context.Context, binary string, args ...string) ([]byte, error) {
			want := []string{"queue", "--thread", "01a07eab-99d5-7923-84ac-f41374506751", "--message", "run news"}
			if binary != "codex-test" || strings.Join(args, "\x00") != strings.Join(want, "\x00") {
				t.Fatalf("command = %q %q", binary, args)
			}
			return []byte("Queued message item-1"), nil
		},
	}
	session := Session{Host: HostCodexDesktop, ThreadID: "01a07eab-99d5-7923-84ac-f41374506751"}
	if err := channel.Wake(session, "run news"); err != nil {
		t.Fatal(err)
	}
}

func TestCodexQueueFailureIsObservable(t *testing.T) {
	channel := &CodexQueue{
		Timeout: time.Second,
		Run: func(context.Context, string, ...string) ([]byte, error) {
			return []byte("thread is not loaded"), errors.New("exit 1")
		},
	}
	err := channel.Wake(Session{Host: HostCodexDesktop, ThreadID: "thread-1"}, "news")
	if err == nil || !strings.Contains(err.Error(), "thread is not loaded") {
		t.Fatalf("Wake() error = %v", err)
	}
}

func TestCodexQueueRequiresExplicitThread(t *testing.T) {
	channel := &CodexQueue{}
	if _, ok := channel.Capture(); ok {
		t.Fatal("Codex queue inferred a parent")
	}
	if err := channel.Wake(Session{Host: HostCodexDesktop}, "news"); err == nil {
		t.Fatal("Wake() accepted an empty thread")
	}
}
