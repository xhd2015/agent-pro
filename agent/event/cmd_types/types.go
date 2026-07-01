package cmd_types

import "encoding/json"

// Role discriminates the role of a cmd session event.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ContentBlockType discriminates the type of a content block.
type ContentBlockType string

const (
	BlockTypeText      ContentBlockType = "text"
	BlockTypeReasoning ContentBlockType = "reasoning"
	BlockTypeToolCall  ContentBlockType = "tool-call"
	BlockTypeToolResult ContentBlockType = "tool-result"
)

// Event models one JSONL line from a cmd session file.
type Event struct {
	ID        string          `json:"id"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	ParentID  string          `json:"parentId"`
	Role      Role            `json:"role"`
	Content   json.RawMessage `json:"content"` // string or []ContentBlock
	GitBranch string          `json:"gitBranch,omitempty"`
	Metadata  *EventMetadata  `json:"metadata,omitempty"`
}

// EventMetadata contains optional metadata fields.
type EventMetadata struct {
	Timestamp   string `json:"timestamp,omitempty"`
	Source      string `json:"source,omitempty"`
	Version     int    `json:"version,omitempty"`
	Entrypoint  string `json:"entrypoint,omitempty"`
	MessageID   string `json:"messageId,omitempty"`
	IsAutomated bool   `json:"isAutomated,omitempty"`
}

// ContentBlock is one element in a content array.
type ContentBlock struct {
	Type       ContentBlockType `json:"type"`
	Text       string           `json:"text,omitempty"`
	ToolCallID string           `json:"toolCallId,omitempty"`
	ToolName   string           `json:"toolName,omitempty"`
	Input      json.RawMessage  `json:"input,omitempty"`
	Output     *ToolOutput      `json:"output,omitempty"`
}

// ToolOutput wraps the result of a tool execution.
type ToolOutput struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
