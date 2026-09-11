// Package codex is Gimble's HarnessAdapter for Codex, through `codex
// app-server`. Each turn runs in its own app-server process that resumes
// the thread; the process that started or forked a thread is kept for its
// first turn, because a thread with no turn is not yet resumable.
package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tylergannon/gimble"
)

const (
	requestTimeout = 30 * time.Second
	controlTimeout = 5 * time.Second
)

// Adapter runs Codex sessions. The zero value is not usable; call New.
type Adapter struct {
	mu       sync.Mutex
	sessions map[string]*session
}

type session struct {
	model   string
	workdir string

	mu     sync.Mutex
	fresh  *connection // the process that made the thread, until its first turn
	active *activeTurn
}

type activeTurn struct {
	conn   *connection
	turnID string
	emit   *projector
}

// New returns an adapter that launches `codex app-server --stdio`.
func New() *Adapter {
	return &Adapter{sessions: make(map[string]*session)}
}

// CreateSession starts a Codex thread.
func (a *Adapter) CreateSession(ctx context.Context, model, workdir string) (string, error) {
	return a.thread(ctx, "thread/start", map[string]any{
		"model":          model,
		"cwd":            workdir,
		"approvalPolicy": "never",
		"sandbox":        "danger-full-access",
		"serviceName":    "gimble",
	}, &session{model: model, workdir: workdir})
}

// Fork starts a new thread with the conversation of sessionID so far.
func (a *Adapter) Fork(ctx context.Context, sessionID string) (string, error) {
	parent, err := a.session(sessionID)
	if err != nil {
		return "", err
	}
	return a.thread(ctx, "thread/fork", map[string]any{
		"threadId":       sessionID,
		"cwd":            parent.workdir,
		"approvalPolicy": "never",
		"sandbox":        "danger-full-access",
	}, &session{model: parent.model, workdir: parent.workdir})
}

// thread calls a method that makes a thread and keeps its process for the
// thread's first turn.
func (a *Adapter) thread(ctx context.Context, method string, params map[string]any, s *session) (string, error) {
	conn, err := start(ctx)
	if err != nil {
		return "", err
	}
	callCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	result, err := conn.call(callCtx, method, params)
	if err != nil {
		conn.close()
		return "", err
	}
	id, err := threadID(result)
	if err != nil {
		conn.close()
		return "", err
	}
	s.fresh = conn
	a.mu.Lock()
	a.sessions[id] = s
	a.mu.Unlock()
	return id, nil
}

