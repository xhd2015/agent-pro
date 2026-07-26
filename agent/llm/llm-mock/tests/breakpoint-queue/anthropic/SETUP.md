# Scenario

**Feature**: Anthropic Messages API encoder (`agent/llm/anthropic`)

```
DequeueToBreakpoint -> anthropic/messages encode -> content[] (thinking, tool_use, text)
```

## Preconditions

- Server registers `POST /v1/messages` route (implementation).
- Native `{type:"thinking"}` blocks for prefix thinks; one `tool_use` or `text` breakpoint per response.

## Steps

1. Grouping `messages/` leaves POST to `/v1/messages`.
2. `Assert` checks Anthropic `content[]` block types and agent-events order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.Endpoint = "/v1/messages"
	req.Method = "POST"
	return nil
}
```