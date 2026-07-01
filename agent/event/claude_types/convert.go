package claude_types

import (
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// FromClaude converts a sequence of Claude Code headless stream-json events
// into the canonical AgentEvent stream. It walks the native events in order,
// emitting zero or more canonical events per native event.
func FromClaude(events []StreamEvent, sessionID string) []types.AgentEvent {
	// Initialize as a non-nil empty slice so json.Marshal produces "[]" rather
	// than "null" when nothing is emitted (e.g. a lone user tool_result line).
	result := []types.AgentEvent{}
	for _, e := range events {
		switch e.Type {
		case EventSystem:
			// subtype "init" marks the start of a step.
			result = append(result, types.AgentEvent{
				Type: types.ActionStepStart,
			})
		case EventAssistant:
			if e.Message == nil {
				continue
			}
			// Iterate content blocks in array order so mixed messages emit
			// actions in the same order they appeared (message < think < tool_call).
			for _, block := range e.Message.Content {
				switch block.Type {
				case "text":
					result = append(result, types.AgentEvent{
						Type: types.ActionMessage,
						Text: block.Text,
					})
				case "thinking":
					result = append(result, types.AgentEvent{
						Type: types.ActionThink,
						Text: block.Thinking,
					})
				case "tool_use":
					result = append(result, types.AgentEvent{
						Type:      types.ActionToolCall,
						Tool:      block.Name,
						ToolInput: block.Input,
					})
				default:
					// unknown block types are skipped
				}
			}
		case EventUser:
			// tool_result echoes fold into the preceding tool call; emit nothing.
			continue
		case EventResult:
			if e.Subtype == "error" || e.IsError {
				result = append(result, types.AgentEvent{
					Type: types.ActionError,
					Text: e.Result,
				})
			} else {
				result = append(result, types.AgentEvent{
					Type: types.ActionDone,
					Text: e.Result,
				})
			}
		default:
			// unknown event types are skipped
		}
	}
	return result
}

// ToClaude converts a canonical AgentEvent stream into Claude Code headless
// stream-json events, one native event per input event.
func ToClaude(events []types.AgentEvent, sessionID string) []StreamEvent {
	result := []StreamEvent{}
	for _, e := range events {
		var ev StreamEvent
		switch e.Type {
		case types.ActionThink:
			ev = StreamEvent{
				Type: EventAssistant,
				Message: &Message{
					Role: "assistant",
					Content: []ContentBlock{
						{Type: "thinking", Thinking: e.Text},
					},
				},
			}
		case types.ActionMessage:
			ev = StreamEvent{
				Type: EventAssistant,
				Message: &Message{
					Role: "assistant",
					Content: []ContentBlock{
						{Type: "text", Text: e.Text},
					},
				},
			}
		case types.ActionToolCall:
			ev = StreamEvent{
				Type: EventAssistant,
				Message: &Message{
					Role: "assistant",
					Content: []ContentBlock{
						{Type: "tool_use", Name: e.Tool, Input: e.ToolInput},
					},
				},
			}
		case types.ActionError:
			ev = StreamEvent{
				Type:    EventResult,
				Subtype: "error",
				IsError: true,
				Result:  e.Text,
			}
		case types.ActionDone:
			ev = StreamEvent{
				Type:    EventResult,
				Subtype: "success",
				IsError: false,
				Result:  e.Text,
			}
		case types.ActionStepStart:
			ev = StreamEvent{
				Type:    EventSystem,
				Subtype: "init",
			}
		default:
			// unknown action types are skipped
			continue
		}
		if sessionID != "" {
			ev.SessionID = sessionID
		}
		result = append(result, ev)
	}
	return result
}
