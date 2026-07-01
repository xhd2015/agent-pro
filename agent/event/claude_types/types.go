package claude_types

import (
	"encoding/json"
)

// EventType discriminates the top-level stream-json line type emitted by
// Claude Code's headless mode.
type EventType string

const (
	EventSystem    EventType = "system"
	EventAssistant EventType = "assistant"
	EventUser      EventType = "user"
	EventResult    EventType = "result"
)

// StreamEvent models one NDJSON line from `claude -p --output-format
// stream-json --verbose`.
type StreamEvent struct {
	Type       EventType       `json:"type"`
	Subtype    string          `json:"subtype,omitempty"`
	Message    *Message        `json:"message,omitempty"`
	Result     string          `json:"result,omitempty"`
	IsError    bool            `json:"is_error"` // NO omitempty — tests assert "is_error":false
	SessionID  string          `json:"session_id,omitempty"`
	DurationMs int64           `json:"duration_ms,omitempty"`
	NumTurns   int             `json:"num_turns,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	Model      string          `json:"model,omitempty"`
	Cwd        string          `json:"cwd,omitempty"`
	Raw        json.RawMessage `json:"-"` // optional, for round-trip
}

// Message is the Anthropic-style message envelope carried on assistant/user
// events.
type Message struct {
	Role    string         `json:"role,omitempty"`
	Content []ContentBlock `json:"content,omitempty"`
}

// ContentBlock is one entry in a Message.Content array.
type ContentBlock struct {
	Type      string         `json:"type"` // "text" | "thinking" | "tool_use" | "tool_result"
	Text      string         `json:"text,omitempty"`
	Thinking  string         `json:"thinking,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Input     map[string]any `json:"input,omitempty"`
	ToolUseID string         `json:"tool_use_id,omitempty"`
	Content   string         `json:"content,omitempty"` // tool_result content (string form)
	IsError   bool           `json:"is_error,omitempty"`
}
