## Expected
- HTTP 200.
- `finish_reason` is `"length"`.
- Content is as configured.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)
    choice := obj["choices"].([]any)[0].(map[string]any)

    if choice["finish_reason"] != "length" {
        t.Fatalf("expected finish_reason='length', got %q", choice["finish_reason"])
    }

    msg := choice["message"].(map[string]any)
    if msg["content"] != "This story was cut short due to length constraints" {
        t.Fatalf("expected specific content, got %q", msg["content"])
    }
}
```
