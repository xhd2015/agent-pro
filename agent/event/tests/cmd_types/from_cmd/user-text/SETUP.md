# Scenario

**Feature**: a user text event maps to ActionStepStart

```
# user message with a text content block
{"role":"user","content":[{"type":"text","text":"hello"}]} -> ActionStepStart
```

## Preconditions
- `FromCmd` emits one `ActionStepStart` per user `text` block, marking the start of a step.

## Steps
1. Provide one user JSONL line with a single `text` content block.
2. Verify one ActionStepStart.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CmdInput = `{"id":"evt_7","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_7","role":"user","content":[{"type":"text","text":"hello"}]}`
	return nil
}
```
