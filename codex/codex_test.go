package codex

import (
	"context"
	"testing"
)

// TestCloseIsIdempotentForAnUnknownSession covers acceptance item 4: Close
// on an id the adapter never registered (no session, no connection) returns
// nil both times, and needs no daemon: the adapter's session map does not
// contain the id, so Close returns before it ever touches the connection.
func TestCloseIsIdempotentForAnUnknownSession(t *testing.T) {
	ad := New().(*adapter)
	if err := ad.Close(context.Background(), "unknown-thread"); err != nil {
		t.Fatalf("first Close = %v, want nil", err)
	}
	if err := ad.Close(context.Background(), "unknown-thread"); err != nil {
		t.Fatalf("second Close = %v, want nil", err)
	}
}

// TestCloseOnADeadConnectionDropsLocalRoutingState: when the shared
// connection's reader has already failed, Close must still release
// everything the adapter holds locally (the session entry and the thread's
// routing channel on that connection) before it tries to reach the daemon.
// The cleanup context here is already cancelled, so the redial cannot
// succeed: Close must then report the archive it could not do rather than
// return nil, and the local state must be gone regardless.
func TestCloseOnADeadConnectionDropsLocalRoutingState(t *testing.T) {
	ad := New().(*adapter)
	conn := &connection{threads: make(map[string]chan rpcMessage), readDone: make(chan struct{})}
	conn.registerThread("thread-1")
	close(conn.readDone) // the reader has failed: conn.dead() is true
	ad.sharedConn = conn
	ad.sessions["thread-1"] = &session{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ad.Close(ctx, "thread-1"); err == nil {
		t.Fatal("Close = nil, want an error: the archive could not run on a cancelled cleanup context")
	}
	if n := len(ad.sessions); n != 0 {
		t.Fatalf("sessions after Close = %d, want 0", n)
	}
	conn.threadsMu.Lock()
	n := len(conn.threads)
	conn.threadsMu.Unlock()
	if n != 0 {
		t.Fatalf("connection threads after Close = %d, want 0", n)
	}
}
