package codex

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// connection is one `codex app-server` process, spoken to in JSON-RPC over
// stdio.
type connection struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stderr    *lockedBuffer
	nextID    atomic.Int64
	writeMu   sync.Mutex
	pendMu    sync.Mutex
	pending   map[int64]chan rpcMessage
	messages  chan rpcMessage
	readDone  chan struct{}
	readOnce  sync.Once
	readErr   error
	waitDone  chan struct{}
	closeOnce sync.Once
}

type rpcMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("codex app-server error %d: %s", e.Code, e.Message)
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// start launches an app-server and initializes it.
func start(ctx context.Context) (*connection, error) {
	cmd := exec.Command("codex", "app-server", "--stdio")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr := &lockedBuffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("codex: start app-server: %w", err)
	}
	c := &connection{
		cmd:      cmd,
		stdin:    stdin,
		stderr:   stderr,
		pending:  make(map[int64]chan rpcMessage),
		messages: make(chan rpcMessage, 512),
		readDone: make(chan struct{}),
		waitDone: make(chan struct{}),
	}
	go c.read(stdout)
	go func() {
		err := cmd.Wait()
		close(c.waitDone)
		c.finishRead(err)
	}()

	initCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	_, err = c.call(initCtx, "initialize", map[string]any{
		"clientInfo":   map[string]any{"name": "gimble", "title": "Gimble", "version": "dev"},
		"capabilities": map[string]any{},
	})
	if err == nil {
		err = c.send(map[string]any{"method": "initialized", "params": map[string]any{}})
	}
	if err != nil {
		c.close()
		return nil, err
	}
	return c, nil
}

func (c *connection) read(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var message rpcMessage
		if err := json.Unmarshal(line, &message); err != nil {
			c.finishRead(fmt.Errorf("codex: decode app-server message: %w", err))
			return
		}
		c.dispatch(message)
	}
	if err := scanner.Err(); err != nil {
		c.finishRead(err)
		return
	}
	c.finishRead(io.EOF)
}

func (c *connection) finishRead(err error) {
	c.readOnce.Do(func() {
		c.readErr = err
		close(c.readDone)
	})
}

// dispatch routes a response to its caller and everything else to next.
func (c *connection) dispatch(message rpcMessage) {
	if len(message.ID) > 0 && message.Method == "" {
		if id, ok := parseID(message.ID); ok {
			c.pendMu.Lock()
			response := c.pending[id]
			c.pendMu.Unlock()
			if response != nil {
				response <- message
				return
			}
		}
	}
	select {
	case c.messages <- message:
	case <-c.readDone:
	}
}

func (c *connection) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	response := make(chan rpcMessage, 1)
	c.pendMu.Lock()
	c.pending[id] = response
	c.pendMu.Unlock()
	defer func() {
		c.pendMu.Lock()
		delete(c.pending, id)
		c.pendMu.Unlock()
	}()
	if err := c.send(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	select {
	case message := <-response:
		if message.Error != nil {
			return nil, fmt.Errorf("%s: %w", method, message.Error)
		}
		return message.Result, nil
	case <-c.readDone:
		return nil, c.exitError()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *connection) respond(message rpcMessage, result any) error {
	return c.send(map[string]any{"id": message.ID, "result": result})
}

func (c *connection) send(value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err = c.stdin.Write(append(encoded, '\n'))
	return err
}

// next returns the next notification or server request.
func (c *connection) next(ctx context.Context) (rpcMessage, error) {
	select {
	case message := <-c.messages:
		return message, nil
	default:
	}
	select {
	case message := <-c.messages:
		return message, nil
	case <-c.readDone:
		return rpcMessage{}, c.exitError()
	case <-ctx.Done():
		return rpcMessage{}, ctx.Err()
	}
}

func (c *connection) exitError() error {
	if stderr := strings.TrimSpace(c.stderr.String()); stderr != "" {
		return fmt.Errorf("codex app-server exited: %s: %w", stderr, c.readErr)
	}
	return fmt.Errorf("codex app-server exited: %w", c.readErr)
}

func (c *connection) close() {
	c.closeOnce.Do(func() {
		_ = c.stdin.Close()
		select {
		case <-c.waitDone:
		case <-time.After(2 * time.Second):
			_ = c.cmd.Process.Kill()
			<-c.waitDone
		}
	})
}

func parseID(raw json.RawMessage) (int64, bool) {
	var number int64
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, true
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, false
	}
	number, err := strconv.ParseInt(text, 10, 64)
	return number, err == nil
}
