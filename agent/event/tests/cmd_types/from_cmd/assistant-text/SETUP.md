# Scenario

**Feature**: an assistant text block maps to one ActionMessage

```
# assistant message with a text content block
{"role":"assistant","content":[{"type":"text","text":"Hello, world!"}]} -> ActionMessage
```

## Preconditions
- `FromCmd` emits one `ActionMessage` per assistant `text` block, carrying the block's `text`.

## Steps
1. Provide one assistant JSONL line with a single `text` content block.
2. Call `FromCmd` via the root `Run` dispatch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CmdInput = `{"id":"evt_1","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_1","role":"assistant","content":[{"type":"text","text":"Hello, world!"}]}`
	return nil
}
```
