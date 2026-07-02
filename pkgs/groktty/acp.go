package groktty

import (
	"encoding/json"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type acpUpdate struct {
	SessionUpdate string          `json:"sessionUpdate"`
	Content       json.RawMessage `json:"content"`
	ToolCallID    string          `json:"toolCallId"`
	Kind          string          `json:"kind"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
}

type acpTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type acpWireEnvelope struct {
	Method string `json:"method"`
	Params struct {
		SessionID string          `json:"sessionId"`
		Update    json.RawMessage `json:"update"`
	} `json:"params"`
}

// parseACPUpdateLine parses one updates.jsonl line in either flat ACP format
// {"sessionUpdate":...} or grok wire envelope
// {"method":"session/update","params":{"sessionId":"...","update":{...}}}.
func parseACPUpdateLine(line string) (acpUpdate, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return acpUpdate{}, false
	}

	var flat acpUpdate
	if err := json.Unmarshal([]byte(line), &flat); err == nil && strings.TrimSpace(flat.SessionUpdate) != "" {
		return flat, true
	}

	var wire acpWireEnvelope
	if err := json.Unmarshal([]byte(line), &wire); err != nil {
		return acpUpdate{}, false
	}
	if len(wire.Params.Update) == 0 {
		return acpUpdate{}, false
	}
	var nested acpUpdate
	if err := json.Unmarshal(wire.Params.Update, &nested); err != nil {
		return acpUpdate{}, false
	}
	if strings.TrimSpace(nested.SessionUpdate) == "" {
		return acpUpdate{}, false
	}
	return nested, true
}

// ACPConverter converts ACP session updates to AgentEvents with chunk coalescing.
type ACPConverter struct {
	pendingUser      strings.Builder
	pendingThink     strings.Builder
	pendingAssistant strings.Builder
	toolMeta         map[string]toolCallMeta
}

type toolCallMeta struct {
	kind  string
	title string
}

// NewACPConverter creates a converter for one grok session tail.
func NewACPConverter() *ACPConverter {
	return &ACPConverter{toolMeta: make(map[string]toolCallMeta)}
}

// ProcessLine parses one updates.jsonl line and returns AgentEvents to emit.
func (c *ACPConverter) ProcessLine(line string) []types.AgentEvent {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	upd, ok := parseACPUpdateLine(line)
	if !ok {
		return nil
	}
	return c.processUpdate(upd)
}

// Flush emits any buffered chunk coalescence at end of tail.
func (c *ACPConverter) Flush() []types.AgentEvent {
	var out []types.AgentEvent
	out = append(out, c.flushUser()...)
	out = append(out, c.flushThink()...)
	out = append(out, c.flushAssistant()...)
	return out
}

func (c *ACPConverter) processUpdate(upd acpUpdate) []types.AgentEvent {
	switch upd.SessionUpdate {
	case "user_message_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		text := acpTextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingUser.WriteString(text)
		return out
	case "agent_thought_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushAssistant()...)
		text := acpTextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingThink.WriteString(text)
		return out
	case "agent_message_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		text := acpTextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingAssistant.WriteString(text)
		return out
	case "tool_call":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		id := strings.TrimSpace(upd.ToolCallID)
		kind := strings.TrimSpace(upd.Kind)
		title := strings.TrimSpace(upd.Title)
		if id != "" {
			c.toolMeta[id] = toolCallMeta{kind: kind, title: title}
		}
		out = append(out, types.AgentEvent{
			Type:      types.ActionToolCall,
			Tool:      normalizeToolKind(kind),
			Text:      title,
			Timestamp: time.Now().UnixMilli(),
		})
		return out
	case "tool_call_update":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		id := strings.TrimSpace(upd.ToolCallID)
		meta := c.toolMeta[id]
		output := acpToolOutput(upd.Content)
		out = append(out, types.AgentEvent{
			Type:      types.ActionToolCall,
			Tool:      normalizeToolKind(meta.kind),
			Text:      meta.title,
			Output:    output,
			Timestamp: time.Now().UnixMilli(),
		})
		return out
	case "turn_completed":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		return out
	default:
		return nil
	}
}

func (c *ACPConverter) flushUser() []types.AgentEvent {
	if c.pendingUser.Len() == 0 {
		return nil
	}
	text := c.pendingUser.String()
	c.pendingUser.Reset()
	return []types.AgentEvent{{
		Type:      types.ActionMessage,
		Role:      "user",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}}
}

func (c *ACPConverter) flushThink() []types.AgentEvent {
	if c.pendingThink.Len() == 0 {
		return nil
	}
	text := c.pendingThink.String()
	c.pendingThink.Reset()
	return []types.AgentEvent{{
		Type:      types.ActionThink,
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}}
}

func (c *ACPConverter) flushAssistant() []types.AgentEvent {
	if c.pendingAssistant.Len() == 0 {
		return nil
	}
	text := c.pendingAssistant.String()
	c.pendingAssistant.Reset()
	return []types.AgentEvent{{
		Type:      types.ActionMessage,
		Role:      "assistant",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}}
}

func acpTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var block acpTextBlock
	if err := json.Unmarshal(raw, &block); err == nil && block.Text != "" {
		return block.Text
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	return ""
}

func acpToolOutput(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var blocks []map[string]any
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, block := range blocks {
		if content, ok := block["content"].(map[string]any); ok {
			if text, ok := content["text"].(string); ok && strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
			continue
		}
		if text, ok := block["text"].(string); ok && strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func normalizeToolKind(kind string) string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "tool"
	}
	return strings.ToLower(kind)
}