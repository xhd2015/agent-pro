## Expected
- The test either skips (crush not available) or produces canonical event JSON.
- The JSON output contains at least one `ActionMessage` event.
- That message text contains `"paris"` (case-insensitive match).
- No unhandled error is returned.

## Side Effects
- A crush server is started and killed during the test.
- A workspace is created on the local crush instance.
- SSE events are consumed.

## Errors
- If the crush binary is unavailable the test is skipped (`t.Skip`).
- If the server fails to start, health check times out, or the agent errors,
  the error is captured by `Run` and surfaced through `err` (non-nil).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		return
	}
	assertContains(t, resp.Output, `"type":"message"`)

	lower := strings.ToLower(resp.Output)
	if !strings.Contains(lower, "paris") {
		t.Fatalf("expected output to contain \"paris\" (case-insensitive), got:\n%s", resp.Output)
	}
}
```
