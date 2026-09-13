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

		// Kill the connection the first turn used, the same way an
		// ordinary network blip or the daemon-side idle timeout would.
		// The next call must redial and, per the fix, resume this
		// thread before running the next turn on it.
		ad.connMu.Lock()
		conn := ad.sharedConn
		ad.connMu.Unlock()
		if conn == nil {
			t.Fatal("adapter has no connection to kill")
		}
		conn.ws.CloseNow()
		<-conn.readDone // wait for the reader to notice, so the next call redials deterministically

		second, err := session.Generate[gimble.Text](ctx, "Reply with exactly one word: two.")
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(second)) == "" {
			t.Fatal("second turn (after redial) returned no text")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
