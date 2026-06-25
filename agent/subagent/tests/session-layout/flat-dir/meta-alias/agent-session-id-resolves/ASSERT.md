## Expected

- `subagent.Run` succeeds (session resolved via `agent_session_id`).
- Pre-seeded event line (`seed event`) still present in `events.jsonl`.
- New events appended (total line count > 1).
- `meta.json` retains `agent_session_id`; `explicit_session_id` is **not** added.
- `opencode_session_id` set to `inner_alias_sess`.

## Side Effects

- Resume path used because `agent_session_id` matches `SessionID`.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	eventsPath := filepath.Join(req.SessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "seed event") {
		t.Fatalf("pre-seeded event truncated; content:\n%s", content)
	}
	if countJSONLLines(eventsPath) < 2 {
		t.Fatalf("expected appended events, got %d lines", countJSONLLines(eventsPath))
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	metaStr := string(metaData)
	if strings.Contains(metaStr, `"explicit_session_id"`) {
		t.Fatalf("explicit_session_id should not be added when agent_session_id present:\n%s", metaStr)
	}
	if got := readMetaField(t, metaPath, "agent_session_id"); got != "gen_layout_alias_test" {
		t.Fatalf("agent_session_id = %q, want gen_layout_alias_test", got)
	}
	if got := readMetaField(t, metaPath, "opencode_session_id"); got != "inner_alias_sess" {
		t.Fatalf("opencode_session_id = %q, want inner_alias_sess", got)
	}
}```
