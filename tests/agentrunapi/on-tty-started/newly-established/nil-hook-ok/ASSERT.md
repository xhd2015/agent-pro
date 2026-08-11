## Expected

- No harness panic; AutoSendOrResume returns no error.
- `RunCalls == 1`.
- `HookCount == 0` (nil OnTTYStarted never recorded).

## Side Effects

- None beyond dispatch hook.

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
	_ = req
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	if resp.ErrString != "" {
		t.Fatalf("AutoSendOrResume with nil OnTTYStarted must not fail: %s", resp.ErrString)
	}
	if resp.RunCalls != 1 {
		t.Fatalf("RunCalls: got %d, want 1", resp.RunCalls)
	}
	if resp.HookCount != 0 {
		t.Fatalf("nil OnTTYStarted must not record calls; HookCount=%d", resp.HookCount)
	}
}
```
