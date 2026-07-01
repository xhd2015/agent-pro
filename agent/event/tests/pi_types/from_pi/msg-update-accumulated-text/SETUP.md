# Scenario

**Feature**: A message_update event where the accumulated text (Content[0].Text) is larger than the delta

## Preconditions
- A message_update event where the accumulated text (Content[0].Text) is larger than the delta.
- This simulates real-world pi streaming: after several deltas have accumulated, each message_update carries the full running text.
- After the fix, FromPi must prefer the Delta field for message_update events, not the full accumulated Content[0].Text.

## Steps
1. Create a pi message_update event with Content[0].Text = long accumulated text, Delta = " feature.".
2. Call FromPi and marshal result.
3. The AgentEvent.Text must be the delta " feature." only, NOT the accumulated text.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeMessageUpdate,
		Message: &pi_types.AgentMessage{
			Role:    "assistant",
			Content: []pi_types.MessageContent{{Type: "text", Text: "The user has given me a detailed requirement for creating a macOS menu bar app. I need to design a comprehensive doctest tree for this feature."}},
		},
		AssistantMessageEvent: &pi_types.AssistantMessageEvent{
			Type:        "text_delta",
			ContentIndex: 0,
			Delta:       " feature.",
		},
	}}
	return nil
}
```
