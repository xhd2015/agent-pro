## Expected

- Exit code 0; valid JSON with one session.
- `updated_at` is absolute `2026-07-01T15:04:05Z` (or equivalent RFC3339 form containing that instant).
- That field does not contain ` ago` or `just now`.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline, got %q", resp.Stdout)
	}
	var payload struct {
		Sessions []map[string]any `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Stdout)), &payload); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, resp.Stdout)
	}
	if len(payload.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d:\n%s", len(payload.Sessions), resp.Stdout)
	}
	s := payload.Sessions[0]
	id, _ := s["session_id"].(string)
	if id != "json_abs" {
		t.Fatalf("session_id = %q, want json_abs", id)
	}
	updated, _ := s["updated_at"].(string)
	if updated == "" {
		t.Fatalf("missing updated_at in %#v", s)
	}
	if strings.Contains(updated, "ago") || strings.Contains(updated, "just now") {
		t.Fatalf("JSON updated_at must stay absolute, got %q", updated)
	}
	// Accept RFC3339 or RFC3339Nano for the seeded instant.
	if !strings.HasPrefix(updated, "2026-07-01T15:04:05") {
		t.Fatalf("updated_at = %q, want absolute starting with 2026-07-01T15:04:05", updated)
	}
	if !strings.Contains(updated, "T") {
		t.Fatalf("updated_at should look like RFC3339, got %q", updated)
	}
}
```
