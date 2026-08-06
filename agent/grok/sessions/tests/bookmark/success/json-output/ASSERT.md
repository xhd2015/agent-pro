## Expected

- No error; Output non-empty JSON.
- No ANSI escape sequences.
- Decoded payload (array or object) includes session_id, agent_runner=grok,
  title matching fixture (snake_case or camelCase field names both acceptable
  if documented — prefer snake_case `session_id`, `agent_runner`,
  `num_chat_messages`).

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Output == "" {
		t.Fatal("json output empty")
	}
	assertNoANSI(t, resp.Output)

	// Accept list array or single object.
	var arr []map[string]any
	if err := json.Unmarshal([]byte(resp.Output), &arr); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal([]byte(resp.Output), &obj); err2 != nil {
			t.Fatalf("json.Unmarshal: %v / %v\nraw=%s", err, err2, resp.Output)
		}
		arr = []map[string]any{obj}
	}
	if len(arr) < 1 {
		t.Fatal("json empty array")
	}
	found := false
	for _, obj := range arr {
		sid := jsonString(obj, "session_id", "SessionID", "sessionId")
		runner := jsonString(obj, "agent_runner", "AgentRunner", "agentRunner")
		title := jsonString(obj, "title", "Title")
		if sid == req.SessionID {
			found = true
			if runner != "" && runner != "grok" {
				t.Fatalf("agent_runner=%q want grok", runner)
			}
			if title != "" && title != req.Title {
				t.Fatalf("title=%q want %q", title, req.Title)
			}
		}
	}
	if !found {
		t.Fatalf("json missing session_id %s in %s", req.SessionID, resp.Output)
	}
}

func jsonString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}
```
