## Expected

- Second `ReadEvents` call (after first offset) returns zero events.
- No prior event text appears in the second batch.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Events) != 0 {
		t.Fatalf("expected 0 events after offset, got %d: %+v", len(resp.Events), resp.Events)
	}
}
```