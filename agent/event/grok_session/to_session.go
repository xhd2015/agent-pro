package grok_session

import (
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// ToSession converts canonical AgentEvents to grok session updates.
func ToSession(events []types.AgentEvent, opts ToOptions) []SessionUpdate {
	var updates []SessionUpdate
	emittedToolCalls := make(map[string]bool)
	nextToolCallID := 1

	for _, ev := range events {
		switch ev.Type {
		case types.ActionMessage:
			switch ev.Role {
			case "user":
				updates = append(updates, userMessageChunk(ev.Text))
			case "assistant":
				updates = append(updates, assistantMessageChunk(ev.Text))
			}
		case types.ActionThink:
			updates = append(updates, thoughtChunk(ev.Text))
		case types.ActionToolCall:
			id := strings.TrimSpace(ev.ToolCallID)
			if id == "" {
				id = fmt.Sprintf("call_%d", nextToolCallID)
				nextToolCallID++
			}
			status := grokStatus(ev)
			if status == "" && ev.Output != "" {
				status = "completed"
			}
			switch status {
			case "pending":
				updates = append(updates, toolCallWire(id, ev.Tool, ev.Text, "pending"))
				emittedToolCalls[id] = true
			case "completed", "failed":
				if !emittedToolCalls[id] {
					updates = append(updates, toolCallWire(id, ev.Tool, ev.Text, "pending"))
					emittedToolCalls[id] = true
				}
				updates = append(updates, toolCallUpdateWire(id, status, ev.Output))
			default:
				updates = append(updates, toolCallWire(id, ev.Tool, ev.Text, "pending"))
				emittedToolCalls[id] = true
			}
		case types.ActionDone:
			updates = append(updates, SessionUpdate{SessionUpdate: "turn_completed"})
		}
	}
	return updates
}

// ToWireLines marshals session updates to JSONL strings.
func ToWireLines(updates []SessionUpdate, opts ToOptions) []string {
	lines := make([]string, 0, len(updates))
	for _, upd := range updates {
		if opts.SessionID != "" {
			inner, err := json.Marshal(upd)
			if err != nil {
				continue
			}
			payload := map[string]any{
				"method": "_x.ai/session/update",
				"params": map[string]any{
					"sessionId": opts.SessionID,
					"update":    json.RawMessage(inner),
				},
			}
			line, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			lines = append(lines, string(line))
			continue
		}
		line, err := json.Marshal(upd)
		if err != nil {
			continue
		}
		lines = append(lines, string(line))
	}
	return lines
}

func grokStatus(ev types.AgentEvent) string {
	if ev.Extensions == nil || ev.Extensions.GrokSession == nil {
		return ""
	}
	return ev.Extensions.GrokSession.Status
}

func userMessageChunk(text string) SessionUpdate {
	content, _ := json.Marshal(map[string]any{
		"type": "text",
		"text": text,
	})
	return SessionUpdate{
		SessionUpdate: "user_message_chunk",
		Content:       content,
	}
}

func thoughtChunk(text string) SessionUpdate {
	content, _ := json.Marshal(map[string]any{
		"type": "text",
		"text": text,
	})
	return SessionUpdate{
		SessionUpdate: "agent_thought_chunk",
		Content:       content,
	}
}

func assistantMessageChunk(text string) SessionUpdate {
	content, _ := json.Marshal(map[string]any{
		"type": "text",
		"text": text,
	})
	return SessionUpdate{
		SessionUpdate: "agent_message_chunk",
		Content:       content,
	}
}

func toolCallWire(toolCallID, kind, title, status string) SessionUpdate {
	return SessionUpdate{
		SessionUpdate: "tool_call",
		ToolCallID:    toolCallID,
		Kind:          kind,
		Title:         title,
		Status:        status,
	}
}

func toolCallUpdateWire(toolCallID, status, output string) SessionUpdate {
	content := []map[string]any{}
	if output != "" {
		content = append(content, map[string]any{
			"type": "content",
			"content": map[string]any{
				"type": "text",
				"text": output,
			},
		})
	}
	raw, _ := json.Marshal(content)
	return SessionUpdate{
		SessionUpdate: "tool_call_update",
		ToolCallID:    toolCallID,
		Status:        status,
		Content:       raw,
	}
}