## Expected

- `subagent.Run` completes without error.
- `events.jsonl` exists and is non-empty under `SessionLayout.Dir`.
- `messages.jsonl` exists and is non-empty.
- `questions/` directory exists.
- `progress/` directory exists.
- `meta.json` exists with `opencode_session_id` set to mock inner session id.
- No date-nested `sess_*` subdirectory created under temp dir outside `SessionLayout.Dir`.

## Side Effects

- Artifacts written only under flat session dir.

## Errors

- None.

## Exit Code

N/A (library call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	dir := req.SessionDir
	eventsPath := resolvePath(req.Layout, "events.jsonl", req.Layout.EventsPath)
	if !strings.HasPrefix(eventsPath, dir) {
		eventsPath = filepath.Join(dir, "events.jsonl")
	}
	if data, err := os.ReadFile(eventsPath); err != nil || len(data) == 0 {
		t.Fatalf("events.jsonl missing or empty at %s: err=%v", eventsPath, err)
	}

	msgPath := filepath.Join(dir, "messages.jsonl")
	if data, err := os.ReadFile(msgPath); err != nil || len(data) == 0 {
		t.Fatalf("messages.jsonl missing or empty: err=%v", err)
	}

	if !dirExists(filepath.Join(dir, "questions")) {
		t.Fatalf("questions/ missing under %s", dir)
	}
	if !dirExists(filepath.Join(dir, "progress")) {
		t.Fatalf("progress/ missing under %s", dir)
	}

	metaPath := filepath.Join(dir, "meta.json")
	if opencodeID := readMetaField(t, metaPath, "opencode_session_id"); opencodeID != "inner_layout_sess" {
		t.Fatalf("opencode_session_id = %q, want inner_layout_sess", opencodeID)
	}

	// pid removed after run
	if fileExists(filepath.Join(dir, "pid")) {
		t.Fatalf("pid file should be removed after run completes")
	}

	_ = subagent.SessionLayout{}
}
```
