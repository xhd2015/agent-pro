package grok_types

type EventType = string

const (
	EventText    EventType = "text"
	EventThought EventType = "thought"
	EventEnd     EventType = "end"
)

type Event struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	IsError   bool   `json:"isError,omitempty"`
}
