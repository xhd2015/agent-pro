package agent_trace

import (
	"encoding/json"
	"time"

	"github.com/xhd2015/agent-pro/agent_trace/types"
)

func isSubtypeStarted(s types.ActivitySubtype) bool   { return s == types.SubtypeStarted }
func isSubtypeCompleted(s types.ActivitySubtype) bool { return s == types.SubtypeCompleted }

func ParseMessages(lines []string, createdAt string) []types.Message {
	baseMs := parseTraceCreatedAtMs(createdAt)
	messages := make([]types.Message, 0, len(lines))
	for i, line := range lines {
		parsed, ok := types.ParseAgentTraceLine(json.RawMessage(line))
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

func appendTraceToolMessage(messages *[]types.Message, event types.AgentTraceActivity, ts int64) {
	if isSubtypeStarted(event.Subtype) {
		*messages = append(*messages, types.Message{
			Role:      types.RoleToolCall,
			Content:   "",
			ToolCall:  &event,
			StartedAt: &ts,
		})
		return
	}
	for i := len(*messages) - 1; i >= 0; i-- {
		msg := &(*messages)[i]
		if msg.Role != types.RoleToolCall || msg.ToolCall == nil || msg.FinishedAt != nil {
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
		if isSubtypeCompleted(event.Subtype) {
			msg.FinishedAt = &ts
		}
		return
	}
	finishedAt := (*int64)(nil)
	if isSubtypeCompleted(event.Subtype) {
		finishedAt = &ts
	}
	*messages = append(*messages, types.Message{
		Role:       types.RoleToolCall,
		Content:    "",
		ToolCall:   &event,
		StartedAt:  &ts,
		FinishedAt: finishedAt,
	})
}
