package claude

import (
	"encoding/json"
	"strings"
	"sync"

	claudeagent "github.com/roasbeef/claude-agent-sdk-go"
	"github.com/tylergannon/gimble"
)

// projector turns Claude Code messages into Gimble events.
type projector struct {
	mu      sync.Mutex
	emit    func(gimble.Event)
	pending []string // tool calls without a result yet, oldest first
}

func (p *projector) user(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.emit(gimble.Event{Kind: "user", Text: text})
}

func (p *projector) message(message claudeagent.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch native := message.(type) {
	case claudeagent.AssistantMessage:
		p.assistant(native)
	case *claudeagent.AssistantMessage:
		p.assistant(*native)
	case claudeagent.UserMessage:
		p.toolResult(native)
	case *claudeagent.UserMessage:
		p.toolResult(*native)
	case claudeagent.ResultMessage:
		p.usage(native.Usage)
	case *claudeagent.ResultMessage:
		p.usage(native.Usage)
	}
}

func (p *projector) assistant(message claudeagent.AssistantMessage) {
	var texts []string
	for _, block := range message.Message.Content {
		switch block.Type {
		case "text":
			if strings.TrimSpace(block.Text) != "" {
				texts = append(texts, block.Text)
			}
		case "thinking":
			if strings.TrimSpace(block.Text) != "" {
				p.emit(gimble.Event{Kind: "thinking", Text: block.Text})
			}
		case "tool_use":
			p.emit(gimble.Event{Kind: "tool_call", CallID: block.ID, Tool: block.Name, Data: json.RawMessage(block.Input)})
			p.pending = append(p.pending, block.ID)
		}
	}
	if len(texts) > 0 {
		p.emit(gimble.Event{Kind: "assistant", Text: strings.Join(texts, "")})
	}
}

func (p *projector) toolResult(message claudeagent.UserMessage) {
	if message.ToolUseResult == nil {
		return
	}
	var callID string
	if message.ParentToolUseID != nil {
		callID = *message.ParentToolUseID
	} else if len(p.pending) > 0 {
		callID, p.pending = p.pending[0], p.pending[1:]
	}
	p.emit(gimble.Event{Kind: "tool_result", CallID: callID, Data: encode(message.ToolUseResult)})
}

func (p *projector) usage(usage any) {
	if raw := encode(usage); raw != nil && string(raw) != "null" {
		p.emit(gimble.Event{Kind: "usage", Data: raw})
	}
}

func encode(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}