// RunTurn runs one turn on the thread and blocks until it ends.
func (a *Adapter) RunTurn(ctx context.Context, sessionID, prompt string, schema json.RawMessage, onEvent func(gimble.Event)) (json.RawMessage, error) {
	s, err := a.session(sessionID)
	if err != nil {
		return nil, err
	}
	conn, err := s.open(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	defer conn.close()

	params := map[string]any{
		"threadId":       sessionID,
		"input":          input(prompt),
		"model":          s.model,
		"cwd":            s.workdir,
		"approvalPolicy": "never",
		"sandboxPolicy":  map[string]any{"type": "dangerFullAccess"},
	}
	if len(schema) > 0 {
		params["outputSchema"] = schema
	}
	startCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	result, err := conn.call(startCtx, "turn/start", params)
	cancel()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	turn, err := turnID(result)
	if err != nil {
		return nil, err
	}

	active := &activeTurn{conn: conn, turnID: turn, emit: &projector{emit: onEvent, started: map[string]bool{}}}
	s.setActive(active)
	defer s.setActive(nil)
	active.emit.user(prompt)

	text, err := readTurn(ctx, conn, sessionID, turn, active.emit)
	if ctx.Err() != nil {
		interrupt(conn, sessionID, turn)
		drainCtx, drainCancel := context.WithTimeout(context.Background(), controlTimeout)
		_, _ = readTurn(drainCtx, conn, sessionID, turn, active.emit)
		drainCancel()
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	if len(schema) == 0 {
		return json.Marshal(text)
	}
	return json.RawMessage(strings.TrimSpace(text)), nil
}

// Steer sends message into the thread's running turn, if there is one.
func (a *Adapter) Steer(ctx context.Context, sessionID, message string) error {
	s, err := a.session(sessionID)
	if err != nil {
		return err
	}
	active := s.getActive()
	if active == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, controlTimeout)
	defer cancel()
	_, err = active.conn.call(ctx, "turn/steer", map[string]any{
		"threadId":       sessionID,
		"expectedTurnId": active.turnID,
		"input":          input(message),
	})
	if err != nil {
		if s.getActive() != active {
			return nil // the turn ended first: the message is dropped
		}
		return err
	}
	active.emit.user(message)
	return nil
}

func (a *Adapter) Interrupt(ctx context.Context, sessionID string) error {
	s, err := a.session(sessionID)
	if err != nil {
		return err
	}
	active := s.getActive()
	if active == nil {
		return nil
	}
	interrupt(active.conn, sessionID, active.turnID)
	return nil
}

func (a *Adapter) session(id string) (*session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if s := a.sessions[id]; s != nil {
		return s, nil
	}
	return nil, fmt.Errorf("codex: no session %q", id)
}

func (s *session) setActive(active *activeTurn) {
	s.mu.Lock()
	s.active = active
	s.mu.Unlock()
}

func (s *session) getActive() *activeTurn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// open returns a process attached to the thread: the one that made it if
// no turn has used it yet, otherwise a new one that resumes the thread.
// The caller closes it.
func (s *session) open(ctx context.Context, id string) (*connection, error) {
	s.mu.Lock()
	fresh := s.fresh
	s.fresh = nil
	s.mu.Unlock()
	if fresh != nil {
		return fresh, nil
	}
	conn, err := start(ctx)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	result, err := conn.call(callCtx, "thread/resume", map[string]any{
		"threadId":       id,
		"cwd":            s.workdir,
		"approvalPolicy": "never",
		"sandbox":        "danger-full-access",
	})
	if err != nil {
		conn.close()
		return nil, err
	}
	if resumed, err := threadID(result); err != nil || resumed != id {
		conn.close()
		return nil, fmt.Errorf("codex: thread/resume of %s returned %q: %v", id, resumed, err)
	}
	return conn, nil
}

func interrupt(conn *connection, threadID, turnID string) {
	ctx, cancel := context.WithTimeout(context.Background(), controlTimeout)
	defer cancel()
	_, _ = conn.call(ctx, "turn/interrupt", map[string]any{"threadId": threadID, "turnId": turnID})
}

func input(text string) []map[string]any {
	return []map[string]any{{"type": "text", "text": text, "text_elements": []any{}}}
}

// readTurn consumes notifications until the turn completes and returns the
// agent's final message.
func readTurn(ctx context.Context, conn *connection, threadID, turnID string, emit *projector) (string, error) {
	var final string
	for {
		message, err := conn.next(ctx)
		if err != nil {
			return "", err
		}
		if len(message.ID) > 0 && message.Method != "" {
			return "", refuse(conn, message)
		}
		if !matches(message.Params, threadID, turnID) {
			continue
		}
		switch message.Method {
		case "item/started":
			emit.itemStarted(message.Params)
		case "item/completed":
			if text, ok := emit.itemCompleted(message.Params); ok {
				final = text
			}
		case "thread/tokenUsage/updated":
			emit.usage(message.Params)
		case "turn/completed":
			status, failure := completedTurn(message.Params)
			switch status {
			case "completed":
				if strings.TrimSpace(final) == "" {
					return "", errors.New("codex: the turn completed without a final message")
				}
				return final, nil
			case "interrupted":
				return "", errors.New("codex: the turn was interrupted")
			case "failed":
				if failure == "" {
					failure = "the turn failed"
				}
				return "", errors.New("codex: " + failure)
			default:
				return "", fmt.Errorf("codex: the turn ended with status %q", status)
			}
		case "error":
			return "", errorNotification(message.Params)
		}
	}
}

// refuse declines an interactive request; Gimble turns run with approvals
// off and never answer prompts.
func refuse(conn *connection, message rpcMessage) error {
	result := map[string]any{"decision": "decline"}
	switch message.Method {
	case "execCommandApproval", "applyPatchApproval":
		result["decision"] = "denied"
	case "item/tool/requestUserInput":
		result = map[string]any{"answers": map[string]any{}}
	case "mcpServer/elicitation/request":
		result = map[string]any{"action": "decline"}
	case "item/permissions/requestApproval":
		result = map[string]any{"permissions": map[string]any{}, "scope": "turn"}
	}
	if err := conn.respond(message, result); err != nil {
		return err
	}
	return fmt.Errorf("codex: unexpected interactive request %q", message.Method)
}

func threadID(raw json.RawMessage) (string, error) {
	var response struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("codex: decode thread: %w", err)
	}
	if response.Thread.ID == "" {
		return "", errors.New("codex: the response has no thread.id")
	}
	return response.Thread.ID, nil
}

func turnID(raw json.RawMessage) (string, error) {
	var response struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("codex: decode turn: %w", err)
	}
	if response.Turn.ID == "" {
		return "", errors.New("codex: the response has no turn.id")
	}
	return response.Turn.ID, nil
}

func completedTurn(raw json.RawMessage) (status, failure string) {
	var notification struct {
		Turn struct {
			Status string `json:"status"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"turn"`
	}
	_ = json.Unmarshal(raw, &notification)
	if notification.Turn.Error != nil {
		failure = notification.Turn.Error.Message
	}
	return notification.Turn.Status, failure
}

func errorNotification(raw json.RawMessage) error {
	var notification struct {
		Message string `json:"message"`
		Error   *struct {
			Message           string `json:"message"`
			AdditionalDetails string `json:"additionalDetails"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &notification) != nil {
		return fmt.Errorf("codex: error notification: %s", strings.TrimSpace(string(raw)))
	}
	if notification.Message != "" {
		return errors.New("codex: " + notification.Message)
	}
	if e := notification.Error; e != nil && e.Message != "" {
		if e.AdditionalDetails != "" {
			return fmt.Errorf("codex: %s: %s", e.Message, e.AdditionalDetails)
		}
		return errors.New("codex: " + e.Message)
	}
	return errors.New("codex: an empty error notification")
}

// matches reports whether a notification belongs to the turn.
func matches(raw json.RawMessage, threadID, turnID string) bool {
	var envelope struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Turn     struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if json.Unmarshal(raw, &envelope) != nil || envelope.ThreadID != threadID {
		return false
	}
	if envelope.TurnID != "" {
		return envelope.TurnID == turnID
	}
	return envelope.Turn.ID == turnID
}

var _ gimble.HarnessAdapter = (*Adapter)(nil)
