## Expected
- HTTP 200.
- `choices[0].message.content` is `"Hello, world!"`.
- `finish_reason` is `"stop"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    if len(resp.Responses) != 1 {
        t.Fatalf("expected 1 response, got %d", len(resp.Responses))
    }
    r := resp.Responses[0]

    if r.StatusCode != 200 {
        t.Fatalf("expected status 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)

    choices := obj["choices"].([]any)
    choice := choices[0].(map[string]any)
    message := choice["message"].(map[string]any)

    if message["content"] != "Hello, world!" {
        t.Fatalf("expected content='Hello, world!', got %q", message["content"])
    }

    if choice["finish_reason"] != "stop" {
        t.Fatalf("expected finish_reason='stop', got %q", choice["finish_reason"])
    }
}
```
