## Expected

- `onLine` receives NDJSON row containing `WATCHEVENTS_FINISHED_APPEND_MARKER`.
- `WatchEvents` does not return before append solely because `meta.status == "finished"`.

## Errors

- `WatchErr` must be `context.Canceled` or nil after intentional ctx cancel — not silent early return before append.

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
	if !resp.GotAppended {
		t.Fatalf("WatchEvents did not deliver appended line on finished session; received=%v watchErr=%v",
			resp.ReceivedLines, resp.WatchErr)
	}
	if resp.WatchErr != nil && !errors.Is(resp.WatchErr, context.Canceled) {
		t.Fatalf("unexpected watch error: %v", resp.WatchErr)
	}
}
```