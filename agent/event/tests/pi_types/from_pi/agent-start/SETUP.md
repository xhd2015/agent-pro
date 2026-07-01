# Scenario

**Feature**: agent_start → ActionStepStart PhaseStart

## Preconditions
- agent_start → ActionStepStart PhaseStart.

## Steps
1. Create a pi agent_start event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeAgentStart,
	}}
	return nil
}
```
