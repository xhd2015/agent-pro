## Expected

- JSON contains session_id, title, and summary.count == 1; no sendable; no ANSI.

```go
import (
	"encoding/json"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if strings.Contains(resp.Stdout, "\x1b[") {
		t.Fatalf("json must not contain ANSI:\n%q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, `"sendable"`) {
		t.Fatalf("json must not include sendable:\n%s", resp.Stdout)
	}
	var env struct {
		Sessions []struct {
			SessionID string `json:"session_id"`
			Title     string `json:"title"`
		} `json:"sessions"`
		Summary struct {
			Count int `json:"count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &env); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, resp.Stdout)
	}
	if env.Summary.Count != 1 || len(env.Sessions) != 1 {
		t.Fatalf("want 1 session, got summary=%d len=%d", env.Summary.Count, len(env.Sessions))
	}
	if env.Sessions[0].SessionID != fixtureListLiveSID {
		t.Fatalf("session_id=%q", env.Sessions[0].SessionID)
	}
	if env.Sessions[0].Title != "json-title" {
		t.Fatalf("title=%q, want json-title", env.Sessions[0].Title)
	}
}
```
