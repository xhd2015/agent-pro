package agent_trace

import (
	"encoding/json"
	"strings"
	"time"
)

type traceEvent struct {
	Type     string                     `json:"type"`
	Subtype  string                     `json:"subtype,omitempty"`
	CallID   string                     `json:"call_id,omitempty"`
	Message  *traceMessage              `json:"message,omitempty"`
	Result   string                     `json:"result,omitempty"`
	Item     *traceItem                 `json:"item,omitempty"`
	ToolCall map[string]json.RawMessage `json:"tool_call,omitempty"`
	Delta    string                     `json:"delta,omitempty"`
	Text     string                     `json:"text,omitempty"`
}

type traceMessage struct {
	Content []traceContent `json:"content"`
}

type traceContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type traceItem struct {
	ID               string          `json:"id,omitempty"`
	Type             string          `json:"type,omitempty"`
	Text             string          `json:"text,omitempty"`
	Message          string          `json:"message,omitempty"`
	Content          []traceContent  `json:"content,omitempty"`
	Command          string          `json:"command,omitempty"`
	AggregatedOutput string          `json:"aggregated_output,omitempty"`
	ExitCode         *int            `json:"exit_code,omitempty"`
	Status           string          `json:"status,omitempty"`
	Items            []traceTodoItem `json:"items,omitempty"`
	Plan             []tracePlanItem `json:"plan,omitempty"`
	Explanation      string          `json:"explanation,omitempty"`
	Changes          []FileChange    `json:"changes,omitempty"`
	Raw              json.RawMessage `json:"-"`
}

type traceTodoItem struct {
	Text      string `json:"text,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Status    string `json:"status,omitempty"`
}

type tracePlanItem struct {
	Step   string `json:"step,omitempty"`
	Status string `json:"status,omitempty"`
}

func (i *traceItem) UnmarshalJSON(data []byte) error {
	type alias traceItem
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*i = traceItem(parsed)
	i.Raw = append(i.Raw[:0], data...)
	return nil
}

func ParseMessages(lines []string, createdAt string) []Message {
	baseMs := parseTraceCreatedAtMs(createdAt)
	messages := make([]Message, 0, len(lines))
	for i, line := range lines {
		parsed, ok := parseAgentTraceLine(json.RawMessage(line))
		if !ok {
			continue
		}
		ts := baseMs + int64(i)
		if parsed.Message != nil {
			msg := *parsed.Message
			if msg.StartedAt == nil {
				msg.StartedAt = &ts
			}
			messages = append(messages, msg)
			continue
		}
		if parsed.Activity != nil {
			appendTraceToolMessage(&messages, *parsed.Activity, ts)
		}
	}
	return messages
}

func parseTraceCreatedAtMs(createdAt string) int64 {
	if t, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
		return t.UnixMilli()
	}
	return time.Now().UnixMilli()
}

func appendTraceToolMessage(messages *[]Message, event AgentTraceActivity, ts int64) {
	if event.Subtype == "started" {
		*messages = append(*messages, Message{
			Role:      "tool_call",
			Content:   "",
			ToolCall:  &event,
			StartedAt: &ts,
		})
		return
	}
	for i := len(*messages) - 1; i >= 0; i-- {
		msg := &(*messages)[i]
		if msg.Role != "tool_call" || msg.ToolCall == nil || msg.FinishedAt != nil {
			continue
		}
		existing := msg.ToolCall
		match := event.CallID != "" && existing.CallID == event.CallID
		if !match {
			match = event.CallID == "" && existing.ToolName == event.ToolName
		}
		if !match {
			continue
		}
		existing.Subtype = event.Subtype
		if event.Kind != "" {
			existing.Kind = event.Kind
		}
		if event.Status != "" {
			existing.Status = event.Status
		}
		if len(event.FileChanges) > 0 {
			existing.FileChanges = event.FileChanges
		}
		if event.Summary != "" {
			if event.ReplaceSummary {
				existing.Summary = event.Summary
			} else if existing.Summary != "" {
				existing.Summary += "\n" + event.Summary
			} else {
				existing.Summary = event.Summary
			}
		}
		if event.Subtype == "completed" {
			msg.FinishedAt = &ts
		}
		return
	}
	finishedAt := (*int64)(nil)
	if event.Subtype == "completed" {
		finishedAt = &ts
	}
	*messages = append(*messages, Message{
		Role:       "tool_call",
		Content:    "",
		ToolCall:   &event,
		StartedAt:  &ts,
		FinishedAt: finishedAt,
	})
}

func traceIsAssistantItem(itemType string) bool {
	switch itemType {
	case "", "agent_message", "message", "assistant_message", "output_text":
		return true
	default:
		return false
	}
}

func traceItemText(item *traceItem) string {
	if item == nil {
		return ""
	}
	if strings.TrimSpace(item.Text) != "" {
		return item.Text
	}
	if strings.TrimSpace(item.Message) != "" {
		return item.Message
	}
	var b strings.Builder
	for _, part := range item.Content {
		if part.Text != "" {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func traceMessageText(message *traceMessage) string {
	if message == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range message.Content {
		if part.Type == "" || part.Type == "text" {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}
