---
label: e2e
---

## Expected
- HTTP 200 — empty content matches any user message.
- Content is "catch-all response".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)
    msg := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
    if msg["content"] != "catch-all response" {
        t.Fatalf("expected 'catch-all response', got %q", msg["content"])
    }
}
```
