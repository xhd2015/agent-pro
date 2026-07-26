## Expected

- `subagent.Run` succeeds.
- Events written to `CustomEventsPath` (outside `SessionLayout.Dir`).
- `SessionLayout.Dir/events.jsonl` does **not** exist.
- `meta.json` remains under `SessionLayout.Dir` at default location.

## Side Effects

- External events file is non-empty JSONL.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	external := req.CustomEventsPath
	if external == "" {
		t.Fatal("CustomEventsPath not set")
	}
	if data, err := os.ReadFile(external); err != nil || len(data) == 0 {
		t.Fatalf("external events missing or empty at %s: err=%v", external, err)
	}
	if fileExists(filepath.Join(req.SessionDir, "events.jsonl")) {
		t.Fatalf("default events.jsonl should not exist under session dir")
	}
	if _, err := os.Stat(filepath.Join(req.SessionDir, "meta.json")); err != nil {
		t.Fatalf("meta.json missing under session dir: %v", err)
	}
	if got := readMetaField(t, filepath.Join(req.SessionDir, "meta.json"), "opencode_session_id"); got != "inner_custom_events_sess" {
		t.Fatalf("opencode_session_id = %q, want inner_custom_events_sess", got)
	}
}```
