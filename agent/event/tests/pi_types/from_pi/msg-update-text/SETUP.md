## Preconditions
- message_update with text_delta assistant event → ActionMessage PhaseUpdate.
- After fix: FromPi must prefer Delta (" world") over Content[0].Text ("Hello").

## Steps
1. Create a pi message_update event with text_delta assistant message event, Content="Hello" but Delta=" world".
2. Call FromPi and marshal result.
3. Verify Text = Delta " world", not Content "Hello".

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
			Content: []pi_types.MessageContent{{Type: "text", Text: "Hello"}},
		},
		AssistantMessageEvent: &pi_types.AssistantMessageEvent{
			Type:        "text_delta",
			ContentIndex: 0,
			Delta:       " world",
		},
	}}
	return nil
}
```
