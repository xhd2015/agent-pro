## Expected

- `subagent.Run` succeeds (resume via `SessionID` + flat `Dir`, not meta parse).
- Pre-seeded `seed event` still in `events.jsonl`; new events appended.
- `meta.json` bytes unchanged; no `explicit_session_id` or `opencode_session_id` added.
- `OnAgentComplete` delivers `inner_alias_host_sess`.

## Side Effects

- Events append; meta frozen.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"bytes"
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
	if !strings.Contains(string(data), "seed event") {
		t.Fatalf("pre-seeded event truncated")
	}
	if countJSONLLines(eventsPath) < 2 {
		t.Fatalf("expected appended events")
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	after, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !bytes.Equal(req.MetaBytesBeforeRun, after) {
		t.Fatalf("meta.json changed under HostOwnsMeta (meta-alias scenario)\nbefore:\n%s\nafter:\n%s", req.MetaBytesBeforeRun, after)
	}
	if strings.Contains(string(after), `"explicit_session_id"`) {
		t.Fatalf("explicit_session_id added under HostOwnsMeta:\n%s", after)
	}
	if !req.CallbackCalled {
		t.Fatal("OnAgentComplete was not called")
	}
	if req.CallbackInnerSessionID != "inner_alias_host_sess" {
		t.Fatalf("CallbackInnerSessionID = %q, want inner_alias_host_sess", req.CallbackInnerSessionID)
	}
}```
