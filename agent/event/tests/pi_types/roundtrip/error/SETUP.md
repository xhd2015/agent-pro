# Scenario

**Feature**: Roundtrip: ToPi then FromPi should preserve error information

## Preconditions
- Roundtrip: ToPi then FromPi should preserve error information.

## Steps
1. Create an ActionError event with error message.
2. Call roundtrip and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "something broke",
	}}
	return nil
}
```
