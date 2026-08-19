---
label: e2e
---

## Expected

- Among assistant events with non-empty `phase`, every row has the same non-empty `id`.
- When web runs with `StreamPhases: false` (no phases), a complete assistant message is enough.

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
	var streamID string
	phased := 0
	hasAssistant := false
	for _, ev := range events {
		if ev["type"] != "message" || ev["role"] != "assistant" {
			continue
		}
		hasAssistant = true
		phase, _ := ev["phase"].(string)
		if phase == "" {
			continue
		}
		phased++
		id, _ := ev["id"].(string)
		if id == "" {
			t.Fatalf("phased assistant event missing id: %v", ev)
		}
		if streamID == "" {
			streamID = id
			continue
		}
		if id != streamID {
			t.Fatalf("expected single stream id %q, got %q in event %v", streamID, id, ev)
		}
	}
	if phased > 0 && streamID == "" {
		t.Fatal("no phased assistant stream id found")
	}
	if !hasAssistant {
		t.Fatalf("expected assistant message: %v", events)
	}
}
```
