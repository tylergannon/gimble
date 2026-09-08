package hostwake

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fakeHost is a stand-in for a Claude Code session's messaging socket.
type fakeHost struct {
	path   string
	frames chan []map[string]any
}

func startFakeHost(t *testing.T) *fakeHost {
	t.Helper()
	// Socket paths are capped near 104 bytes, so keep it out of t.TempDir().
	dir, err := os.MkdirTemp("", "wake")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	host := &fakeHost{path: filepath.Join(dir, "s.sock"), frames: make(chan []map[string]any, 8)}
	listener, err := net.Listen("unix", host.path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = connection.Close() }()
				frames := []map[string]any{}
				scanner := bufio.NewScanner(connection)
				for scanner.Scan() {
					var frame map[string]any
					if err := json.Unmarshal(scanner.Bytes(), &frame); err == nil {
						frames = append(frames, frame)
					}
				}
				if len(frames) > 0 {
					host.frames <- frames
				}
			}()
		}
	}()
	return host
}

func (h *fakeHost) received(t *testing.T) []map[string]any {
	t.Helper()
	select {
	case frames := <-h.frames:
		return frames
	case <-time.After(2 * time.Second):
		t.Fatal("host received no frames")
		return nil
	}
}

func writeRegistry(t *testing.T, dir string, record map[string]any, key map[string]any) {
	t.Helper()
	pid, _ := record["pid"].(int)
	write := func(name string, value any) {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(strconv.Itoa(pid)+".json", record)
	if key != nil {
		write(strconv.Itoa(pid)+".abc123.key", key)
	}
}

func TestCaptureReadsEnvironmentAndEnrichesFromRegistry(t *testing.T) {
	registry := t.TempDir()
	writeRegistry(t, registry, map[string]any{
		"pid": 4242, "sessionId": "session-1", "kind": "bg",
		"name": "worker-3", "messagingSocketPath": "/tmp/cc-socks/4242.sock",
	}, nil)
	channel := &ClaudeCode{
		SessionsDir: registry,
		LookupEnv:   staticEnv(map[string]string{socketEnv: "/tmp/cc-socks/4242.sock", tokenEnv: "token-1"}),
	}

	session, ok := channel.Capture()
	if !ok {
		t.Fatal("Capture() reported no host session")
	}
	if session.Host != HostClaudeCode || session.SessionID != "session-1" || session.PID != 4242 {
		t.Fatalf("session = %#v", session)
	}
	if session.Kind != KindBackground || session.Interactive() {
		t.Fatalf("session kind = %q, interactive = %v", session.Kind, session.Interactive())
	}
}

func TestBothSpellingsOfBackgroundAreBackground(t *testing.T) {
	for _, kind := range []string{"bg", "background", "BG"} {
		if (Session{Kind: kind}).Interactive() {
			t.Fatalf("kind %q counts as interactive", kind)
		}
	}
	for _, kind := range []string{"interactive", "", "sdk"} {
		if !(Session{Kind: kind}).Interactive() {
			t.Fatalf("kind %q counts as background", kind)
		}
	}
}

func TestCaptureReportsNoSessionWithoutSocket(t *testing.T) {
	channel := &ClaudeCode{SessionsDir: t.TempDir(), LookupEnv: staticEnv(nil)}
	if _, ok := channel.Capture(); ok {
		t.Fatal("Capture() invented a host session")
	}
}

func TestCaptureWithoutRegistryRecordStaysInteractive(t *testing.T) {
	channel := &ClaudeCode{
		SessionsDir: t.TempDir(),
		LookupEnv:   staticEnv(map[string]string{socketEnv: "/tmp/cc-socks/9.sock", tokenEnv: "t"}),
	}
	session, ok := channel.Capture()
	if !ok {
		t.Fatal("Capture() reported no host session")
	}
	if !session.Interactive() {
		t.Fatal("an unknown session kind must count as interactive")
	}
}

func TestWakeDeliversAuthenticatedUserTurn(t *testing.T) {
	host := startFakeHost(t)
	channel := &ClaudeCode{
		SessionsDir: t.TempDir(),
		LookupEnv:   staticEnv(map[string]string{socketEnv: host.path, tokenEnv: "token-1"}),
	}
	session := Session{Host: HostClaudeCode, Socket: host.path, Kind: KindBackground}

	if err := channel.Wake(session, "run news"); err != nil {
		t.Fatal(err)
	}

	frames := host.received(t)
	if len(frames) != 2 {
		t.Fatalf("frames = %#v", frames)
	}
	if frames[0]["type"] != "auth" || frames[0]["token"] != "token-1" {
		t.Fatalf("auth frame = %#v", frames[0])
	}
	if frames[1]["type"] != "user" {
		t.Fatalf("turn frame = %#v", frames[1])
	}
	message, ok := frames[1]["message"].(map[string]any)
	if !ok || message["role"] != "user" || message["content"] != "run news" {
		t.Fatalf("turn message = %#v", frames[1]["message"])
	}
}

func TestWakeFallsBackToRegistryTokenWhenEnvironmentIsAnotherSession(t *testing.T) {
	host := startFakeHost(t)
	registry := t.TempDir()
	writeRegistry(t, registry,
		map[string]any{"pid": 77, "sessionId": "s", "kind": "background", "messagingSocketPath": host.path},
		map[string]any{"peerToken": "peer-token"},
	)
	channel := &ClaudeCode{
		SessionsDir: registry,
		LookupEnv:   staticEnv(map[string]string{socketEnv: "/tmp/other.sock", tokenEnv: "not-ours"}),
	}
	session := Session{Host: HostClaudeCode, Socket: host.path, PID: 77, Kind: KindBackground}

	if err := channel.Wake(session, "news"); err != nil {
		t.Fatal(err)
	}
	frames := host.received(t)
	if frames[0]["token"] != "peer-token" {
		t.Fatalf("auth frame = %#v", frames[0])
	}
}

func TestWakeWithoutCredentialFails(t *testing.T) {
	host := startFakeHost(t)
	channel := &ClaudeCode{SessionsDir: t.TempDir(), LookupEnv: staticEnv(nil)}
	err := channel.Wake(Session{Host: HostClaudeCode, Socket: host.path}, "news")
	if err == nil || !strings.Contains(err.Error(), "credential") {
		t.Fatalf("Wake() error = %v", err)
	}
}

func TestWakeRefusesAnOversizedMessage(t *testing.T) {
	channel := &ClaudeCode{SessionsDir: t.TempDir(), LookupEnv: staticEnv(nil)}
	err := channel.Wake(Session{Host: HostClaudeCode, Socket: "/tmp/x.sock"}, strings.Repeat("x", MaxMessageBytes+1))
	if err == nil || !strings.Contains(err.Error(), "cap") {
		t.Fatalf("Wake() error = %v", err)
	}
}

func TestAvailableFollowsTheSocket(t *testing.T) {
	host := startFakeHost(t)
	channel := &ClaudeCode{SessionsDir: t.TempDir(), LookupEnv: staticEnv(nil)}
	if !channel.Available(Session{Host: HostClaudeCode, Socket: host.path}) {
		t.Fatal("a live socket is not available")
	}
	if channel.Available(Session{Host: HostClaudeCode, Socket: host.path + ".gone"}) {
		t.Fatal("a missing socket is available")
	}
}

func staticEnv(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
