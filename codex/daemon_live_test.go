package codex

import (
	"context"
	"os"
	"strings"
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
