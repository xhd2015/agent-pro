## Expected

- `ReadEvents` succeeds with no error.
- Returned events slice is empty (length 0).
- `EventsOffset` is 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.Events == nil {
		t.Fatal("expected non-nil empty events slice")
	}
	if len(resp.Events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(resp.Events))
	}
	assertEqual(t, "EventsOffset", resp.EventsOffset, int64(0))
}
```