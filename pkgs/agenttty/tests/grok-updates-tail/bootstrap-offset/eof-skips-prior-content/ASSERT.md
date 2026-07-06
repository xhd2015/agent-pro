## Expected

- Stale marker `STALE_EOF_SKIP_MARKER` does **not** appear in `EventTexts`.
- No user/assistant events emitted when starting at EOF with no new appends.
- Tail ends with `context.Canceled` after explicit cancel.

## Exit Code

N/A (direct package call)

```go
import (
	"context"
	"errors"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if eventsContainText(resp.EventTexts, staleEOFSkipMarker) {
		t.Fatalf("stale content replayed at EOF offset; texts=%v", resp.EventTexts)
	}
	for _, text := range resp.EventTexts {
		if text != "" {
			t.Fatalf("expected no events when startOffset=EOF and no new appends; got texts=%v", resp.EventTexts)
		}
	}
	if resp.TailErr != nil && !errors.Is(resp.TailErr, context.Canceled) {
		t.Fatalf("expected context.Canceled from tail, got %v", resp.TailErr)
	}
}
```