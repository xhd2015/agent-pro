---
label: e2e
explanation: >-
  Real Codex TUI via llm-mock-run-codex; a no-submit draft must keep the
  watchdog from SoftExit (occupancy complement of placeholder-exits).
---

## Expected

- After ready, a distinctive no-submit draft is in the composer (`tty snapshot`
  contains `DRAFT_OCCUPANCY_HOLD_zz9`). Do not assert `tty status input_box`.
- After the same timeout + grace + probe slack window, the TTY is **still live**.

## Side Effects

- Keep-alive `__serve` remains up; draft was not submitted as a turn.

## Errors

- None after a successful no-submit inject.

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.PolicyJSON == "" && resp.DetachExit != 0 {
		t.Fatalf("detach exit=%d stderr=%s stdout=%s", resp.DetachExit, resp.DetachStderr, resp.DetachStdout)
	}
	if !strings.Contains(resp.PolicyJSON, `"exit_on_idle":true`) || !strings.Contains(resp.PolicyJSON, `"idle_timeout":"10s"`) {
		t.Fatalf("idle-policy.json want exit_on_idle + 10s; got %q", resp.PolicyJSON)
	}
	draft := req.Draft
	if draft == "" {
		draft = defaultDraft
	}
	if !resp.DraftInjected {
		t.Fatal("expected no-submit draft inject after sendable")
	}
	if !strings.Contains(resp.DraftSnapshot, draft) {
		t.Fatalf("composer snapshot missing draft %q:\n%s", draft, resp.DraftSnapshot)
	}
	if !resp.Alive {
		t.Fatalf("session not live after idle-timeout+grace with occupied draft; status=%s snapshot=%s", resp.StatusJSON, resp.DraftSnapshot)
	}
}
```
