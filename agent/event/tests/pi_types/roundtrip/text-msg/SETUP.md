## Preconditions
- Roundtrip: ToPi then FromPi should preserve text message content.
- After fix: FromPi uses Delta for message_update events. Since ToPi sets Delta = Text for PhaseUpdate,
  the roundtrip correctly preserves the text via the delta path.
- For instant phase (Phase=""), ToPi creates msg_start + msg_update + msg_end.
  msg_update uses Delta, msg_end uses empty Text (deltas already shown).

## Steps
1. Create an ActionMessage event with text content.
2. Call roundtrip (ToPi → FromPi) and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Text: "Hello world",
	}}
	return nil
}
```
