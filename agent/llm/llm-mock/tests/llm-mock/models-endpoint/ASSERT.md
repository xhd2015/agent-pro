## Expected
- HTTP 200.
- `object` is `"list"`.
- `data` is an array with one entry.
- `data[0].id` is `"mock-model"`.
- `data[0].object` is `"model"`.
- `data[0].owned_by` is `"llm-mock"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    r := resp.Responses[0]
    if r.StatusCode != 200 {
        t.Fatalf("expected 200, got %d\nbody: %s", r.StatusCode, r.Body)
    }

    obj := parseJSON(t, r.Body)

    if obj["object"] != "list" {
        t.Fatalf("expected object='list', got %q", obj["object"])
    }

    data, ok := obj["data"].([]any)
    if !ok {
        t.Fatalf("expected data array, got %v", obj["data"])
    }
    if len(data) != 1 {
        t.Fatalf("expected 1 model, got %d", len(data))
    }

    model := data[0].(map[string]any)
    if model["id"] != "mock-model" {
        t.Fatalf("expected id='mock-model', got %q", model["id"])
    }
    if model["object"] != "model" {
        t.Fatalf("expected object='model', got %q", model["object"])
    }
    if model["owned_by"] != "llm-mock" {
        t.Fatalf("expected owned_by='llm-mock', got %q", model["owned_by"])
    }
}
```
