# Scenario

**Feature**: message_update with non-assistant role (user/toolResult) produces no action

## Preconditions
- message_update with non-assistant role (user/toolResult) produces no action.

## Steps
1. Create a pi message_update with user role.
2. Call FromPi and marshal result.

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
			Role:    "user",
			Content: []pi_types.MessageContent{{Type: "text", Text: "hello"}},
		},
	}}
	return nil
}
```
