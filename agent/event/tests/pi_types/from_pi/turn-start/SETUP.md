# Scenario

**Feature**: turn_start → ActionStepStart PhaseStart

## Preconditions
- turn_start → ActionStepStart PhaseStart.

## Steps
1. Create a pi turn_start event.
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
		Type: pi_types.EventTypeTurnStart,
	}}
	return nil
}
```
