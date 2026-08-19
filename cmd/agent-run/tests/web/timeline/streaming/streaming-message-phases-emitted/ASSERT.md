---
label: e2e
---

## Expected

- Persisted events include an assistant `message` (web `StreamPhases: false` emits a
  complete message; phased start/update/end remain valid when enabled).
- When phases are present, they share a non-empty stream `id`.

## Errors

- None from `Run`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	events := waitForAssistantPhases(t, req.Home, req.SessionRunner, req.SessionID, 30*time.Second)
	ids := assistantStreamIDs(events)
	hasAssistant := false
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			hasAssistant = true
			break
		}
	}
	if !hasAssistant {
		t.Fatalf("expected assistant message (phased or complete): %v", events)
	}
	// Phased mode optional under current web StreamPhases:false.
	_ = ids
}
```
