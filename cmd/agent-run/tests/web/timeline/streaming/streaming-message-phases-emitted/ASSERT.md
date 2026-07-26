---
label: e2e
---

## Expected

- Persisted events include assistant `message` rows with `phase` of `start`, `update`, and `end`.
- Phased assistant events share a non-empty `id` (one stream per turn).

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
	if len(ids) < 1 {
		t.Fatalf("expected at least one assistant stream id across phased events: %v", events)
	}
}
```