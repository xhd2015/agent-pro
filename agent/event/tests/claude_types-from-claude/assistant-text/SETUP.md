# Scenario

**Feature**: an assistant text block maps to one ActionMessage

```
# assistant message with a text content block
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"pong"}]}} -> ActionMessage
```

## Preconditions
- `FromClaude` emits one `ActionMessage` per assistant `text` block, carrying the block's `text`.

## Steps
1. Provide one assistant NDJSON line whose `message.content` has a single `text` block.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeInput = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"pong"}]},"session_id":"sess_claude"}`
	return nil
}
```
