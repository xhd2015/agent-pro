## Expected
- Server starts successfully (exit code 0).
- `resp.Port` is 8081 (the fallback port, since 8080 was blocked).
- HTTP request to port 8081 returns 200 with the configured response.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    // Port should have fallen back to 8081
    if resp.Port != 8081 {
        t.Fatalf("expected port 8081 (fallback), got %d", resp.Port)
    }

    // Request should have succeeded on fallback port
    if len(resp.Responses) != 1 {
        t.Fatalf("expected 1 response, got %d", len(resp.Responses))
    }
    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected status 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)
    msg := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
    if msg["content"] != "port fallback works" {
        t.Fatalf("expected 'port fallback works', got %q", msg["content"])
    }
}
```
