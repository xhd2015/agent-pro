# Scenario

**Feature**: `--agent-events-file` logs served AgentEvents (not RecordedRequest)

```
mock server --agent-events-file path.jsonl
HTTP serve (prefix or random fallback) -> append AgentEvent JSONL per served event
```

## Preconditions

- Separate from `--events-file` (RecordedRequest admin log).
- Random fallback may split think and message across two HTTP requests (grok pattern).

## Steps

1. Grouping `Setup` ensures chat-completions endpoint defaults.
2. Leaf `Setup` sets config, ordered requests, and optional `AgentEventsFile`.
3. `Run` starts server with `--agent-events-file`, sends HTTP requests, reads log file.

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