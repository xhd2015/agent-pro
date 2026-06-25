## Expected

- `subagent.Run` succeeds.
- `OnAgentComplete` called exactly once.
- `CallbackInnerSessionID` is `inner_callback_sess`.
- `CallbackAgentRunner` is `fake-codex`.
- `meta.json` has no `opencode_session_id` (subagent did not write; host applies callback externally).

## Side Effects

- `events.jsonl` created.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !req.CallbackCalled {
		t.Fatal("OnAgentComplete was not called")
	}
	if req.CallbackInnerSessionID != "inner_callback_sess" {
		t.Fatalf("CallbackInnerSessionID = %q, want inner_callback_sess", req.CallbackInnerSessionID)
	}
	if req.CallbackAgentRunner != "fake-codex" {
		t.Fatalf("CallbackAgentRunner = %q, want fake-codex", req.CallbackAgentRunner)
	}
	if got := readMetaField(t, filepath.Join(req.SessionDir, "meta.json"), "opencode_session_id"); got != "" {
		t.Fatalf("subagent wrote opencode_session_id = %q under HostOwnsMeta", got)
	}
	if _, err := os.Stat(filepath.Join(req.SessionDir, "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl missing: %v", err)
	}
}```
