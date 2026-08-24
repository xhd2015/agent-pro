## Expected

- JSON includes grok session_id, iterm_session_id, app, contents.

```go
import (
	"encoding/json"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	var got map[string]any
	if err := json.Unmarshal([]byte(resp.Stdout), &got); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, resp.Stdout)
	}
	if got["session_id"] != req.SessionID {
		t.Fatalf("session_id = %v, want %s", got["session_id"], req.SessionID)
	}
	if got["iterm_session_id"] != "w2t1p0" {
		t.Fatalf("iterm_session_id = %v, want w2t1p0", got["iterm_session_id"])
	}
	if got["app"] != "/Applications/iTerm.app" {
		t.Fatalf("app = %v", got["app"])
	}
	if got["contents"] != "json pane" {
		t.Fatalf("contents = %v", got["contents"])
	}
	if got["source"] != "iterm" {
		t.Fatalf("source = %v, want iterm", got["source"])
	}
	if _, ok := got["agent_run_session_id"]; ok {
		t.Fatalf("agent_run_session_id present on iterm path: %v", got["agent_run_session_id"])
	}
	if strings.Contains(resp.Stdout, "\x1b[") {
		t.Fatalf("JSON must have no ANSI: %q", resp.Stdout)
	}
}
```
