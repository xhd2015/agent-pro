package crush

import (
	"encoding/json"

	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
)

func UnwrapEvent(data []byte) (*crush_types.Event, error) {
	var outer struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &outer); err != nil {
		return nil, nil
	}

	var inner struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(outer.Payload, &inner); err != nil {
		return nil, nil
	}

	var eventType crush_types.EventType
	switch outer.Type {
	case "message":
		eventType = crush_types.EventMessage
	case "agent_event":
		eventType = crush_types.EventAgentEvent
	case "run_complete":
		eventType = crush_types.EventRunComplete
	default:
		return nil, nil
	}

	return &crush_types.Event{
		Type:    eventType,
		Payload: inner.Payload,
	}, nil
}
