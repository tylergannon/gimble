package agy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/tylergannon/gimble"
)

type envelope struct {
	Event          string      `json:"event"`
	ConversationID string      `json:"conversation_id"`
	Init           *initEvent  `json:"init"`
	StepUpdate     *stepUpdate `json:"step_update"`
	Result         *result     `json:"result"`
}

type initEvent struct {
	ConversationID string `json:"conversation_id"`
}

type stepUpdate struct {
	ConversationID string         `json:"conversation_id"`
	StepIndex      int            `json:"step_index"`
	State          string         `json:"state"`
	StepType       string         `json:"step_type"`
	TextDelta      string         `json:"text_delta"`
	ToolInfo       *toolInfo      `json:"tool_info"`
	Usage          map[string]any `json:"usage"`
}

type toolInfo struct {
	Name       string `json:"name"`
	Parameters any    `json:"parameters"`
	Output     any    `json:"output"`
	Error      any    `json:"error"`
}

type result struct {
	ConversationID   string         `json:"conversation_id"`
	Status           string         `json:"status"`
	Response         string         `json:"response"`
	Error            string         `json:"error"`
	StructuredOutput any            `json:"structured_output"`
	Usage            map[string]any `json:"usage"`
}

func (e envelope) conversationID() string {
	if e.ConversationID != "" {
		return e.ConversationID
	}
	if e.Init != nil && e.Init.ConversationID != "" {
		return e.Init.ConversationID
	}
	if e.Result != nil {
		return e.Result.ConversationID
	}
	return ""
}

type projector struct {
	mu        sync.Mutex
	sessionID string
	model     string
	emit      func(gimble.AgentEvent) error
	steps     map[int]*projectedStep
}

type projectedStep struct {
	kind, messageID, itemID, name string
	text                          strings.Builder
	textOpen, toolCalled, ended   bool
}

func newProjector(sessionID, model string, emit func(gimble.AgentEvent) error) *projector {
	if emit == nil {
		emit = func(gimble.AgentEvent) error { return nil }
	}
	return &projector{sessionID: sessionID, model: model, emit: emit, steps: make(map[int]*projectedStep)}
}

func (p *projector) envelope(envelope envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if envelope.StepUpdate == nil {
		return nil
	}
	update := envelope.StepUpdate
	if update.ConversationID != "" && update.ConversationID != p.sessionID {
		return fmt.Errorf("agy: step returned conversation %q, want %q", update.ConversationID, p.sessionID)
	}
	switch update.StepType {
	case "user_input":
		return nil
	case "agent_response", "tool":
		return p.project(update)
	default:
		return nil
	}
}

func (p *projector) project(update *stepUpdate) error {
	step := p.steps[update.StepIndex]
	if step == nil {
		step = &projectedStep{
			kind:      update.StepType,
			messageID: fmt.Sprintf("%s/step.%d", p.sessionID, update.StepIndex),
			itemID:    fmt.Sprintf("%s/item.%d", p.sessionID, update.StepIndex),
		}
		p.steps[update.StepIndex] = step
		if err := p.event("session.step.started", map[string]any{
			"assistantMessageID": step.messageID, "agent": "agy",
			"model": map[string]any{"providerID": "google", "id": p.model},
		}, p.ref(step, update.Usage)); err != nil {
			return err
		}
	} else if step.kind != update.StepType {
		return fmt.Errorf("agy: step %d changed type from %q to %q", update.StepIndex, step.kind, update.StepType)
	}
	if step.ended {
		return fmt.Errorf("agy: update for ended step %d", update.StepIndex)
	}

	if update.StepType == "agent_response" {
		if update.TextDelta != "" {
			if !step.textOpen {
				step.textOpen = true
				if err := p.event("session.text.started", map[string]any{"assistantMessageID": step.messageID, "ordinal": 0}, p.ref(step, nil)); err != nil {
					return err
				}
			}
			step.text.WriteString(update.TextDelta)
			if err := p.event("session.text.delta", map[string]any{"assistantMessageID": step.messageID, "ordinal": 0, "delta": update.TextDelta}, p.ref(step, nil)); err != nil {
				return err
			}
		}
	} else if err := p.projectTool(step, update); err != nil {
		return err
	}

	if update.State != "DONE" {
		return nil
	}
	if step.kind == "agent_response" && step.textOpen {
		if err := p.event("session.text.ended", map[string]any{"assistantMessageID": step.messageID, "ordinal": 0, "text": step.text.String()}, p.ref(step, nil)); err != nil {
			return err
		}
		step.textOpen = false
	}
	if err := p.event("session.step.streamed", map[string]any{"assistantMessageID": step.messageID}, p.ref(step, nil)); err != nil {
		return err
	}
	usage := normalizeUsage(update.Usage)
	finish := "stop"
	if step.kind == "tool" || step.text.Len() == 0 {
		finish = "tool-calls"
	}
	if err := p.event("session.step.ended", map[string]any{
		"assistantMessageID": step.messageID, "finish": finish, "cost": 0,
		"tokens": map[string]any{
			"input": usage.input, "output": usage.output, "reasoning": usage.reasoning,
			"cache": map[string]any{"read": usage.cacheRead, "write": usage.cacheWrite},
		},
	}, p.ref(step, update.Usage)); err != nil {
		return err
	}
	step.ended = true
	return nil
}

