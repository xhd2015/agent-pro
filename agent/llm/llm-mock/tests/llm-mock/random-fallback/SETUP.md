# Scenario

**Feature**: prefix exchanges exhausted → `GenerateEvents` random fallback (not HTTP 400 overflow)

```
# prefix script consumed sequentially per HTTP request
exchange matcher -> prefix exchanges[] (config + optional events input)

# after prefix exhausted
exchange matcher -> events.GenerateEvents(seed, prompt) -> AgentEvent queue
AgentEvent -> OpenAI chat completion JSON (think / tool_calls / message)
```

## Preconditions

- Config may be empty (`exchanges: []`) or have prefix exchanges only.
- After prefix is fully consumed, each HTTP request dequeues the next generated event.
- Old overflow behavior (replay last exchange ×3 then HTTP 400 `no_match`) is removed.

## Steps

1. Grouping `Setup` sets chat-completions endpoint defaults.
2. Leaf `Setup` narrows prefix depth (none vs one) and request count.
3. `Run` starts server, sends ordered HTTP requests, collects responses.
4. Leaf `Assert` checks HTTP 200 and response shape (prefix vs generated).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Endpoint = "/v1/chat/completions"
	req.Method = "POST"
	return nil
}
```