package codex_types

import (
	"encoding/json"
	"strings"
)

type TraceEvent struct {
	Type     string                     `json:"type"`
	Subtype  string                     `json:"subtype,omitempty"`
	CallID   string                     `json:"call_id,omitempty"`
	Message  *TraceMessage              `json:"message,omitempty"`
	Result   string                     `json:"result,omitempty"`
	Item     *TraceItem                 `json:"item,omitempty"`
	ToolCall map[string]json.RawMessage `json:"tool_call,omitempty"`
	Delta    string                     `json:"delta,omitempty"`
	Text     string                     `json:"text,omitempty"`
}

type TraceMessage struct {
	Content []TraceContent `json:"content"`
}

type TraceContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type TraceItem struct {
	ID               string          `json:"id,omitempty"`
	Type             string          `json:"type,omitempty"`
	Text             string          `json:"text,omitempty"`
	Message          string          `json:"message,omitempty"`
	Content          []TraceContent  `json:"content,omitempty"`
	Command          string          `json:"command,omitempty"`
	AggregatedOutput string          `json:"aggregated_output,omitempty"`
	ExitCode         *int            `json:"exit_code,omitempty"`
	Status           string          `json:"status,omitempty"`
	Items            []TraceTodoItem `json:"items,omitempty"`
	Plan             []TracePlanItem `json:"plan,omitempty"`
	Explanation      string          `json:"explanation,omitempty"`
	Changes          []FileChange    `json:"changes,omitempty"`
	Raw              json.RawMessage `json:"-"`
}

type TraceTodoItem struct {
	Text      string `json:"text,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Status    string `json:"status,omitempty"`
}

type TracePlanItem struct {
	Step   string `json:"step,omitempty"`
	Status string `json:"status,omitempty"`
}

func (i *TraceItem) UnmarshalJSON(data []byte) error {
	type alias TraceItem
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*i = TraceItem(parsed)
	i.Raw = append(i.Raw[:0], data...)
	return nil
}

func TraceIsAssistantItem(itemType string) bool {
	switch itemType {
	case "", "agent_message", "message", "assistant_message", "output_text":
		return true
	default:
		return false
	}
}

func TraceItemText(item *TraceItem) string {
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

func TraceMessageText(message *TraceMessage) string {
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