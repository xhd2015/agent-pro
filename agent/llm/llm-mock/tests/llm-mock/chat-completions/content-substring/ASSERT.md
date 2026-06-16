## Expected
- HTTP 200 — substring "French capital" matches "Tell me the French capital please".
- Content is "Paris is the answer".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)
    msg := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
    if msg["content"] != "Paris is the answer" {
        t.Fatalf("expected 'Paris is the answer', got %q", msg["content"])
    }
}
```
