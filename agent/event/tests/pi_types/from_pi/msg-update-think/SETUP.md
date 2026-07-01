# Scenario

**Feature**: message_update with thinking_delta assistant event → ActionThink PhaseUpdate

## Preconditions
- message_update with thinking_delta assistant event → ActionThink PhaseUpdate.
- After fix: FromPi must prefer Delta (" deeper") over Content[0].Thinking ("think").

## Steps
1. Create a pi message_update event with thinking_delta, Content[0].Thinking="think" but Delta=" deeper".
2. Call FromPi and marshal result.
3. Verify Text = Delta " deeper", not Content[0].Thinking "think".

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
			Content: []pi_types.MessageContent{{Type: "thinking", Thinking: "think"}},
		},
		AssistantMessageEvent: &pi_types.AssistantMessageEvent{
			Type:        "thinking_delta",
			ContentIndex: 0,
			Delta:       " deeper",
		},
	}}
	return nil
}
```
