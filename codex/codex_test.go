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
// connection's reader has already failed there is no daemon to talk to, but
// Close must still release everything the adapter holds locally: the
// session entry and the thread's routing channel on that connection.
func TestCloseOnADeadConnectionDropsLocalRoutingState(t *testing.T) {
	ad := New().(*adapter)
	conn := &connection{threads: make(map[string]chan rpcMessage), readDone: make(chan struct{})}
	conn.registerThread("thread-1")
	close(conn.readDone) // the reader has failed: conn.dead() is true
	ad.sharedConn = conn
	ad.sessions["thread-1"] = &session{}

	if err := ad.Close(context.Background(), "thread-1"); err != nil {
		t.Fatalf("Close = %v, want nil", err)
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
