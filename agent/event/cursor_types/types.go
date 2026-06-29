package cursor_types

import "encoding/json"

type TraceEvent struct {
	Type     string                     `json:"type"`
	Subtype  string                     `json:"subtype,omitempty"`
	CallID   string                     `json:"call_id,omitempty"`
	ToolCall map[string]json.RawMessage `json:"tool_call,omitempty"`
}