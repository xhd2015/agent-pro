package grok_types

import (
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

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
		case EventToolStarted:
			tool := normalizeGrokToolName(grokToolName(e))
			if tool == "" {
				continue
			}
			result = append(result, types.AgentEvent{
				Type: types.ActionToolCall,
				Tool: tool,
				Text: grokToolName(e),
			})
		case EventToolCompleted:
			// Completion metadata is captured on the matching tool_started event.
			continue
		// unknown types: skipped — no agent events emitted
		}
	}
	return result
}

func grokToolName(e Event) string {
	if name := strings.TrimSpace(e.ToolName); name != "" {
		return name
	}
	return strings.TrimSpace(e.Data)
}

func normalizeGrokToolName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)
	switch lower {
	case "shell":
		return "bash"
	default:
		return lower
	}
}
