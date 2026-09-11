package codex

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/tylergannon/gimble"
)

// projector turns Codex notifications into Gimble events.
type projector struct {
	mu      sync.Mutex
	emit    func(gimble.AgentEvent)
	started map[string]bool // tool items whose call was emitted
}

func (p *projector) user(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.emit(gimble.UserMessage{Text: text})
}

func (p *projector) usage(params json.RawMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var notification struct {
		TokenUsage json.RawMessage `json:"tokenUsage"`
	}
	if json.Unmarshal(params, &notification) == nil && len(notification.TokenUsage) > 0 {
		p.emit(gimble.Usage{Data: gimble.JSONText(notification.TokenUsage)})
	}
}

func (p *projector) itemStarted(params json.RawMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if item, ok := decodeItem(params); ok && isTool(item.kind) {
		p.toolCall(item)
	}
}

// itemCompleted emits the item and returns the text of an agent message.
func (p *projector) itemCompleted(params json.RawMessage) (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	item, ok := decodeItem(params)
	if !ok {
		return "", false
	}
	switch {
	case item.kind == "agentMessage":
		text, _ := item.value["text"].(string)
		p.emit(gimble.AssistantMessage{ID: item.id, Text: text})
		return text, true
	case item.kind == "reasoning":
		text := joined(item.value["summary"])
		if text == "" {
			text = joined(item.value["content"])
		}
		if text != "" {
			p.emit(gimble.Thinking{ID: item.id, Text: text})
		}
	case isTool(item.kind):
		if !p.started[item.id] {
			p.toolCall(item)
		}
		p.emit(gimble.ToolResult{CallID: item.id, Output: gimble.JSONText(encode(toolOutput(item)))})
	}
	return "", false
}

func (p *projector) toolCall(item nativeItem) {
	p.started[item.id] = true
	p.emit(gimble.ToolCall{CallID: item.id, Tool: toolName(item), Input: gimble.JSONText(encode(toolArgs(item)))})
}

type nativeItem struct {
	id    string
	kind  string
	value map[string]any
}

func decodeItem(params json.RawMessage) (nativeItem, bool) {
	var envelope struct {
		Item map[string]any `json:"item"`
	}
	if json.Unmarshal(params, &envelope) != nil || envelope.Item == nil {
		return nativeItem{}, false
	}
	id, _ := envelope.Item["id"].(string)
	kind, _ := envelope.Item["type"].(string)
	return nativeItem{id: id, kind: kind, value: envelope.Item}, id != "" && kind != ""
}

func isTool(kind string) bool {
	switch kind {
	case "commandExecution", "fileChange", "mcpToolCall", "dynamicToolCall",
		"collabAgentToolCall", "subAgentActivity", "webSearch", "imageView",
		"sleep", "imageGeneration":
		return true
	}
	return false
}

func toolName(item nativeItem) string {
	switch item.kind {
	case "mcpToolCall":
		return fmt.Sprintf("%v/%v", item.value["server"], item.value["tool"])
	case "dynamicToolCall":
		if namespace, _ := item.value["namespace"].(string); namespace != "" {
			return namespace + "/" + fmt.Sprint(item.value["tool"])
		}
		return fmt.Sprint(item.value["tool"])
	}
	return item.kind
}

func toolArgs(item nativeItem) any {
	switch item.kind {
	case "commandExecution":
		return map[string]any{"command": item.value["command"], "cwd": item.value["cwd"]}
	case "fileChange":
		return map[string]any{"changes": item.value["changes"]}
	case "mcpToolCall", "dynamicToolCall":
		return item.value["arguments"]
	case "webSearch":
		return map[string]any{"query": item.value["query"], "action": item.value["action"]}
	}
	return item.value
}

func toolOutput(item nativeItem) any {
	switch item.kind {
	case "commandExecution":
		return item.value["aggregatedOutput"]
	case "fileChange":
		return map[string]any{"status": item.value["status"], "changes": item.value["changes"]}
	case "mcpToolCall":
		if item.value["error"] != nil {
			return item.value["error"]
		}
		return item.value["result"]
	case "dynamicToolCall":
		return map[string]any{"success": item.value["success"], "contentItems": item.value["contentItems"]}
	case "webSearch":
		return item.value["results"]
	}
	return item.value
}

func encode(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}

func joined(value any) string {
	values, _ := value.([]any)
	var parts []string
	for _, v := range values {
		if text, ok := v.(string); ok && strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}
