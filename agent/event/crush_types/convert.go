package crush_types

import (
	"encoding/json"
	"fmt"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func ToCrush(events []types.AgentEvent, sessionID string) []Event {
	var result []Event
	for i, e := range events {
		id := e.ID
		if id == "" {
			id = fmt.Sprintf("evt_%d", i+1)
		}
		switch e.Type {
		case types.ActionThink:
			part := Part{
				Type: PartReasoning,
				Data: mustMarshal(ReasoningData{Thinking: e.Text}),
			}
			payload := MessagePayload{
				ID:        id,
				Role:      "assistant",
				SessionID: sessionID,
				Parts:     []Part{part},
			}
			result = append(result, Event{
				Type:    EventMessage,
				Payload: mustMarshal(payload),
			})
		case types.ActionMessage:
			part := Part{
				Type: PartText,
				Data: mustMarshal(TextData{Text: e.Text}),
			}
			payload := MessagePayload{
				ID:        id,
				Role:      "assistant",
				SessionID: sessionID,
				Parts:     []Part{part},
			}
			result = append(result, Event{
				Type:    EventMessage,
				Payload: mustMarshal(payload),
			})
		case types.ActionToolCall:
			input := map[string]any{}
			if e.ToolInput != nil {
				for k, v := range e.ToolInput {
					input[k] = v
				}
			}
			part := Part{
				Type: PartToolCall,
				Data: mustMarshal(map[string]any{
					"id":    id,
					"name":  e.Tool,
					"input": input,
				}),
			}
			payload := MessagePayload{
				ID:        id,
				Role:      "assistant",
				SessionID: sessionID,
				Parts:     []Part{part},
			}
			result = append(result, Event{
				Type:    EventMessage,
				Payload: mustMarshal(payload),
			})
		case types.ActionError:
			payload := AgentEventPayload{
				Type:  "error",
				Error: e.Text,
			}
			result = append(result, Event{
				Type:    EventAgentEvent,
				Payload: mustMarshal(payload),
			})
		case types.ActionDone:
			payload := RunCompletePayload{
				SessionID: sessionID,
				MessageID: id,
			}
			result = append(result, Event{
				Type:    EventRunComplete,
				Payload: mustMarshal(payload),
			})
		}
	}
	return result
}

func FromCrush(events []Event, _ string) []types.AgentEvent {
	result := make([]types.AgentEvent, 0)
	for _, e := range events {
		switch e.Type {
		case EventMessage:
			mp := e.messagePayload()
			if mp == nil || mp.Role != "assistant" {
				continue
			}
			for _, part := range mp.Parts {
				switch part.Type {
				case PartReasoning:
					d := part.reasoningData()
					if d != nil {
						result = append(result, types.AgentEvent{
							ID:   "crush:" + mp.ID,
							Type: types.ActionThink,
							Text: d.Thinking,
						})
					}
				case PartText:
					d := part.textData()
					if d != nil {
						result = append(result, types.AgentEvent{
							ID:   "crush:" + mp.ID,
							Type: types.ActionMessage,
							Text: d.Text,
						})
					}
				case PartToolCall:
					var toolInput map[string]any
					var toolName string
					var toolID string
					if raw, ok := part.parseDataMap(); ok {
						if v, ok := raw["name"].(string); ok {
							toolName = v
						}
						if v, ok := raw["id"].(string); ok {
							toolID = v
						}
						switch v := raw["input"].(type) {
						case string:
							_ = json.Unmarshal([]byte(v), &toolInput)
						case map[string]any:
							toolInput = v
						}
					}
					if toolInput == nil {
						toolInput = map[string]any{}
					}
					if toolName != "" {
						result = append(result, types.AgentEvent{
							ID:        "crush:" + toolID,
							Type:      types.ActionToolCall,
							Tool:      toolName,
							ToolInput: toolInput,
						})
					}
				case PartToolResult:
				case PartFinish:
					d := part.finishData()
					if d != nil && d.Reason == FinishReasonError {
						result = append(result, types.AgentEvent{
							ID:   "crush:" + mp.ID,
							Type: types.ActionError,
							Text: d.Message,
						})
					}
				}
			}
		case EventAgentEvent:
			aep := e.agentEventPayload()
			if aep != nil && aep.Type == "error" {
				id := "crush:" + aep.RunID
				if aep.RunID == "" {
					id = ""
				}
				result = append(result, types.AgentEvent{
					ID:   id,
					Type: types.ActionError,
					Text: aep.Error,
				})
			}
		case EventRunComplete:
			rcp := e.runCompletePayload()
			if rcp != nil {
				text := rcp.Text
				if text == "" {
					text = rcp.Error
				}
				id := "crush:" + rcp.MessageID
				if rcp.MessageID == "" {
					id = "crush:" + rcp.RunID
				}
				if rcp.MessageID == "" && rcp.RunID == "" {
					id = ""
				}
				result = append(result, types.AgentEvent{
					ID:   id,
					Type: types.ActionDone,
					Text: text,
				})
			}
		}
	}
	return result
}
