## Expected
- Every emitted event (including ActionDone) has `turn_index=0`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Events) < 5 {
		t.Fatalf("expected at least 5 events, got %d: %s", len(resp.Events), formatEvents(resp.Events))
	}
	assertAllTurnIndex(t, resp.Events, 0)
}
```
