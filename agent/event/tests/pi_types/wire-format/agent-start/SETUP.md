# Scenario

**Feature**: agent_start has no payload fields beyond `type`

## Preconditions
- agent_start has no payload fields beyond `type`.

## Steps
1. Parse the agent_start event JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"agent_start"}`
	return nil
}
```
