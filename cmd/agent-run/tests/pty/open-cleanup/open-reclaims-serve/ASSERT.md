---
label: e2e
---

## Expected

- Open path reaches a keep-alive serve (registry PID and/or listen_addr) under
  the leaf `AGENT_RUN_HOME`.
- After harness reclaim, that serve is no longer alive (process dead and/or
  listen port closed).
- Open CLI itself should exit 0 when instant-attach + fake TUI succeed.

## Side Effects

- No live `__serve` left for the test session/home after reclaim.
- Product KeepAlive for real users is not changed by this leaf.

## Exit Code

0 (open CLI)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("open-cleanup run error: %v\nstdout:\n%s\nstderr:\n%s",
			err, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("open exit code = %d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !resp.ServeAliveBefore {
		t.Fatalf("expected a live serve after run --open before reclaim; terminal=%q pid=%d listen=%q\nstderr:\n%s",
			resp.TerminalSessionID, resp.ServePID, resp.RegistryListenAddr, resp.Stderr)
	}
	if resp.ServeAliveAfter {
		t.Fatalf("serve still alive after harness reclaim; pid=%d listen=%q (leak)",
			resp.ServePID, resp.RegistryListenAddr)
	}
}
```
