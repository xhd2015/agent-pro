## Expected

- First AutoSendOrResume (ModeRun): no error; `RunCalls >= 1`.
- Second AutoSendOrResume (ModeSend): no error; `SendCalls == 1`.
- `HookCount == 1` total — live follow-up must not call OnTTYStarted again.
- First (only) hook call SessionID matches request.

## Side Effects

- One OnTTYStarted for first establish only.

## Errors

- None on either call.

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
		t.Fatalf("first AutoSendOrResume (ModeRun): %s", resp.ErrString)
	}
	if resp.ErrString2 != "" {
		t.Fatalf("second AutoSendOrResume (ModeSend): %s", resp.ErrString2)
	}
	if resp.RunCalls < 1 {
		t.Fatalf("expected ModeRun dispatch; RunCalls=%d", resp.RunCalls)
	}
	if resp.SendCalls != 1 {
		t.Fatalf("expected ModeSend dispatch once; SendCalls=%d", resp.SendCalls)
	}
	if resp.HookCount != 1 {
		t.Fatalf("OnTTYStarted must not re-fire on live follow-up; HookCount=%d calls=%#v",
			resp.HookCount, resp.HookCalls)
	}
	if resp.HookCalls[0].SessionID != req.SessionID {
		t.Fatalf("info.SessionID: got %q, want %q", resp.HookCalls[0].SessionID, req.SessionID)
	}
}
```
