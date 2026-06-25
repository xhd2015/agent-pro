## Expected

- `subagent.Run` succeeds.
- `meta.json` bytes identical to pre-run snapshot (no `opencode_session_id` patch by subagent).
- `OnAgentComplete` delivers `inner_merged_host_sess`.
- Host fields would remain via external callback persist (not exercised in-file).

## Side Effects

- `events.jsonl` created; meta frozen.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	metaPath := filepath.Join(req.SessionDir, "meta.json")
	after, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !bytes.Equal(req.MetaBytesBeforeRun, after) {
		t.Fatalf("meta.json patched under HostOwnsMeta (merged-meta scenario)\nbefore:\n%s\nafter:\n%s", req.MetaBytesBeforeRun, after)
	}
	if !req.CallbackCalled {
		t.Fatal("OnAgentComplete was not called")
	}
	if req.CallbackInnerSessionID != "inner_merged_host_sess" {
		t.Fatalf("CallbackInnerSessionID = %q, want inner_merged_host_sess", req.CallbackInnerSessionID)
	}
	if _, err := os.Stat(filepath.Join(req.SessionDir, "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl missing: %v", err)
	}
}```
