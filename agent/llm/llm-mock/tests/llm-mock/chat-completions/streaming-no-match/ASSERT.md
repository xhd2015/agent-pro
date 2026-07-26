---
label: e2e
---

## Expected
- HTTP 400.
- Response is JSON (not SSE).
- `error.message` is `"no matching exchange"`.
- `error.type` is `"no_match"`.
- Body does NOT contain `data:` (confirming SSE was not started).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("unexpected run error: %v", resp.Err)
    }

    r := resp.Responses[0]

    if r.StatusCode != 400 {
        t.Fatalf("expected status 400, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    // Should be JSON, not SSE
    assertNotContains(t, r.Body, "data: ")

    obj := parseJSON(t, r.Body)
    errObj := obj["error"].(map[string]any)
    if errObj["message"] != "no matching exchange" {
        t.Fatalf("expected error.message='no matching exchange', got %q", errObj["message"])
    }
    if errObj["type"] != "no_match" {
        t.Fatalf("expected error.type='no_match', got %q", errObj["type"])
    }
}
```
