package pi_types

import (
	"fmt"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func ToPi(events []types.AgentEvent) []Event {
	var result []Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case types.ActionThink:
			result = append(result, convertThinkToPi(e, id)...)
		case types.ActionMessage:
			result = append(result, convertMessageToPi(e, id)...)
		case types.ActionToolCall:
			result = append(result, convertToolCallToPi(e, id)...)
		case types.ActionError:
			msgStart := Event{
				Type: EventTypeMessageStart,
				Message: &AgentMessage{
					Role:    "assistant",
					Content: []MessageContent{{Type: "text", Text: e.Text}},
				},
			}
			msgEnd := Event{
				Type: EventTypeMessageEnd,
				Message: &AgentMessage{
					Role:    "assistant",
					Content: []MessageContent{{Type: "text", Text: e.Text}},
				},
			}
			result = append(result, msgStart, msgEnd)
		case types.ActionDone:
			result = append(result, Event{
				Type: EventTypeAgentEnd,
			})
		case types.ActionStepStart:
			result = append(result, Event{
				Type: EventTypeTurnStart,
			})
		case types.ActionStepFinish:
			result = append(result, Event{
				Type: EventTypeTurnEnd,
			})
		}
	}
	return result
}

func convertThinkToPi(e types.AgentEvent, id string) []Event {
	switch e.Phase {
	case types.PhaseUpdate:
		return []Event{{
			Type: EventTypeMessageUpdate,
			Message: &AgentMessage{
				Role:    "assistant",
				Content: []MessageContent{{Type: "thinking", Thinking: e.Text}},
			},
			AssistantMessageEvent: &AssistantMessageEvent{
				Type:  "thinking_delta",
				Delta: e.Text,
			},
		}}
	default:
		msgStart := Event{
			Type: EventTypeMessageStart,
			Message: &AgentMessage{
				Role:    "assistant",
				Content: []MessageContent{{Type: "thinking", Thinking: e.Text}},
			},
		}
		msgUpdate := Event{
			Type: EventTypeMessageUpdate,
			Message: &AgentMessage{
				Role:    "assistant",
				Content: []MessageContent{{Type: "thinking", Thinking: e.Text}},
			},
			AssistantMessageEvent: &AssistantMessageEvent{
				Type:  "thinking_delta",
				Delta: e.Text,
			},
		}
		msgEnd := Event{
			Type: EventTypeMessageEnd,
			Message: &AgentMessage{
				Role:    "assistant",
				Content: []MessageContent{{Type: "thinking", Thinking: e.Text}},
			},
		}
		return []Event{msgStart, msgUpdate, msgEnd}
	}
}

func convertMessageToPi(e types.AgentEvent, id string) []Event {
	content := []MessageContent{{Type: "text", Text: e.Text}}
	switch e.Phase {
	case types.PhaseStart:
		return []Event{{
			Type:    EventTypeMessageStart,
			Message: &AgentMessage{Role: "assistant", Content: content},
		}}
	case types.PhaseUpdate:
		return []Event{{
			Type:    EventTypeMessageUpdate,
			Message: &AgentMessage{Role: "assistant", Content: content},
			AssistantMessageEvent: &AssistantMessageEvent{
				Type:  "text_delta",
				Delta: e.Text,
			},
		}}
	case types.PhaseEnd:
		return []Event{{
			Type:    EventTypeMessageEnd,
			Message: &AgentMessage{Role: "assistant", Content: content},
		}}
	default:
		msgStart := Event{
			Type:    EventTypeMessageStart,
			Message: &AgentMessage{Role: "assistant", Content: content},
		}
		msgUpdate := Event{
			Type:    EventTypeMessageUpdate,
			Message: &AgentMessage{Role: "assistant", Content: content},
			AssistantMessageEvent: &AssistantMessageEvent{
				Type:  "text_delta",
				Delta: e.Text,
			},
		}
		msgEnd := Event{
			Type:    EventTypeMessageEnd,
			Message: &AgentMessage{Role: "assistant", Content: content},
		}
		return []Event{msgStart, msgUpdate, msgEnd}
	}
}

func convertToolCallToPi(e types.AgentEvent, id string) []Event {
	switch e.Phase {
	case types.PhaseStart:
		return []Event{{
			Type:       EventTypeToolExecStart,
			ToolCallID: id,
			ToolName:   e.Tool,
			Args:       e.ToolInput,
		}}
	case types.PhaseUpdate:
		return []Event{{
			Type:          EventTypeToolExecUpdate,
			ToolCallID:    id,
			ToolName:      e.Tool,
			PartialResult: e.Output,
		}}
	case types.PhaseEnd:
		return []Event{{
			Type:       EventTypeToolExecEnd,
			ToolCallID: id,
			ToolName:   e.Tool,
			Result:     e.Output,
			IsError:    false,
		}}
	default:
		start := Event{
			Type:       EventTypeToolExecStart,
			ToolCallID: id,
			ToolName:   e.Tool,
			Args:       e.ToolInput,
		}
		end := Event{
			Type:       EventTypeToolExecEnd,
			ToolCallID: id,
			ToolName:   e.Tool,
			Result:     e.Output,
			IsError:    false,
		}
		return []Event{start, end}
	}
}

