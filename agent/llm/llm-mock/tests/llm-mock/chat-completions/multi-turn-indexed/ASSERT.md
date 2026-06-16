## Expected
- Two HTTP 200 responses.
- First response: content="first answer".
- Second response: content="second answer".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    if len(resp.Responses) != 2 {
        t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
    }

    r1 := resp.Responses[0]
    if r1.StatusCode != 200 {
        t.Fatalf("response 1: expected 200, got %d", r1.StatusCode)
    }
    obj1 := parseJSON(t, r1.Body)
    m1 := obj1["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
    if m1["content"] != "first answer" {
        t.Fatalf("response 1: expected 'first answer', got %q", m1["content"])
    }

    r2 := resp.Responses[1]
    if r2.StatusCode != 200 {
        t.Fatalf("response 2: expected 200, got %d", r2.StatusCode)
    }
    obj2 := parseJSON(t, r2.Body)
    m2 := obj2["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
    if m2["content"] != "second answer" {
        t.Fatalf("response 2: expected 'second answer', got %q", m2["content"])
    }
}
```
