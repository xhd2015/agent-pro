## Expected

- `ReadEvents` from offset 0 returns all three appended events.
- Event texts appear in append order: `alpha`, `beta`, `gamma`.
- `EventsOffset` is greater than zero after reading.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(resp.Events))
	}
	assertEqual(t, "events[0].Text", resp.Events[0].Text, "alpha")
	assertEqual(t, "events[1].Text", resp.Events[1].Text, "beta")
	assertEqual(t, "events[2].Text", resp.Events[2].Text, "gamma")
	if resp.EventsOffset <= 0 {
		t.Fatalf("expected positive EventsOffset, got %d", resp.EventsOffset)
	}
}
```