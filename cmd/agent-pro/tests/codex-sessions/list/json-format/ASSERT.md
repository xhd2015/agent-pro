## Expected

- JSON parses as `{"sessions":[{...}]}` with one entry.
- Entry contains non-empty `id`, `started_at`, `cwd`, and `path` fields.
- Session id matches the fixture UUID.

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	var payload struct {
		Sessions []struct {
			ID        string `json:"id"`
			StartedAt string `json:"started_at"`
			CWD       string `json:"cwd"`
			Path      string `json:"path"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(resp.JSON, &payload); err != nil {
		t.Fatalf("unmarshal list json: %v\n%s", err, string(resp.JSON))
	}
	if len(payload.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(payload.Sessions))
	}
	s := payload.Sessions[0]
	if s.ID != "01900005-bbbb-7bbb-bbbb-bbbbbbbbbbbb" {
		t.Fatalf("id = %q, want fixture uuid", s.ID)
	}
	if s.StartedAt == "" || s.CWD != "/tmp/project-b" || s.Path == "" {
		t.Fatalf("unexpected session fields: %+v", s)
	}
}
```