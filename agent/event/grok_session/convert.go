package grok_session

import (
	"encoding/json"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type toolCallMeta struct {
	kind  string
	title string
}

// Converter converts grok session updates to AgentEvents with chunk coalescing.
type Converter struct {
	pendingUser      strings.Builder
	pendingThink     strings.Builder
	pendingAssistant strings.Builder
	toolMeta         map[string]toolCallMeta
	turnIndex        int
}

// NewConverter creates a converter for one grok session stream.
func NewConverter() *Converter {
	return &Converter{
		toolMeta:  make(map[string]toolCallMeta),
		turnIndex: 0,
	}
}

// SetTurnIndex restores the converter turn counter from a grok-sync checkpoint.
func (c *Converter) SetTurnIndex(idx int) {
	c.turnIndex = idx
}

// TurnIndex returns the current grok turn index stamped on emitted events.
func (c *Converter) TurnIndex() int {
	return c.turnIndex
}

// ProcessLine parses one updates.jsonl line and returns AgentEvents to emit.
func (c *Converter) ProcessLine(line string) []types.AgentEvent {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	upd, ok := ParseLine(line)
	if !ok {
		return nil
	}
	return c.processUpdate(upd)
}

// Flush emits any buffered chunk coalescence at end of stream.
func (c *Converter) Flush() []types.AgentEvent {
	var out []types.AgentEvent
	out = append(out, c.flushUser()...)
	out = append(out, c.flushThink()...)
	out = append(out, c.flushAssistant()...)
	return out
}

// FromUpdatesJSONL walks wire lines through a converter, flushes at end, and
// returns all canonical events.
func FromUpdatesJSONL(lines []string) []types.AgentEvent {
	c := NewConverter()
	var out []types.AgentEvent
	for _, line := range lines {
		out = append(out, c.ProcessLine(line)...)
	}
	out = append(out, c.Flush()...)
	return out
}

func (c *Converter) processUpdate(upd SessionUpdate) []types.AgentEvent {
	switch upd.SessionUpdate {
	case "user_message_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		text := TextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingUser.WriteString(text)
		return out
	case "agent_thought_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushAssistant()...)
		text := TextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingThink.WriteString(text)
		return out
	case "agent_message_chunk":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		text := TextContent(upd.Content)
		if text == "" {
			return out
		}
		c.pendingAssistant.WriteString(text)
		out = append(out, c.flushAssistant()...)
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
		out = append(out, c.withGrokSession(types.AgentEvent{
			Type:       types.ActionToolCall,
			Tool:       normalizeToolKind(kind),
			Text:       title,
			ToolCallID: id,
		}, "pending"))
		return out
	case "tool_call_update":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		id := strings.TrimSpace(upd.ToolCallID)
		meta := c.toolMeta[id]
		status := strings.TrimSpace(upd.Status)
		if status == "" {
			status = "completed"
		}
		out = append(out, c.withGrokSession(types.AgentEvent{
			Type:       types.ActionToolCall,
			Tool:       normalizeToolKind(meta.kind),
			Text:       meta.title,
			ToolCallID: id,
			Output:     toolOutput(upd.Content),
		}, status))
		return out
	case "turn_completed":
		var out []types.AgentEvent
		out = append(out, c.flushUser()...)
		out = append(out, c.flushThink()...)
		out = append(out, c.flushAssistant()...)
		out = append(out, c.withGrokSession(types.AgentEvent{
			Type: types.ActionDone,
		}, ""))
		c.turnIndex++
		return out
	default:
		return nil
	}
}

func (c *Converter) flushUser() []types.AgentEvent {
	if c.pendingUser.Len() == 0 {
		return nil
	}
	text := c.pendingUser.String()
	c.pendingUser.Reset()
	return []types.AgentEvent{c.withGrokSession(types.AgentEvent{
		Type:      types.ActionMessage,
		Role:      "user",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}, "")}
}

func (c *Converter) flushThink() []types.AgentEvent {
	if c.pendingThink.Len() == 0 {
		return nil
	}
	text := c.pendingThink.String()
	c.pendingThink.Reset()
	return []types.AgentEvent{c.withGrokSession(types.AgentEvent{
		Type:      types.ActionThink,
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}, "")}
}

func (c *Converter) flushAssistant() []types.AgentEvent {
	if c.pendingAssistant.Len() == 0 {
		return nil
	}
	text := c.pendingAssistant.String()
	c.pendingAssistant.Reset()
	return []types.AgentEvent{c.withGrokSession(types.AgentEvent{
		Type:      types.ActionMessage,
		Role:      "assistant",
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}, "")}
}

func (c *Converter) withGrokSession(ev types.AgentEvent, status string) types.AgentEvent {
	ext := &types.GrokSessionExtension{TurnIndex: c.turnIndex}
	if status != "" {
		ext.Status = status
	}
	ev.Extensions = &types.EventExtensions{GrokSession: ext}
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().UnixMilli()
	}
	return ev
}

func toolOutput(raw json.RawMessage) string {
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