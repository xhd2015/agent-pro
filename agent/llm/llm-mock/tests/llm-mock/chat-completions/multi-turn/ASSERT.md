---
label: e2e
---

## Expected
- Two HTTP 200 responses.
- First response: content="Paris", finish_reason="stop".
- Second response: content="you asked for the French capital", finish_reason="stop".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    if len(resp.Responses) != 2 {
        t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
    }

    // First response
    r1 := resp.Responses[0]
    if r1.StatusCode != 200 {
        t.Fatalf("response 1: expected status 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
    }
    obj1 := parseJSON(t, r1.Body)
    c1 := obj1["choices"].([]any)[0].(map[string]any)
    m1 := c1["message"].(map[string]any)
    if m1["content"] != "Paris" {
        t.Fatalf("response 1: expected content='Paris', got %q", m1["content"])
    }
    if c1["finish_reason"] != "stop" {
        t.Fatalf("response 1: expected finish_reason='stop', got %q", c1["finish_reason"])
    }

    // Second response
    r2 := resp.Responses[1]
    if r2.StatusCode != 200 {
        t.Fatalf("response 2: expected status 200, got %d\nbody: %s", r2.StatusCode, r2.Body)
    }
    obj2 := parseJSON(t, r2.Body)
    c2 := obj2["choices"].([]any)[0].(map[string]any)
    m2 := c2["message"].(map[string]any)
    if m2["content"] != "you asked for the French capital" {
        t.Fatalf("response 2: expected content='you asked for the French capital', got %q", m2["content"])
    }
    if c2["finish_reason"] != "stop" {
        t.Fatalf("response 2: expected finish_reason='stop', got %q", c2["finish_reason"])
    }
}
```
