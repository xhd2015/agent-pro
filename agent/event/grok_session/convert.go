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
	pendingUserTS    int64 // wire ms of first chunk of coalesced user message; 0 if unknown
	pendingUserTSSet bool  // true once first non-empty user chunk was seen (even if ts unknown)
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
	// Enrich nested-envelope lines: timestamps may sit on the outer object.
	if wireTS := sessionUpdateTimestampMs(upd); wireTS == 0 {
		if ms := extractTimestampFromRawLine(line); ms > 0 {
			v := ms
			upd.AgentTimestampMs = &v
		}
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
		// First chunk of a coalesced user message wins for timestamp.
		if !c.pendingUserTSSet {
			c.pendingUserTS = sessionUpdateTimestampMs(upd)
			c.pendingUserTSSet = true
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
		// Buffer like user/think chunks; flush on tool/turn/Flush so multi-chunk
		// assistant text coalesces (e.g. "Hello " + "world" -> "Hello world").
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
	ts := c.pendingUserTS
	c.pendingUserTS = 0
	c.pendingUserTSSet = false
	// Historical user messages: use wire timestamp when present; keep 0 when
	// unknown (do not stamp convert-time Now — callers format zero as [—]).
	ev := c.withGrokSession(types.AgentEvent{
		Type:      types.ActionMessage,
		Role:      "user",
		Text:      text,
		Timestamp: ts,
	}, "")
	if ts == 0 {
		ev.Timestamp = 0
	}
	return []types.AgentEvent{ev}
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

// sessionUpdateTimestampMs prefers agentTimestampMs / _meta ms, else top-level timestamp.
func sessionUpdateTimestampMs(upd SessionUpdate) int64 {
	if upd.AgentTimestampMs != nil && *upd.AgentTimestampMs > 0 {
		return normalizeUnixMs(*upd.AgentTimestampMs)
	}
	if len(upd.Meta) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(upd.Meta, &meta); err == nil {
			for _, key := range []string{"agentTimestampMs", "timestampMs", "timestamp_ms", "timestamp"} {
				if v, ok := meta[key]; ok {
					if ms := anyToUnixMs(v); ms > 0 {
						return ms
					}
				}
			}
		}
	}
	if len(upd.Timestamp) > 0 {
		return rawTimestampToMs(upd.Timestamp)
	}
	return 0
}

// extractTimestampFromRawLine scans a full JSONL line (flat or nested envelope)
// for timestamp / agentTimestampMs fields.
func extractTimestampFromRawLine(line string) int64 {
	var root map[string]any
	if err := json.Unmarshal([]byte(line), &root); err != nil {
		return 0
	}
	if ms := mapTimestampMs(root); ms > 0 {
		return ms
	}
	// Nested wire envelope: method + params.update
	if params, ok := root["params"].(map[string]any); ok {
		if ms := mapTimestampMs(params); ms > 0 {
			return ms
		}
		if upd, ok := params["update"].(map[string]any); ok {
			return mapTimestampMs(upd)
		}
	}
	return 0
}

func mapTimestampMs(m map[string]any) int64 {
	if m == nil {
		return 0
	}
	if v, ok := m["agentTimestampMs"]; ok {
		if ms := anyToUnixMs(v); ms > 0 {
			return ms
		}
	}
	if meta, ok := m["_meta"].(map[string]any); ok {
		for _, key := range []string{"agentTimestampMs", "timestampMs", "timestamp_ms", "timestamp"} {
			if v, ok := meta[key]; ok {
				if ms := anyToUnixMs(v); ms > 0 {
					return ms
				}
			}
		}
	}
	if v, ok := m["timestamp"]; ok {
		return anyToUnixMs(v)
	}
	return 0
}

func rawTimestampToMs(raw json.RawMessage) int64 {
	if len(raw) == 0 {
		return 0
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		return normalizeUnixMsFloat(n)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.TrimSpace(s)
		if s == "" {
			return 0
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UnixMilli()
			}
		}
	}
	return 0
}

func anyToUnixMs(v any) int64 {
	switch x := v.(type) {
	case float64:
		return normalizeUnixMsFloat(x)
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return 0
		}
		return normalizeUnixMsFloat(f)
	case int64:
		return normalizeUnixMs(x)
	case int:
		return normalizeUnixMs(int64(x))
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UnixMilli()
			}
		}
		var f float64
		if err := json.Unmarshal([]byte(s), &f); err == nil {
			return normalizeUnixMsFloat(f)
		}
	}
	return 0
}

// normalizeUnixMs treats values below 1e12 as seconds (common wire ambiguity).
func normalizeUnixMs(v int64) int64 {
	if v <= 0 {
		return 0
	}
	if v < 1_000_000_000_000 {
		return v * 1000
	}
	return v
}

func normalizeUnixMsFloat(v float64) int64 {
	if v <= 0 {
		return 0
	}
	if v < 1e12 {
		return int64(v * 1000)
	}
	return int64(v)
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