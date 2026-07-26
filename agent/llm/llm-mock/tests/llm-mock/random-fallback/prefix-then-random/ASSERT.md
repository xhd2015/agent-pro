---
label: e2e
---

## Expected

- Two HTTP 200 responses.
- First response `choices[0].message.content` is `from-prefix`.
- Second response is HTTP 200 with generated content (not `from-prefix`, not `no_match`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(resp.Responses))
	}

	r1 := resp.Responses[0]
	if r1.StatusCode != 200 {
		t.Fatalf("response 1: expected 200, got %d\nbody: %s", r1.StatusCode, r1.Body)
	}
	obj1 := parseJSON(t, r1.Body)
	m1 := obj1["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	if m1["content"] != "from-prefix" {
		t.Fatalf("response 1: expected from-prefix, got %q", m1["content"])
	}

	r2 := resp.Responses[1]
	if r2.StatusCode != 200 {
		t.Fatalf("response 2: expected 200 (random fallback), got %d\nbody: %s", r2.StatusCode, r2.Body)
	}
	obj2 := parseJSON(t, r2.Body)
	if errObj, ok := obj2["error"].(map[string]any); ok {
		if errObj["type"] == "no_match" {
			t.Fatalf("response 2: expected random fallback, got no_match: %s", r2.Body)
		}
	}
	choice2 := obj2["choices"].([]any)[0].(map[string]any)
	msg2 := choice2["message"].(map[string]any)
	content2, _ := msg2["content"].(string)
	if content2 == "from-prefix" {
		t.Fatalf("response 2: expected generated content, got prefix replay %q", content2)
	}
	if content2 == "" && msg2["tool_calls"] == nil {
		t.Fatalf("response 2: expected generated think/message/tool_calls, got empty message: %v", msg2)
	}
}
```