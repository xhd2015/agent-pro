## Expected

- Stdout JSON has `source=agent-run`, agent-run session id, sanitized contents.
- Contents not called; AgentRun probed once.

```go
import (
	"encoding/json"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoContents(t, resp)
	if len(resp.AgentRunCalls) != 1 || resp.AgentRunCalls[0] != req.SessionID {
		t.Fatalf("AgentRunCalls = %v, want [%s]", resp.AgentRunCalls, req.SessionID)
	}
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0", resp.ListITermCalls)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(resp.Stdout), &got); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, resp.Stdout)
	}
	if got["source"] != "agent-run" {
		t.Fatalf("source = %v, want agent-run", got["source"])
	}
	if got["agent_run_session_id"] != "ar-fixture-session" {
		t.Fatalf("agent_run_session_id = %v", got["agent_run_session_id"])
	}
	if got["session_id"] != req.SessionID {
		t.Fatalf("session_id = %v", got["session_id"])
	}
	contents, _ := got["contents"].(string)
	if !strings.Contains(contents, "agent-run single frame") {
		t.Fatalf("contents missing agent-run text: %q", contents)
	}
	if strings.Count(contents, "❯") != 1 {
		t.Fatalf("want exactly one composer glyph, got %q", contents)
	}
}
```
