---
label: e2e
---

## Expected

- Detach exits 0 and writes `idle-policy.json` with `exit_on_idle=true` and
  compact `2s`.
- At settle (`ObserveAfter=2s`), `tty status --json` is live and sendable.
- `screen_status` is `idle` (not `banner` on modern boxed chrome).
- `input_box` is `empty` (boxed `│ ❯ … │` is not draft).

## Side Effects

- Keep-alive `__serve` still up at T+2s (timeout+grace is later).

## Errors

- None.

## Exit Code

0 for detach.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.DetachExit != 0 {
		t.Fatalf("detach exit=%d stderr=%s", resp.DetachExit, resp.DetachStderr)
	}
	if !strings.Contains(resp.PolicyJSON, `"exit_on_idle":true`) || !strings.Contains(resp.PolicyJSON, `"idle_timeout":"2s"`) {
		t.Fatalf("idle-policy.json want exit_on_idle + 2s; got %q", resp.PolicyJSON)
	}
	if !resp.Alive || !resp.Sendable {
		t.Fatalf("want live sendable session at settle; alive=%v sendable=%v status=%s", resp.Alive, resp.Sendable, resp.StatusJSON)
	}
	if resp.ScreenStatus != "idle" {
		t.Fatalf("screen_status=%q want idle (live boxed chrome); status=%s", resp.ScreenStatus, resp.StatusJSON)
	}
	if resp.InputBox != "empty" {
		t.Fatalf("input_box=%q want empty (boxed empty composer); status=%s", resp.InputBox, resp.StatusJSON)
	}
}
```