func (p *projector) projectTool(step *projectedStep, update *stepUpdate) error {
	if update.ToolInfo == nil {
		if update.State == "DONE" {
			return errors.New("agy: completed tool step omitted tool_info")
		}
		return nil
	}
	info := update.ToolInfo
	if info.Name != "" {
		step.name = info.Name
	}
	if step.name == "" {
		return errors.New("agy: tool step omitted name")
	}
	if !step.toolCalled {
		step.toolCalled = true
		ref := p.ref(step, nil)
		input := objectValue(info.Parameters)
		inputRaw, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("agy: encode tool input: %w", err)
		}
		if err := p.event("session.tool.input.started", map[string]any{"assistantMessageID": step.messageID, "id": step.itemID, "name": step.name}, ref); err != nil {
			return err
		}
		if err := p.event("session.tool.input.ended", map[string]any{"assistantMessageID": step.messageID, "id": step.itemID, "text": string(inputRaw)}, ref); err != nil {
			return err
		}
		if err := p.event("session.tool.called", map[string]any{"assistantMessageID": step.messageID, "id": step.itemID, "input": input, "executed": true}, ref); err != nil {
			return err
		}
	}
	if update.State != "DONE" {
		return nil
	}
	if info.Error != nil {
		return p.event("session.tool.failed", map[string]any{
			"assistantMessageID": step.messageID, "id": step.itemID, "executed": true,
			"error": map[string]any{"type": "tool", "message": outputText(info.Error)},
		}, p.ref(step, nil))
	}
	return p.event("session.tool.success", map[string]any{
		"assistantMessageID": step.messageID, "id": step.itemID, "executed": true,
		"content": []any{map[string]any{"type": "text", "text": outputText(info.Output)}},
	}, p.ref(step, nil))
}

func (p *projector) event(eventType string, data map[string]any, nativeRef map[string]any) error {
	data["sessionID"] = p.sessionID
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ref, err := json.Marshal(nativeRef)
	if err != nil {
		return err
	}
	return p.emit(gimble.AgentEvent{Type: eventType, Data: raw, NativeRef: ref})
}

func (p *projector) ref(step *projectedStep, rawUsage map[string]any) map[string]any {
	ref := map[string]any{
		"provider": "agy", "sessionID": p.sessionID, "messageID": step.messageID,
	}
	if step.kind == "tool" {
		ref["itemID"] = step.itemID
	}
	if rawUsage != nil {
		usage := normalizeUsage(rawUsage)
		ref["accounting"] = map[string]any{
			"tokensAvailable": usage.available, "costAvailable": false, "costSource": "unavailable",
			"fieldAvailability": usage.fields, "rawProviderAccounting": rawUsage,
		}
	}
	return ref
}

type normalizedUsage struct {
	input, output, reasoning, cacheRead, cacheWrite float64
	available                                       bool
	fields                                          map[string]bool
}

func normalizeUsage(value map[string]any) normalizedUsage {
	read := func(key string) (float64, bool) {
		number, ok := value[key].(float64)
		return number, ok
	}
	input, inputOK := read("input_tokens")
	outputTotal, outputOK := read("output_tokens")
	reasoning, reasoningOK := read("thinking_tokens")
	cacheRead, cacheReadOK := read("cache_read_tokens")
	cacheWrite, cacheWriteOK := read("cache_write_tokens")
	fields := map[string]bool{
		"input": inputOK, "output": outputOK && reasoningOK, "reasoning": reasoningOK,
		"cacheRead": cacheReadOK, "cacheWrite": cacheWriteOK,
	}
	return normalizedUsage{
		input: input, output: max(0, outputTotal-reasoning), reasoning: reasoning,
		cacheRead: cacheRead, cacheWrite: cacheWrite,
		available: inputOK && outputOK && reasoningOK && cacheReadOK && cacheWriteOK,
		fields:    fields,
	}
}

func objectValue(value any) map[string]any {
	if object, ok := value.(map[string]any); ok && object != nil {
		return object
	}
	if value == nil {
		return map[string]any{}
	}
	return map[string]any{"value": value}
}

func outputText(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(raw)
}
