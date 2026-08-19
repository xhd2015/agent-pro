---
label: e2e
---

## Expected

- Detach exits 0; `idle-policy.json` has `exit_on_idle=true` and `2s`.
- After idle-timeout + 5s grace + slack, the TTY is **not** live
  (`tcp_reachable` false / status non-zero / no pid).

## Side Effects

- `__serve` / mock grok PTY gone (watchdog `/exit` + shutdown).

## Errors

- None required on detach.

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
	if !strings.Contains(resp.PolicyJSON, `"exit_on_idle":true`) {
		t.Fatalf("idle-policy.json missing exit_on_idle; got %q", resp.PolicyJSON)
	}
	if resp.Alive {
		t.Fatalf("session still live after idle-timeout+grace; screen=%q box=%q status=%s", resp.ScreenStatus, resp.InputBox, resp.StatusJSON)
	}
}
```
