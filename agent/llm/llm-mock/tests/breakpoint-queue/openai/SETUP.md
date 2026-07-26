# Scenario

**Feature**: OpenAI wire encoders (`agent/llm/openai`) for chat and responses APIs

```
DequeueToBreakpoint -> openai/chat or openai/responses encoder -> HTTP JSON/SSE
```

## Preconditions

- Breakpoint dequeue semantics apply before encoding.
- Chat encoder: think merged into message content; omitted from tool_call wire.
- Responses encoder: reasoning item before function_call (option B); codex tool remap.

## Steps

1. Grouping splits on endpoint family: `chat/` vs `responses/`.
2. Leaves set preset, endpoint, and stream flag.
3. `Assert` checks OpenAI-specific wire fields.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{"port": 8080, "exchanges": []}`
	req.Method = "POST"
	return nil
}
```