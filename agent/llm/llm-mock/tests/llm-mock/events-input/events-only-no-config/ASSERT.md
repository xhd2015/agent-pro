## Expected

- HTTP 200.
- `choices[0].message.content` is `from-events-only` (sole prefix from events file).
- Server starts without `--config` or `LLM_MOCK_CONFIG` env.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(resp.Responses))
	}
	r := resp.Responses[0]
	if r.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d\nbody: %s", r.StatusCode, r.Body)
	}

	obj := parseJSON(t, r.Body)
	message := obj["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	if message["content"] != "from-events-only" {
		t.Fatalf("expected from-events-only, got %q", message["content"])
	}
}
```