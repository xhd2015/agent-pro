## Expected
- HTTP 200 — the empty role matches the "system" role in the request.
- Content is "matched any role".

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
    if msg["content"] != "matched any role" {
        t.Fatalf("expected 'matched any role', got %q", msg["content"])
    }
}
```
