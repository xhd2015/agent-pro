package crush_types

import (
	"encoding/json"
	"fmt"
)

type EventType string

const (
	EventMessage     EventType = "message"
	EventAgentEvent  EventType = "agent_event"
	EventRunComplete EventType = "run_complete"
)

type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type PartType string

const (
	PartReasoning  PartType = "reasoning"
	PartText       PartType = "text"
	PartToolCall   PartType = "tool_call"
	PartToolResult PartType = "tool_result"
	PartFinish     PartType = "finish"
)

type Part struct {
	Type PartType        `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MessagePayload struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	Parts     []Part `json:"parts"`
	Model     string `json:"model,omitempty"`
	Provider  string `json:"provider,omitempty"`
}

type ReasoningData struct {
	Thinking  string `json:"thinking"`
	StartedAt int64  `json:"started_at,omitempty"`
}

type TextData struct {
	Text string `json:"text"`
}

type ToolCallData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Input    string `json:"input"`
	Finished bool   `json:"finished,omitempty"`
}

type ToolResultData struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
}

type FinishReason string

const (
	FinishReasonEndTurn FinishReason = "end_turn"
	FinishReasonError   FinishReason = "error"
)

type FinishData struct {
	Reason  FinishReason `json:"reason"`
	Time    int64        `json:"time"`
	Message string       `json:"message,omitempty"`
}

type AgentEventPayload struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
	RunID string `json:"run_id,omitempty"`
}

type RunCompletePayload struct {
	SessionID string `json:"session_id,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	Text      string `json:"text,omitempty"`
	Error     string `json:"error,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal: %v", err))
	}
	return data
}

func (e Event) messagePayload() *MessagePayload {
	var p MessagePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil
	}
	return &p
}

func (e Event) agentEventPayload() *AgentEventPayload {
	var p AgentEventPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil
	}
	return &p
}

func (e Event) runCompletePayload() *RunCompletePayload {
	var p RunCompletePayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		return nil
	}
	return &p
}

func (p Part) reasoningData() *ReasoningData {
	var d ReasoningData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func (p Part) textData() *TextData {
	var d TextData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func (p Part) toolCallData() *ToolCallData {
	var d ToolCallData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func (p Part) finishData() *FinishData {
	var d FinishData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func (p Part) parseDataMap() (map[string]any, bool) {
	var m map[string]any
	if err := json.Unmarshal(p.Data, &m); err != nil {
		return nil, false
	}
	return m, true
}
