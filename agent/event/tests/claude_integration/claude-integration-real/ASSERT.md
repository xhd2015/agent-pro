---
label: slow, heavy
explanation: runs the real claude binary headless; parse + LLM latency
---

## Expected
- The test either skips (claude not available) or produces canonical event JSON.
- The JSON output contains at least one `ActionMessage` event whose text contains `pong` (case-insensitive).
- No unhandled error is returned when claude is available.

## Side Effects
- A `claude` headless process is started and waited on during the test.

## Errors
- If the claude binary is unavailable the test is skipped (`t.Skip`).
- If `CLAUDE_SKIP_INTEGRATION=1` is set the test is skipped.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		return
	}
	assertContains(t, resp.Output, `"type":"message"`)
	lower := strings.ToLower(resp.Output)
	if !strings.Contains(lower, "pong") {
		t.Fatalf("expected output to contain \"pong\" (case-insensitive), got:\n%s", resp.Output)
	}
}
```
