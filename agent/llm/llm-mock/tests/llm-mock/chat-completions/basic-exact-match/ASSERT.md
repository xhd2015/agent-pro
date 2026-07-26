---
label: e2e
---

## Expected
- HTTP 200.
- Response is valid JSON with `id`, `object`, `created`, `model`, `choices`, `usage` fields.
- `object` is `"chat.completion"`.
- `model` echoes `"gpt-4"` from the request.
- `choices[0].message.role` is `"assistant"`.
- `choices[0].message.content` is `"Paris"`.
- `choices[0].finish_reason` is `"stop"`.
- `usage` contains `prompt_tokens`, `completion_tokens`, `total_tokens` all set to 0.
- `id` starts with `"chatcmpl-"`.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    if len(resp.Responses) != 1 {
        t.Fatalf("expected 1 response, got %d", len(resp.Responses))
    }
    r := resp.Responses[0]

    if r.StatusCode != 200 {
        t.Fatalf("expected status 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)

    // id starts with "chatcmpl-"
    id, _ := obj["id"].(string)
    if !strings.HasPrefix(id, "chatcmpl-") {
        t.Fatalf("expected id to start with 'chatcmpl-', got %q", id)
    }

    // object
    if obj["object"] != "chat.completion" {
        t.Fatalf("expected object='chat.completion', got %q", obj["object"])
    }

    // model echoed
    if obj["model"] != "gpt-4" {
        t.Fatalf("expected model='gpt-4', got %q", obj["model"])
    }

    // choices
    choices, ok := obj["choices"].([]any)
    if !ok || len(choices) != 1 {
        t.Fatalf("expected choices[1], got %v", obj["choices"])
    }
    choice := choices[0].(map[string]any)

    if idx, _ := choice["index"].(float64); idx != 0 {
        t.Fatalf("expected choice.index=0, got %v", choice["index"])
    }

    message, ok := choice["message"].(map[string]any)
    if !ok {
        t.Fatalf("expected choice.message, got %v", choice["message"])
    }

    if message["role"] != "assistant" {
        t.Fatalf("expected message.role='assistant', got %q", message["role"])
    }

    if message["content"] != "Paris" {
        t.Fatalf("expected message.content='Paris', got %q", message["content"])
    }

    if choice["finish_reason"] != "stop" {
        t.Fatalf("expected finish_reason='stop', got %q", choice["finish_reason"])
    }

    // usage
    usage, ok := obj["usage"].(map[string]any)
    if !ok {
        t.Fatalf("expected usage object, got %v", obj["usage"])
    }
    if pt, _ := usage["prompt_tokens"].(float64); pt != 0 {
        t.Fatalf("expected prompt_tokens=0, got %v", pt)
    }
    if ct, _ := usage["completion_tokens"].(float64); ct != 0 {
        t.Fatalf("expected completion_tokens=0, got %v", ct)
    }
    if tt, _ := usage["total_tokens"].(float64); tt != 0 {
        t.Fatalf("expected total_tokens=0, got %v", tt)
    }
}
```
