---
label: e2e
---

## Expected
- HTTP 200 — config loaded via `LLM_MOCK_CONFIG` env var.
- Content is "env config works".

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
    if msg["content"] != "env config works" {
        t.Fatalf("expected 'env config works', got %q", msg["content"])
    }
}
```
