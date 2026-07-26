---
label: e2e
---

## Expected

- Three HTTP 200 responses (no `no_match` on the second user turn).
- Requests 1 and 2 return generated content from the shared stream.
- Request 3 returns HTTP 200 with generated content for `"what's wrong with me?"` (not HTTP 400 `no_match`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	if len(resp.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(resp.Responses))
	}

	for i, r := range resp.Responses {
		if r.StatusCode != 200 {
			t.Fatalf("response %d: expected 200 (random fallback), got %d\nbody: %s", i+1, r.StatusCode, r.Body)
		}
		obj := parseJSON(t, r.Body)
		if errObj, ok := obj["error"].(map[string]any); ok {
			if errObj["type"] == "no_match" {
				t.Fatalf("response %d: expected random fallback for second user turn, got no_match: %s", i+1, r.Body)
			}
		}
		choices, ok := obj["choices"].([]any)
		if !ok || len(choices) == 0 {
			t.Fatalf("response %d: expected choices, got %v", i+1, obj)
		}
		msg := choices[0].(map[string]any)["message"].(map[string]any)
		content, _ := msg["content"].(string)
		if content == "" && msg["tool_calls"] == nil {
			t.Fatalf("response %d: expected generated think/message/tool_calls, got empty message: %v", i+1, msg)
		}
	}

	// Third response must answer the new user prompt (not replay turn-1 assistant text).
	r3 := resp.Responses[2]
	obj3 := parseJSON(t, r3.Body)
	msg3 := obj3["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)
	content3, _ := msg3["content"].(string)
	if content3 == "Here's the result for your request about Hello. I've made the necessary changes." {
		t.Fatalf("response 3: expected new generated content for second user turn, got turn-1 replay: %q", content3)
	}
}
```