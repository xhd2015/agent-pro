---
label: e2e
---

## Expected

- HTTP 200 (not HTTP 400).
- Valid OpenAI chat completion JSON with `choices[0].message`.
- Response body does not contain `error.type` = `"no_match"`.

```go
import (
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
		t.Fatalf("expected status 200 (random fallback), got %d\nbody: %s", r.StatusCode, r.Body)
	}

	obj := parseJSON(t, r.Body)
	if errObj, ok := obj["error"].(map[string]any); ok {
		if errObj["type"] == "no_match" {
			t.Fatalf("expected random fallback, got no_match error: %s", r.Body)
		}
	}

	choices, ok := obj["choices"].([]any)
	if !ok || len(choices) == 0 {
		t.Fatalf("expected choices array, got %v", obj["choices"])
	}
	choice := choices[0].(map[string]any)
	if _, ok := choice["message"].(map[string]any); !ok {
		t.Fatalf("expected message in choice, got %v", choice)
	}
}
```