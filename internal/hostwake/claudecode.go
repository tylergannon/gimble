package hostwake

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Environment variables a Claude Code session exports into every process it
// launches. A run started from inside a session therefore already holds the
// coordinates of the session that started it.
const (
	socketEnv = "CLAUDE_CODE_MESSAGING_SOCKET"
	tokenEnv  = "CLAUDE_CODE_MESSAGING_TOKEN"
)

// ClaudeCode wakes a Claude Code session over its messaging socket.
//
// Two newline-delimited JSON frames carry a turn: an auth frame, then a user
// frame. The user frame shape matters. A {"type":"message"} frame is accepted
// by the socket and silently discarded; only the frame written here is
// delivered.
type ClaudeCode struct {
	// LookupEnv reads the process environment. Nil means os.LookupEnv.
	LookupEnv func(string) (string, bool)
	// SessionsDir is the host's session registry. Empty means ~/.claude/sessions.
	SessionsDir string
	// Timeout bounds dialling and writing. Zero means five seconds.
	Timeout time.Duration
}

// NewClaudeCode returns a channel reading the ambient environment and the
// default session registry.
func NewClaudeCode() *ClaudeCode { return &ClaudeCode{} }

func (c *ClaudeCode) lookup(name string) string {
	lookupEnv := c.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	value, _ := lookupEnv(name)
	return strings.TrimSpace(value)
}

func (c *ClaudeCode) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 5 * time.Second
}

func (c *ClaudeCode) sessionsDir() string {
	if strings.TrimSpace(c.SessionsDir) != "" {
		return c.SessionsDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "sessions")
}

// Capture reads the launching session out of the inherited environment and
// enriches it from the session registry, which is the only place the session
// id and its interactive-or-background kind are recorded.
func (c *ClaudeCode) Capture() (Session, bool) {
	socket := c.lookup(socketEnv)
	if socket == "" {
		return Session{}, false
	}
	session := Session{Host: HostClaudeCode, Socket: socket}
	if record, ok := c.findRecord(socket); ok {
		session.SessionID = record.SessionID
		session.PID = record.PID
		session.Kind = record.Kind
		session.Name = record.Name
	}
	return session, true
}

// Available reports whether the session's socket still accepts a connection.
func (c *ClaudeCode) Available(session Session) bool {
	if session.Host != HostClaudeCode || strings.TrimSpace(session.Socket) == "" {
		return false
	}
	connection, err := net.DialTimeout("unix", session.Socket, c.timeout())
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

// Wake delivers message as a user turn. A session that has already gone idle
// picks the turn up and acts on it.
func (c *ClaudeCode) Wake(session Session, message string) error {
	if session.Host != HostClaudeCode || strings.TrimSpace(session.Socket) == "" {
		return errors.New("no Claude Code session captured")
	}
	if len(message) > MaxMessageBytes {
		return fmt.Errorf("message is %d bytes, over the host's %d byte cap", len(message), MaxMessageBytes)
	}
	token, err := c.token(session)
	if err != nil {
		return err
	}
	connection, err := net.DialTimeout("unix", session.Socket, c.timeout())
	if err != nil {
		return fmt.Errorf("dial host session: %w", err)
	}
	defer func() { _ = connection.Close() }()
	if err := connection.SetDeadline(time.Now().Add(c.timeout())); err != nil {
		return fmt.Errorf("set host session deadline: %w", err)
	}
	encoder := json.NewEncoder(connection)
	if err := encoder.Encode(authFrame{Type: "auth", Token: token}); err != nil {
		return fmt.Errorf("authenticate to host session: %w", err)
	}
	if err := encoder.Encode(userFrame{Type: "user", Message: userMessage{Role: "user", Content: message}}); err != nil {
		return fmt.Errorf("deliver host session turn: %w", err)
	}
	return c.readRejection(connection)
}

// readRejection reports an error the host sends back. The host stays silent
// on success, so a read that times out is a delivery.
func (c *ClaudeCode) readRejection(connection net.Conn) error {
	if err := connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		return nil
	}
	buffer := make([]byte, 4096)
	read, err := connection.Read(buffer)
	if err != nil || read == 0 {
		return nil
	}
	var reply struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(firstLine(string(buffer[:read]))), &reply); err != nil {
		return nil
	}
	if reply.Error != "" {
		return fmt.Errorf("host session rejected the wake: %s", reply.Error)
	}
	if reply.Type == "error" {
		return errors.New("host session rejected the wake")
	}
	return nil
}

func firstLine(raw string) string {
	line, _, _ := strings.Cut(raw, "\n")
	return line
}

// token resolves the credential for session. The inherited environment holds
// it whenever this process was launched by that same session; otherwise the
// registry's peer token addresses the same socket.
func (c *ClaudeCode) token(session Session) (string, error) {
	if socket := c.lookup(socketEnv); socket != "" && socket == session.Socket {
		if token := c.lookup(tokenEnv); token != "" {
			return token, nil
		}
	}
	if session.PID > 0 {
		if token, err := c.registryToken(session.PID); err == nil && token != "" {
			return token, nil
		}
	}
	return "", errors.New("no credential for the host session")
}

type authFrame struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type userMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type userFrame struct {
	Type    string      `json:"type"`
	Message userMessage `json:"message"`
}

// sessionRecord is the part of ~/.claude/sessions/<pid>.json Gimble reads.
type sessionRecord struct {
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Socket    string `json:"messagingSocketPath"`
}

func (c *ClaudeCode) findRecord(socket string) (sessionRecord, bool) {
	dir := c.sessionsDir()
	if dir == "" {
		return sessionRecord{}, false
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return sessionRecord{}, false
	}
	for _, entry := range entries {
		raw, err := os.ReadFile(entry)
		if err != nil {
			continue
		}
		var record sessionRecord
		if err := json.Unmarshal(raw, &record); err != nil {
			continue
		}
		if record.Socket == socket {
			return record, true
		}
	}
	return sessionRecord{}, false
}

func (c *ClaudeCode) registryToken(pid int) (string, error) {
	dir := c.sessionsDir()
	if dir == "" {
		return "", errors.New("no session registry")
	}
	matches, err := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%d.*.key", pid)))
	if err != nil || len(matches) == 0 {
		return "", errors.New("no session key file")
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		return "", err
	}
	var key struct {
		PeerToken string `json:"peerToken"`
	}
	if err := json.Unmarshal(raw, &key); err != nil {
		return "", err
	}
	if key.PeerToken == "" {
		return "", errors.New("session key file has no peer token")
	}
	return key.PeerToken, nil
}
