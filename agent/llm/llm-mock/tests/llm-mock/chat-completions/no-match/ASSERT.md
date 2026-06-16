## Expected
- HTTP 400.
- Response body contains JSON with `error.message` = `"no matching exchange"`.
- Response body contains `error.type` = `"no_match"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    // Run may return a response even on HTTP errors
    if resp.Err != nil && resp.ExitCode == 0 {
        t.Fatalf("unexpected run error: %v", resp.Err)
    }

    if len(resp.Responses) != 1 {
        t.Fatalf("expected 1 response, got %d", len(resp.Responses))
    }
    r := resp.Responses[0]

    if r.StatusCode != 400 {
        t.Fatalf("expected status 400, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)

    errObj, ok := obj["error"].(map[string]any)
    if !ok {
        t.Fatalf("expected error object, got %v", obj["error"])
    }

    if errObj["message"] != "no matching exchange" {
        t.Fatalf("expected error.message='no matching exchange', got %q", errObj["message"])
    }

    if errObj["type"] != "no_match" {
        t.Fatalf("expected error.type='no_match', got %q", errObj["type"])
    }
}
```
