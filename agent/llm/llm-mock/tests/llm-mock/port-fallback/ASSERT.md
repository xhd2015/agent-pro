## Expected
- Server starts successfully (exit code 0).
- `resp.Port` is the fallback port (blocked port + 1).
- HTTP request to the fallback port returns 200 with the configured response.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatal(err)
    }
    assertSuccess(t, resp)

    if req.ExpectedFallbackPort <= 0 {
        t.Fatal("ExpectedFallbackPort must be set by Setup")
    }
    if resp.Port != req.ExpectedFallbackPort {
        t.Fatalf("expected port %d (fallback), got %d", req.ExpectedFallbackPort, resp.Port)
    }

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