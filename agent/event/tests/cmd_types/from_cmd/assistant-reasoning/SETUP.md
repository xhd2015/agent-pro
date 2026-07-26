# Scenario

**Feature**: an assistant reasoning block maps to one ActionThink

```
# assistant message with a reasoning block
{"role":"assistant","content":[{"type":"reasoning","text":"Let me think..."}]} -> ActionThink
```

## Preconditions
- `FromCmd` emits one `ActionThink` per assistant `reasoning` block, carrying the block's `text`.

## Steps
1. Provide one assistant JSONL line with a single `reasoning` content block.
2. Call `FromCmd` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CmdInput = `{"id":"evt_2","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_2","role":"assistant","content":[{"type":"reasoning","text":"Let me think..."}]}`
	return nil
}
```
