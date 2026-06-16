package grok_types

import types "github.com/xhd2015/agent-pro/agent/event/types"

func ToGrok(events []types.AgentEvent, sessionID string) []Event {
	var result []Event
	for _, e := range events {
		switch e.Type {
		case types.ActionMessage:
			result = append(result, Event{
				Type: EventText,
				Data: e.Text,
			})
		case types.ActionThink:
			result = append(result, Event{
				Type: EventThought,
				Data: e.Text,
			})
		case types.ActionDone:
			result = append(result, Event{
				Type:      EventEnd,
				SessionID: sessionID,
			})
		case types.ActionError:
			result = append(result, Event{
				Type:    EventText,
				Data:    e.Text,
				IsError: true,
			})
		case types.ActionToolCall, types.ActionStepStart, types.ActionStepFinish:
			// skipped — no grok equivalent
		}
	}
	return result
}

func FromGrok(events []Event) []types.AgentEvent {
	var result []types.AgentEvent
	for _, e := range events {
		if e.IsError {
			result = append(result, types.AgentEvent{
				Type: types.ActionError,
				Text: e.Data,
			})
			continue
		}
		switch e.Type {
		case EventText:
			result = append(result, types.AgentEvent{
				Type: types.ActionMessage,
				Text: e.Data,
			})
		case EventThought:
			result = append(result, types.AgentEvent{
				Type: types.ActionThink,
				Text: e.Data,
			})
		case EventEnd:
			result = append(result, types.AgentEvent{
				Type:      types.ActionDone,
				ToolInput: map[string]any{"session_id": e.SessionID},
			})
		// unknown types: skipped — no agent events emitted
		}
	}
	return result
}
