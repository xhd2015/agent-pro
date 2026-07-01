# Scenario

**Feature**: an assistant thinking block maps to one ActionThink

```
# thinking content block becomes a canonical think action
{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"pondering"}]}} -> ActionThink
```

## Preconditions
- `FromClaude` emits one `ActionThink` per assistant `thinking` block, with `Text` set to the block's `thinking` text.

## Steps
1. Provide one assistant NDJSON line whose `message.content` has a single `thinking` block.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudeInput = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"thinking","thinking":"pondering the question"}]},"session_id":"sess_claude"}`
	return nil
}
```