func FromPi(events []Event) []types.AgentEvent {
	var result []types.AgentEvent
	var lastWasMsgStart bool
	var lastMsgText string
	for _, e := range events {
		switch e.Type {
		case EventTypeSession:
			continue
		case EventTypeAgentStart:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:  types.ActionStepStart,
				Phase: types.PhaseStart,
			})
		case EventTypeAgentEnd:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:  types.ActionDone,
				Phase: types.PhaseEnd,
			})
		case EventTypeTurnStart:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:  types.ActionStepStart,
				Phase: types.PhaseStart,
			})
		case EventTypeTurnEnd:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:  types.ActionStepFinish,
				Phase: types.PhaseEnd,
			})
		case EventTypeMessageStart:
			lastWasMsgStart = true
			lastMsgText = extractText(e.Message)
			result = append(result, types.AgentEvent{
				Type:  types.ActionMessage,
				Phase: types.PhaseStart,
				Text:  lastMsgText,
			})
		case EventTypeMessageEnd:
			// Prefer delta over full Content text; deltas are already shown via message_update.
			// Don't output full accumulated text again to prevent duplication.
			text := ""
			if e.AssistantMessageEvent != nil {
				text = e.AssistantMessageEvent.Delta
			}
			if lastWasMsgStart {
				lastWasMsgStart = false
				combined := lastMsgText
				if combined == "" {
					combined = text
				}
				result = append(result, types.AgentEvent{
					Type: types.ActionError,
					Text: combined,
				})
			} else {
				result = append(result, types.AgentEvent{
					Type:  types.ActionMessage,
					Phase: types.PhaseEnd,
					Text:  text,
				})
			}
			lastMsgText = ""
		case EventTypeMessageUpdate:
			lastWasMsgStart = false
			lastMsgText = ""
			if e.Message == nil || e.Message.Role != "assistant" {
				continue
			}
			// Prefer delta over accumulated Content text (streaming UX)
			text := ""
			if e.AssistantMessageEvent != nil {
				text = e.AssistantMessageEvent.Delta
			}
			if text == "" {
				text = extractText(e.Message)
			}
			var agentEvent types.AgentEvent
			if e.AssistantMessageEvent != nil {
				switch e.AssistantMessageEvent.Type {
				case "text_delta":
					agentEvent = types.AgentEvent{
						Type:  types.ActionMessage,
						Phase: types.PhaseUpdate,
						Text:  text,
					}
				case "thinking_delta":
					agentEvent = types.AgentEvent{
						Type:  types.ActionThink,
						Phase: types.PhaseUpdate,
						Text:  text,
					}
				default:
					agentEvent = types.AgentEvent{
						Type:  types.ActionMessage,
						Phase: types.PhaseUpdate,
						Text:  text,
					}
				}
			} else {
				agentEvent = types.AgentEvent{
					Type:  types.ActionMessage,
					Phase: types.PhaseUpdate,
					Text:  text,
				}
			}
			result = append(result, agentEvent)
		case EventTypeToolExecStart:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:      types.ActionToolCall,
				Phase:     types.PhaseStart,
				Tool:      e.ToolName,
				ToolInput: e.Args,
			})
		case EventTypeToolExecUpdate:
			lastWasMsgStart = false
			lastMsgText = ""
			result = append(result, types.AgentEvent{
				Type:   types.ActionToolCall,
				Phase:  types.PhaseUpdate,
				Tool:   e.ToolName,
				Output: toString(e.PartialResult),
			})
		case EventTypeToolExecEnd:
			lastWasMsgStart = false
			lastMsgText = ""
			evt := types.AgentEvent{
				Type:   types.ActionToolCall,
				Phase:  types.PhaseEnd,
				Tool:   e.ToolName,
				Output: toString(e.Result),
			}
			if e.IsError {
				exitCode := 1
				evt.ExitCode = &exitCode
			}
			result = append(result, evt)
		}
	}
	return result
}

func extractText(msg *AgentMessage) string {
	if msg == nil {
		return ""
	}
	for _, c := range msg.Content {
		if c.Text != "" {
			return c.Text
		}
		if c.Thinking != "" {
			return c.Thinking
		}
	}
	return ""
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
