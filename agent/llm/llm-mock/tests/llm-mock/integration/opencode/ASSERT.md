## Expected
- HTTP 200 with model echoed as "gpt-4".
- Response content contains "Paris".
- Response has correct chat completion structure.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)

    // Model echoed back
    if obj["model"] != "gpt-4" {
        t.Fatalf("expected model='gpt-4', got %q", obj["model"])
    }

    // Content contains expected response
    choices := obj["choices"].([]any)
    msg := choices[0].(map[string]any)["message"].(map[string]any)
    content := msg["content"].(string)
    if !strings.Contains(content, "Paris") {
        t.Fatalf("expected content to contain 'Paris', got %q", content)
    }

    // Finish reason
    if choices[0].(map[string]any)["finish_reason"] != "stop" {
        t.Fatalf("expected finish_reason='stop', got %q", choices[0].(map[string]any)["finish_reason"])
    }

    // ID and object
    id, _ := obj["id"].(string)
    if !strings.HasPrefix(id, "chatcmpl-") {
        t.Fatalf("expected id to start with 'chatcmpl-', got %q", id)
    }
    if obj["object"] != "chat.completion" {
        t.Fatalf("expected object='chat.completion', got %q", obj["object"])
    }
}
```
