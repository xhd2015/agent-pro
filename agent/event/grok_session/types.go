package grok_session

import "encoding/json"

// SessionUpdate is one ACP session update record from updates.jsonl.
type SessionUpdate struct {
	SessionUpdate string          `json:"sessionUpdate"`
	Content       json.RawMessage `json:"content,omitempty"`
	ToolCallID    string          `json:"toolCallId,omitempty"`
	Kind          string          `json:"kind,omitempty"`
	Title         string          `json:"title,omitempty"`
	Status        string          `json:"status,omitempty"`
	// RawInput is the tool argument object (e.g. {"command":"..."} for shell tools).
	RawInput json.RawMessage `json:"rawInput,omitempty"`
	// Optional wire timing. Prefer AgentTimestampMs / _meta ms; else Timestamp.
	Timestamp        json.RawMessage `json:"timestamp,omitempty"`
	AgentTimestampMs *int64          `json:"agentTimestampMs,omitempty"`
	Meta             json.RawMessage `json:"_meta,omitempty"`
}

// ToOptions controls reverse conversion wire shape.
// Zero value emits flat sessionUpdate objects.
type ToOptions struct {
	SessionID string
}