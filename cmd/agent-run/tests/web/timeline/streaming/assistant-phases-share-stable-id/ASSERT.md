## Expected

- Among assistant events with non-empty `phase`, every row has the same non-empty `id`.

## Errors

- None from `Run`.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	events := waitForAssistantPhases(t, req.Home, req.SessionRunner, req.SessionID, 30*time.Second)
	var streamID string
	for _, ev := range events {
		if ev["type"] != "message" || ev["role"] != "assistant" {
			continue
		}
		phase, _ := ev["phase"].(string)
		if phase == "" {
			continue
		}
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
	if streamID == "" {
		t.Fatal("no phased assistant stream id found")
	}
}
```