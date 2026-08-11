## Expected

- No API error from AutoSendOrResume.
- `RunCalls == 1` (ModeRun dispatch via RunSession hook).
- `HookCount == 1` (OnTTYStarted exactly once).
- Recorded call `SessionID` equals request SessionID.
- When product fills them, Runner/Workspace match request (non-empty expected).

## Side Effects

- One OnTTYStarted invocation; no second call within this run.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	if resp.ErrString != "" {
		t.Fatalf("AutoSendOrResume: %s", resp.ErrString)
	}
	if resp.RunCalls != 1 {
		t.Fatalf("RunCalls: got %d, want 1", resp.RunCalls)
	}
	if resp.HookCount != 1 {
		t.Fatalf("OnTTYStarted must fire once on first establish; HookCount=%d calls=%#v",
			resp.HookCount, resp.HookCalls)
	}
	got := resp.HookCalls[0]
	if got.SessionID != req.SessionID {
		t.Fatalf("info.SessionID: got %q, want %q", got.SessionID, req.SessionID)
	}
	// Runner/Workspace when known should match opts (implementer fills from opts/meta).
	if got.Runner != "" && got.Runner != req.Runner {
		t.Fatalf("info.Runner: got %q, want %q (or empty if unknown)", got.Runner, req.Runner)
	}
	if got.Workspace != "" && got.Workspace != req.Workspace {
		t.Fatalf("info.Workspace: got %q, want %q (or empty if unknown)", got.Workspace, req.Workspace)
	}
	// Prefer non-empty when opts provided them.
	if got.Runner == "" {
		t.Fatalf("info.Runner: expected %q from opts, got empty", req.Runner)
	}
	if got.Workspace == "" {
		t.Fatalf("info.Workspace: expected %q from opts, got empty", req.Workspace)
	}
}
```
