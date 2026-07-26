# Scenario

**Feature**: session event is metadata and produces no AgentEvent

## Preconditions
- session event is metadata and produces no AgentEvent.

## Steps
1. Create a pi session event.
2. Call FromPi and marshal result.

```go
import (
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_pi"
	req.PiEvents = []pi_types.Event{{
		Type: pi_types.EventTypeSession,
		ID:   "sess_123",
	}}
	return nil
}
```
