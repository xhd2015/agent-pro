---
label: e2e
---

## Expected
- HTTP 200 for both requests.
- First response model echoes "gpt-4", content is "world".
- Second response model echoes "gpt-3.5-turbo", content is "second response".
- Both responses have correct structure (id, object, choices, usage).

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    if len(resp.Responses) != 2 {
        t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
    }

    // First request
    r0 := resp.Responses[0]
    if r0.StatusCode != 200 {
        t.Fatalf("request 0: expected 200, got %d\nbody: %s", r0.StatusCode, r0.Body)
    }
    obj0 := parseJSON(t, r0.Body)
    if obj0["model"] != "gpt-4" {
        t.Fatalf("request 0: expected model='gpt-4', got %q", obj0["model"])
    }
    choices0 := obj0["choices"].([]any)
    msg0 := choices0[0].(map[string]any)["message"].(map[string]any)
    if msg0["content"] != "world" {
        t.Fatalf("request 0: expected content='world', got %q", msg0["content"])
    }
    id0, _ := obj0["id"].(string)
    if !strings.HasPrefix(id0, "chatcmpl-") {
        t.Fatalf("request 0: expected id to start with 'chatcmpl-', got %q", id0)
    }

    // Second request
    r1 := resp.Responses[1]
    if r1.StatusCode != 200 {
        t.Fatalf("request 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
    }
    obj1 := parseJSON(t, r1.Body)
    if obj1["model"] != "gpt-3.5-turbo" {
        t.Fatalf("request 1: expected model='gpt-3.5-turbo', got %q", obj1["model"])
    }
    choices1 := obj1["choices"].([]any)
    msg1 := choices1[0].(map[string]any)["message"].(map[string]any)
    if msg1["content"] != "second response" {
        t.Fatalf("request 1: expected content='second response', got %q", msg1["content"])
    }
    id1, _ := obj1["id"].(string)
    if !strings.HasPrefix(id1, "chatcmpl-") {
        t.Fatalf("request 1: expected id to start with 'chatcmpl-', got %q", id1)
    }

    // Usage check
    usage0 := obj0["usage"].(map[string]any)
    if pt, _ := usage0["prompt_tokens"].(float64); pt != 0 {
        t.Fatalf("request 0: expected prompt_tokens=0, got %v", pt)
    }
}
```
