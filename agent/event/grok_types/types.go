package grok_types

type EventType = string

const (
	EventText          EventType = "text"
	EventThought       EventType = "thought"
	EventEnd           EventType = "end"
	EventToolStarted   EventType = "tool_started"
	EventToolCompleted EventType = "tool_completed"
)

type Event struct {
	Type       string `json:"type"`
	Data       string `json:"data,omitempty"`
	SessionID  string `json:"sessionId,omitempty"`
	IsError    bool   `json:"isError,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	Outcome    string `json:"outcome,omitempty"`
	DurationMs int    `json:"duration_ms,omitempty"`
}
