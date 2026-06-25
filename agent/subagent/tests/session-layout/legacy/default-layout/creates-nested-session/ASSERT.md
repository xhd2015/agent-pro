## Expected

- `subagent.Run` completes without error.
- Session directory is created under `HOME/.agent-pro/subagent/layout-test/sessions/<date>/sess_*`.
- Path contains a date segment (`YYYY/MM/DD`) before `sess_*`.
- `events.jsonl`, `messages.jsonl`, `questions/`, and `progress/` exist in the nested session dir.
- `meta.json` has `opencode_session_id` set to `inner_legacy_sess`.
- No extra `sess_*` directories appear outside the discovered session dir.

## Side Effects

- Legacy layout uses default-on questions and progress feature flags.

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

	dir := resp.SessionDir
	if dir == "" {
		t.Fatal("legacy session dir not discovered")
	}

	base := filepath.Join(req.HomeDir, ".agent-pro", "subagent", "layout-test", "sessions")
	if !strings.HasPrefix(dir, base) {
		t.Fatalf("session dir %q not under legacy base %q", dir, base)
	}
	rel, err := filepath.Rel(base, dir)
	if err != nil {
		t.Fatalf("rel path: %v", err)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 3 || !strings.HasPrefix(parts[len(parts)-1], "sess_") {
		t.Fatalf("expected <date>/sess_* layout, got rel=%q", rel)
	}

	eventsPath := filepath.Join(dir, "events.jsonl")
	if data, err := os.ReadFile(eventsPath); err != nil || len(data) == 0 {
		t.Fatalf("events.jsonl missing or empty: err=%v", err)
	}
	msgPath := filepath.Join(dir, "messages.jsonl")
	if data, err := os.ReadFile(msgPath); err != nil || len(data) == 0 {
		t.Fatalf("messages.jsonl missing or empty: err=%v", err)
	}
	if !dirExists(filepath.Join(dir, "questions")) {
		t.Fatalf("questions/ missing in legacy session")
	}
	if !dirExists(filepath.Join(dir, "progress")) {
		t.Fatalf("progress/ missing in legacy session")
	}
	if got := readMetaField(t, filepath.Join(dir, "meta.json"), "opencode_session_id"); got != "inner_legacy_sess" {
		t.Fatalf("opencode_session_id = %q, want inner_legacy_sess", got)
	}

	assertNoNestedSessUnder(t, req.TempDir, dir)
}```
