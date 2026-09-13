package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tylergannon/gimble"
)

// TestRunTurnAfterRedialResumesThread proves finding 1's fix live: killing
// the adapter's current WebSocket out from under it must not orphan a turn
// on a session that already exists. It talks to the machine's real shared
// `codex app-server` daemon (never restarting it) using the cheap model
// gpt-5.6-luna, so it only runs when explicitly requested:
//
//	GIMBLE_LIVE=1 go test ./codex -run TestRunTurnAfterRedialResumesThread -v
//
// Before the fix in adapter.conn/resumeThreads, the second turn below hangs
// until the bounded context expires: turn/start succeeds on the redialed
// connection, but that connection was never subscribed to the thread, so
// its notifications go nowhere and readTurn never sees turn/completed.
// After the fix, the redial resumes every known thread before handing the
// connection back, and the second turn returns text well within the bound.
func TestRunTurnAfterRedialResumesThread(t *testing.T) {
	if os.Getenv("GIMBLE_LIVE") != "1" {
		t.Skip("set GIMBLE_LIVE=1 to run against the live codex app-server daemon")
	}

	ad, ok := New().(*adapter)
	if !ok {
		t.Fatalf("codex.New() did not return *adapter")
	}

	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ctx = gimble.Project(ctx, dir)

	err := gimble.Run(ctx, "daemon-live", func(ctx context.Context) error {
		session := gimble.NewSession(ctx, "live", ad, "gpt-5.6-luna", dir)

		first, err := session.Generate[gimble.Text](ctx, "Reply with exactly one word: one.")
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(first)) == "" {
			t.Fatal("first turn returned no text")
		}
		conn := ad.current()
		if conn == nil {
			t.Fatal("adapter has no connection after the first turn")
		}

		// An ordinary second turn reuses the connection the first turn
		// dialed: this is the one-connection contract, observed directly
		// rather than inferred from process counts.
		second, err := session.Generate[gimble.Text](ctx, "Reply with exactly one word: two.")
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(second)) == "" {
			t.Fatal("second turn returned no text")
		}
		if ad.current() != conn {
			t.Fatal("an ordinary second turn opened a new connection instead of reusing the first")
		}

		// Kill that connection, the same way a network blip or the
		// daemon-side idle timeout would. The next call must redial and,
		// per the fix, resume this thread before running a turn on it.
		conn.ws.CloseNow()
		<-conn.readDone // wait for the reader to notice, so the next call redials deterministically

		third, err := session.Generate[gimble.Text](ctx, "Reply with exactly one word: three.")
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(third)) == "" {
			t.Fatal("third turn (after redial) returned no text")
		}
		if ad.current() == conn {
			t.Fatal("the turn after the kill ran on the dead connection")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// current returns the adapter's shared connection without dialing.
func (a *adapter) current() *connection {
	a.connMu.Lock()
	defer a.connMu.Unlock()
	return a.sharedConn
}

// TestCloseUnsubscribesThreadsWithoutTouchingTheDaemon proves commit 2's
// contract live: HarnessAdapter.Close on the Codex adapter deregisters a
// session's routing and asks the daemon to stop notifying this connection
// about it (thread/unsubscribe), but never stops or restarts the shared
// daemon, and never archives or deletes the thread. It creates a session,
// runs one short turn, forks it without ever running a turn on the fork
// (proving Close also releases a session whose only allocation was
// thread/fork, not thread/start), and lets gimble.Run's root scope end.
// It talks to the machine's real shared `codex app-server` daemon, using
// the cheap model gpt-5.6-luna, so it only runs when explicitly requested:
//
//	GIMBLE_LIVE=1 go test ./codex -run TestCloseUnsubscribesThreadsWithoutTouchingTheDaemon -v
func TestCloseUnsubscribesThreadsWithoutTouchingTheDaemon(t *testing.T) {
	if os.Getenv("GIMBLE_LIVE") != "1" {
		t.Skip("set GIMBLE_LIVE=1 to run against the live codex app-server daemon")
	}

	beforePID, err := managedDaemonPID()
	if err != nil {
		t.Fatal(err)
	}

	ad, ok := New().(*adapter)
	if !ok {
		t.Fatalf("codex.New() did not return *adapter")
	}
	var mu sync.Mutex
	statuses := map[string]string{}
	ad.onUnsubscribe = func(threadID, status string) {
		mu.Lock()
		defer mu.Unlock()
		statuses[threadID] = status
	}

	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ctx = gimble.Project(ctx, dir)

	err = gimble.Run(ctx, "daemon-close-live", func(ctx context.Context) error {
		session := gimble.NewSession(ctx, "live", ad, "gpt-5.6-luna", dir)
		if _, err := session.Generate[gimble.Text](ctx, "Reply with exactly one word: proof."); err != nil {
			return err
		}
		// thread/fork subscribes the fork's thread immediately; forking
		// without ever running a turn on it still leaves something for
		// Close to release when the scope ends.
		if _, err := session.Fork(ctx, "forked"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ad.mu.Lock()
	remainingSessions := len(ad.sessions)
	ad.mu.Unlock()
	if remainingSessions != 0 {
		t.Fatalf("adapter session map after Run = %d entries, want 0", remainingSessions)
	}

	conn := ad.current()
	if conn == nil {
		t.Fatal("adapter has no connection after the run")
	}
	conn.threadsMu.Lock()
	remainingThreads := len(conn.threads)
	conn.threadsMu.Unlock()
	if remainingThreads != 0 {
		t.Fatalf("connection thread map after Run = %d entries, want 0", remainingThreads)
	}

	mu.Lock()
	statusesCopy := make(map[string]string, len(statuses))
	for id, status := range statuses {
		statusesCopy[id] = status
	}
	mu.Unlock()
	if len(statusesCopy) != 2 {
		t.Fatalf("thread/unsubscribe statuses = %v, want 2 (the session and its fork)", statusesCopy)
	}
	for id, status := range statusesCopy {
		t.Logf("thread/unsubscribe %s: %s", id, status)
		if status != "unsubscribed" && status != "notSubscribed" && status != "notLoaded" {
			t.Errorf("thread/unsubscribe %s returned unexpected status %q", id, status)
		}
	}

	// The adapter's connection is still alive (Close never touches it), so
	// thread/loaded/list is queried straight over it: this is the daemon's
	// own answer, not a guess from process state.
	callCtx, cancelCall := context.WithTimeout(context.Background(), requestTimeout)
	loadedRaw, err := conn.call(callCtx, "thread/loaded/list", map[string]any{})
	cancelCall()
	if err != nil {
		t.Fatal(err)
	}
	loadedIDs := loadedThreadIDs(loadedRaw)
	t.Logf("daemon thread/loaded/list after Close: %v", loadedIDs)
	for id := range statusesCopy {
		t.Logf("unsubscribed thread %s still loaded in the daemon: %v", id, slices.Contains(loadedIDs, id))
	}

	afterPID, err := managedDaemonPID()
	if err != nil {
		t.Fatal(err)
	}
	if beforePID != afterPID {
		t.Fatalf("daemon pid changed from %s to %s: Close must never restart the shared daemon", beforePID, afterPID)
	}
	t.Logf("daemon pid before=%s after=%s (unchanged)", beforePID, afterPID)
}

// loadedThreadIDs decodes thread/loaded/list's result, whose entries have
// appeared as a bare id, {"id":...}, or {"thread":{"id":...}}.
func loadedThreadIDs(raw json.RawMessage) []string {
	var response struct {
		Data []json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &response) != nil {
		return nil
	}
	var ids []string
	for _, entry := range response.Data {
		if id, ok := decodeThreadEntry(entry); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func decodeThreadEntry(entry json.RawMessage) (string, bool) {
	var bare string
	if json.Unmarshal(entry, &bare) == nil && bare != "" {
		return bare, true
	}
	var withID struct {
		ID     string `json:"id"`
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if json.Unmarshal(entry, &withID) != nil {
		return "", false
	}
	if withID.ID != "" {
		return withID.ID, true
	}
	if withID.Thread.ID != "" {
		return withID.Thread.ID, true
	}
	return "", false
}

// managedDaemonPID returns the pid of the machine's managed app-server
// daemon: the process started by `codex app-server daemon start`,
// listening on its control socket. `pgrep -fl 'app-server --listen unix'`
// also matches the shell wrapper that launched it (its command line embeds
// the same text), so this keeps only the pid whose command actually starts
// with `codex `, the same check ephemeral/attest/codex-daemon's proof uses.
func managedDaemonPID() (string, error) {
	out, err := exec.Command("pgrep", "-fl", "app-server --listen unix").Output()
	if err != nil {
		return "", fmt.Errorf("pgrep: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(fields) == 2 && strings.HasPrefix(fields[1], "codex ") {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("no managed daemon process found in:\n%s", out)
}
