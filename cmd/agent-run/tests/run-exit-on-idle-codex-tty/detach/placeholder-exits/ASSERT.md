---
label: e2e
explanation: >-
  Real Codex TUI via llm-mock-run-codex; empty/placeholder composer must idle-exit
  through the space probe (not tty status input_box).
---

## Expected

- Detach wrote `idle-policy.json` with `exit_on_idle=true` and `10s`.
- After timeout + grace + probe slack, the TTY is **not** live
  (`tcp_reachable` false / status non-zero / no pid).
- Do not require `tty status input_box==empty` at settle.

## Side Effects

- `__serve` / real Codex PTY gone (watchdog `/exit` + shutdown).

## Errors

- None required on detach when `idle-policy.json` exists (WaitReady may time out).

## Exit Code

0 for detach when WaitReady succeeds; policy-present is sufficient otherwise.

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
	if resp.PolicyJSON == "" && resp.DetachExit != 0 {
		t.Fatalf("detach exit=%d stderr=%s stdout=%s", resp.DetachExit, resp.DetachStderr, resp.DetachStdout)
	}
	if !strings.Contains(resp.PolicyJSON, `"exit_on_idle":true`) || !strings.Contains(resp.PolicyJSON, `"idle_timeout":"10s"`) {
		t.Fatalf("idle-policy.json want exit_on_idle + 10s; got %q", resp.PolicyJSON)
	}
	if resp.Alive {
		t.Fatalf("session still live after idle-timeout+grace; pid=%d tcp=%v status=%s", resp.PID, resp.TCPReachable, resp.StatusJSON)
	}
}
```
