package pi_types

import "encoding/json"

type EventType string

const (
	EventTypeSession          EventType = "session"
	EventTypeAgentStart       EventType = "agent_start"
	EventTypeAgentEnd         EventType = "agent_end"
	EventTypeTurnStart        EventType = "turn_start"
	EventTypeTurnEnd          EventType = "turn_end"
	EventTypeMessageStart     EventType = "message_start"
	EventTypeMessageUpdate    EventType = "message_update"
	EventTypeMessageEnd       EventType = "message_end"
	EventTypeToolExecStart    EventType = "tool_execution_start"
	EventTypeToolExecUpdate   EventType = "tool_execution_update"
	EventTypeToolExecEnd      EventType = "tool_execution_end"
)

type StopReason string

const (
	StopReasonStop    StopReason = "stop"
	StopReasonLength  StopReason = "length"
	StopReasonToolUse StopReason = "toolUse"
	StopReasonError   StopReason = "error"
	StopReasonAborted StopReason = "aborted"
)

type Event struct {
	Type                  EventType                `json:"type"`
	ID                    string                   `json:"id,omitempty"`
	Message               *AgentMessage            `json:"message,omitempty"`
	Messages              []AgentMessage           `json:"messages,omitempty"`
	AssistantMessageEvent *AssistantMessageEvent   `json:"assistantMessageEvent,omitempty"`
	ToolResults           []ToolResultMessage      `json:"toolResults,omitempty"`
	ToolCallID            string                   `json:"toolCallId,omitempty"`
	ToolName              string                   `json:"toolName,omitempty"`
	Args                  map[string]any           `json:"args,omitempty"`
	Result                any                      `json:"result,omitempty"`
	IsError               bool                     `json:"isError"`
	PartialResult         any                      `json:"partialResult,omitempty"`
}

type AgentMessage struct {
	Role           string             `json:"role"`
	Content        MessageContentList `json:"content"`
	API            string             `json:"api,omitempty"`
	Provider       string             `json:"provider,omitempty"`
	Model          string             `json:"model,omitempty"`
	ResponseModel  string             `json:"responseModel,omitempty"`
	ResponseID     string             `json:"responseId,omitempty"`
	SessionID      string             `json:"session_id,omitempty"`
	Diagnostics    any                `json:"diagnostics,omitempty"`
	Usage          *Usage             `json:"usage,omitempty"`
	StopReason     StopReason         `json:"stopReason,omitempty"`
	ErrorMessage   string             `json:"errorMessage,omitempty"`
	Timestamp      int64              `json:"timestamp,omitempty"`
	Details        any                `json:"details,omitempty"`
}

type MessageContentList []MessageContent

func (l *MessageContentList) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*l = []MessageContent{{Type: "text", Text: s}}
		return nil
	}
	return json.Unmarshal(data, (*[]MessageContent)(l))
}

type MessageContent struct {
	Type      string         `json:"type"`
	Text      string         `json:"text,omitempty"`
	Thinking  string         `json:"thinking,omitempty"`
	Redacted  bool           `json:"redacted,omitempty"`
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Data      string         `json:"data,omitempty"`
	MimeType  string         `json:"mimeType,omitempty"`
	ToolCallID string        `json:"toolCallId,omitempty"`
	ToolName  string         `json:"toolName,omitempty"`
	IsError   bool           `json:"isError,omitempty"`
	Content   MessageContentList `json:"content,omitempty"`
}

type AssistantMessageEvent struct {
	Type         string        `json:"type"`
	ContentIndex int           `json:"contentIndex,omitempty"`
	Delta        string        `json:"delta,omitempty"`
	Content      string        `json:"content,omitempty"`
	ToolCall     *MessageContent `json:"toolCall,omitempty"`
	Partial      *AgentMessage `json:"partial,omitempty"`
	Reason       string        `json:"reason,omitempty"`
	Error        string        `json:"error,omitempty"`
}

type ToolResultMessage struct {
	Role       string             `json:"role"`
	ToolCallID string             `json:"toolCallId"`
	ToolName   string             `json:"toolName"`
	Content    MessageContentList `json:"content,omitempty"`
	IsError    bool               `json:"isError"`
	Timestamp  int64              `json:"timestamp,omitempty"`
}

type Usage struct {
	Input       int   `json:"input"`
	Output      int   `json:"output"`
	CacheRead   int   `json:"cacheRead"`
	CacheWrite  int   `json:"cacheWrite"`
	TotalTokens int   `json:"totalTokens"`
	Cost        *Cost `json:"cost,omitempty"`
}

type Cost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}
