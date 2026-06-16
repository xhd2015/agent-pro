## Expected
- HTTP 200 with `Content-Type: text/event-stream`.
- Body contains SSE events delimited by `data: ` lines and double newlines.
- First event: role delta `"assistant"`, no content, finish_reason null.
- Content events: ~3-character chunks of "Hello streaming world".
- Final event: empty delta, finish_reason `"stop"`.
- Body ends with `data: [DONE]`.
- All events share the same `id` starting with `"chatcmpl-"`.
- All events have `object: "chat.completion.chunk"`.

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

    // Verify content-type header
    ct := r.Headers["Content-Type"]
    if !strings.Contains(ct, "text/event-stream") {
        t.Fatalf("expected Content-Type to contain 'text/event-stream', got %q", ct)
    }

    // Parse SSE events
    events := parseSSEEvents(t, r.Body)

    if len(events) < 3 {
        t.Fatalf("expected at least 3 SSE events, got %d\nbody:\n%s", len(events), r.Body)
    }

    // Verify [DONE] is present
    if !strings.Contains(r.Body, "data: [DONE]") {
        t.Fatalf("expected 'data: [DONE]' in SSE body:\n%s", r.Body)
    }

    // All events should have the same id
    var chatID string
    for i, ev := range events {
        id, _ := ev["id"].(string)
        if !strings.HasPrefix(id, "chatcmpl-") {
            t.Fatalf("event %d: expected id to start with 'chatcmpl-', got %q", i, id)
        }
        if chatID == "" {
            chatID = id
        } else if id != chatID {
            t.Fatalf("event %d: expected id=%q, got %q", i, chatID, id)
        }

        if ev["object"] != "chat.completion.chunk" {
            t.Fatalf("event %d: expected object='chat.completion.chunk', got %q", i, ev["object"])
        }

        choices, ok := ev["choices"].([]any)
        if !ok || len(choices) != 1 {
            t.Fatalf("event %d: expected choices[1], got %v", i, ev["choices"])
        }
        choice := choices[0].(map[string]any)
        delta, ok := choice["delta"].(map[string]any)
        if !ok {
            t.Fatalf("event %d: expected delta object, got %v", i, choice["delta"])
        }
        if i == 0 {
            if delta["role"] != "assistant" {
                t.Fatalf("event 0: expected delta.role='assistant', got %q", delta["role"])
            }
        }
    }

    // Verify all content appeared across chunks
    allContent := ""
    for _, ev := range events {
        choices := ev["choices"].([]any)
        choice := choices[0].(map[string]any)
        delta := choice["delta"].(map[string]any)
        if content, ok := delta["content"].(string); ok {
            allContent += content
        }
    }
    if allContent != "Hello streaming world" {
        t.Fatalf("expected accumulated content='Hello streaming world', got %q", allContent)
    }

    // Final event should have finish_reason "stop"
    lastEvent := events[len(events)-1]
    lastChoice := lastEvent["choices"].([]any)[0].(map[string]any)
    if lastChoice["finish_reason"] != "stop" {
        t.Fatalf("last event: expected finish_reason='stop', got %q", lastChoice["finish_reason"])
    }
}
```
