## Expected

- User text `bootstrap user from zero` appears in `EventTexts`.
- Assistant text `BOOTSTRAP_ASSISTANT_FROM_ZERO` appears in `EventTexts`.
- Content came from bootstrap (no `AppendSchedules` required).

## Exit Code

N/A (direct package call)

```go
import (
	"context"
	"errors"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Events) == 0 {
		t.Fatalf("expected bootstrap events from startOffset=0; got none\nTailErr=%v", resp.TailErr)
	}
	if !eventsContainText(resp.EventTexts, bootstrapUserText) {
		t.Fatalf("missing bootstrap user text; texts=%v", resp.EventTexts)
	}
	if !eventsContainText(resp.EventTexts, bootstrapAssistantText) {
		t.Fatalf("missing bootstrap assistant text; texts=%v", resp.EventTexts)
	}
	if resp.TailErr != nil && !errors.Is(resp.TailErr, context.Canceled) {
		t.Fatalf("expected context.Canceled from tail, got %v", resp.TailErr)
	}
}
```