## Expected

- Pre-seeded user text appears in collected events.
- Mid-run marker `MID_RUN_APPEND_MARKER` appears after delayed append (regression guard).
- Tool output fragment (`agent`) or marker present — tail delivered post-bootstrap content.

## Exit Code

N/A (direct package call)

```go
import (
	"context"
	"errors"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !eventsContainText(resp.EventTexts, midRunSeedUserText) {
		t.Fatalf("missing seeded user text; texts=%v", resp.EventTexts)
	}
	if !eventsContainText(resp.EventTexts, midRunAppendMarker) {
		t.Fatalf("expected mid-run marker %q before ctx cancel; texts=%v\nTailErr=%v",
			midRunAppendMarker, resp.EventTexts, resp.TailErr)
	}
	combined := strings.ToLower(strings.Join(resp.EventTexts, "\n"))
	if !strings.Contains(combined, "agent") && !strings.Contains(combined, strings.ToLower(midRunAppendMarker)) {
		t.Fatalf("expected streamed tool/assistant content; texts=%v", resp.EventTexts)
	}
	if resp.TailErr != nil && !errors.Is(resp.TailErr, context.Canceled) {
		t.Fatalf("expected context.Canceled from tail, got %v", resp.TailErr)
	}
}
```