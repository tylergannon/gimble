// Package hostwake delivers a run's news into the agent session that launched
// it, so a detached run can reach back to a session that has gone idle.
package hostwake

import "strings"

// Host names one kind of agent session Tractor can wake.
const (
	HostClaudeCode   = "claude_code"
	HostCodexDesktop = "codex_desktop"
)

// Session kinds as the host records them. Claude Code writes "bg" for a
// session started with --bg; "background" is accepted alongside it so a host
// that spells it out is understood too. A session whose kind could not be
// determined is treated as interactive, because that is the answer that
// leaves a human undisturbed.
const (
	KindInteractive = "interactive"
	KindBackground  = "bg"
)

var backgroundKinds = []string{KindBackground, "background"}

// Session identifies the host agent session that launched a run.
//
// Everything in it is safe to record in a run manifest offered as evidence:
// the credential needed to reach the session is deliberately absent and is
// never written to the run directory. A waker resolves it at delivery time
// from the environment or from the host's own session registry.
type Session struct {
	Host      string `json:"host"`
	HostID    string `json:"host_id,omitempty"`
	ThreadID  string `json:"thread_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Socket    string `json:"socket,omitempty"`
}

// Interactive reports whether a human is also using this session. Unknown
// kinds count as interactive.
func (s Session) Interactive() bool {
	kind := strings.TrimSpace(s.Kind)
	for _, background := range backgroundKinds {
		if strings.EqualFold(kind, background) {
			return false
		}
	}
	return true
}

// Empty reports whether no host session was captured.
func (s Session) Empty() bool {
	return strings.TrimSpace(s.Host) == ""
}

// Channel reaches one kind of host agent session.
type Channel interface {
	// Capture records the coordinates of the host session that launched this
	// process, reporting false when there is no such session.
	Capture() (Session, bool)
	// Available reports whether a captured session is still reachable.
	Available(Session) bool
	// Wake delivers one rendered digest as a user turn.
	Wake(Session, string) error
}

// MaxMessageBytes is the host's own cap on a delivered message. Wakers refuse
// anything larger rather than discovering the limit at the socket.
const MaxMessageBytes = 1 << 20
