## Expected
- HTTP 200.
- `message.content` is null.
- `message.tool_calls` is present with one entry.
- `tool_calls[0].id` is `"call_1"`.
- `tool_calls[0].type` is `"function"`.
- `tool_calls[0].function.name` is `"get_weather"`.
- `tool_calls[0].function.arguments` is `{"city":"Paris"}`.
- `finish_reason` is `"tool_calls"`.

```go
import (
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)
    choice := obj["choices"].([]any)[0].(map[string]any)
    message := choice["message"].(map[string]any)

    // content should be null
    if message["content"] != nil {
        t.Fatalf("expected null content, got %v", message["content"])
    }

    // tool_calls
    toolCalls, ok := message["tool_calls"].([]any)
    if !ok {
        t.Fatalf("expected tool_calls array, got %v", message["tool_calls"])
    }
    if len(toolCalls) != 1 {
        t.Fatalf("expected 1 tool_call, got %d", len(toolCalls))
    }

    tc := toolCalls[0].(map[string]any)
    if tc["id"] != "call_1" {
        t.Fatalf("expected id='call_1', got %q", tc["id"])
    }
    if tc["type"] != "function" {
        t.Fatalf("expected type='function', got %q", tc["type"])
    }

    fn := tc["function"].(map[string]any)
    if fn["name"] != "get_weather" {
        t.Fatalf("expected function.name='get_weather', got %q", fn["name"])
    }
    if fn["arguments"] != `{"city":"Paris"}` {
        t.Fatalf("expected arguments='{\"city\":\"Paris\"}', got %q", fn["arguments"])
    }

    // finish_reason
    if choice["finish_reason"] != "tool_calls" {
        t.Fatalf("expected finish_reason='tool_calls', got %q", choice["finish_reason"])
    }
}
```
