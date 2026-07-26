# Scenario

**Feature**: `DequeueToBreakpoint` — one breakpoint per HTTP serve across endpoints

```
genQueue [think*, breakpoint, ...] -> DequeueToBreakpoint
#1 HTTP consumes think* + first tool_call or message
#2 HTTP consumes next breakpoint (never two tool_calls in one response)
```

## Preconditions

- Empty `exchanges: []` unless leaf overrides.
- `--mock-events-preset` seeds `genQueue` at server startup.
- Breakpoints are `tool_call` and `message` only; `think` collapses forward.

## Steps

1. Grouping `Setup` defaults endpoint to chat completions unless leaf sets `/v1/responses`.
2. Leaves set preset name and HTTP request count/order.
3. `Assert` checks response wire shape and `--agent-events-file` consumption order.

## Context

- `think-tool-message` preset queue: `[think, tool_call bash, message]`.
- `two-tool-message` preset queue (implementer adds): `[tool_call bash, tool_call read, message]`.
- `think-message` preset queue: `[think, message]` — single HTTP merges think into reply on chat wire.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.Endpoint = "/v1/chat/completions"
	req.Method = "POST"
	return nil
}
```