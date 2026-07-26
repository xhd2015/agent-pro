## Expected

- Turn 1 user and assistant text appear in collected events.
- At least one `ActionDone` event from `turn_completed` is present.
- Turn 2 marker `TURN_TWO_TAIL_MARKER` appears in `EventTexts` (post-`turn_completed` append).
- Tail returns `context.Canceled` after explicit cancel (not nil early exit).

## Exit Code

N/A (direct package call)

```go
import (
	"context"
	"errors"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !eventsContainText(resp.EventTexts, turnOneUserText) {
		t.Fatalf("missing turn 1 user text; texts=%v events=%d", resp.EventTexts, len(resp.Events))
	}
	if !eventsContainText(resp.EventTexts, turnOneAssistantText) {
		t.Fatalf("missing turn 1 assistant text; texts=%v", resp.EventTexts)
	}
	if !eventsContainActionDone(t, resp.Events) {
		t.Fatalf("expected ActionDone from turn_completed; events=%d", len(resp.Events))
	}
	if !eventsContainText(resp.EventTexts, turnTwoTailMarker) {
		t.Fatalf("expected turn 2 marker %q after turn_completed; texts=%v\nTailErr=%v",
			turnTwoTailMarker, resp.EventTexts, resp.TailErr)
	}
	if resp.TailErr != nil && !errors.Is(resp.TailErr, context.Canceled) {
		t.Fatalf("expected context.Canceled from tail, got %v", resp.TailErr)
	}
	combined := strings.Join(resp.EventTexts, "\n")
	if !strings.Contains(strings.ToLower(combined), strings.ToLower(turnTwoUserText)) {
		t.Fatalf("expected turn 2 user text in events; texts=%v", resp.EventTexts)
	}
}
```