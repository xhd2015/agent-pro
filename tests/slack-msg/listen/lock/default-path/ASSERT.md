---
label: unit
explanation: product default lock path under HOME; second instance conflicts
---

## Expected

- First instance stays running until SIGTERM.
- Second instance exits non-zero with `another slack-msg is already running`.
- First process output mentions the default lock path
  (`…/.agent-pro/slack-msg.listen.lock`) or equivalently reports lock held at
  that location in the startup banner.

## Exit Code

0 (first instance stopped cleanly after probe)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SecondExitCode == 0 {
		t.Fatalf("expected second instance non-zero exit; stdout=%q stderr=%q", resp.SecondStdout, resp.SecondStderr)
	}
	combined2 := resp.SecondStdout + resp.SecondStderr
	if !strings.Contains(combined2, "another slack-msg is already running") {
		t.Fatalf("second instance output missing singleton message:\n%s", combined2)
	}
	wantLock := expectedDefaultLockPath(req.HomeDir)
	combined1 := resp.Stdout + resp.Stderr
	if !strings.Contains(combined1, wantLock) && !strings.Contains(combined1, "slack-msg.listen.lock") {
		t.Fatalf("first instance output should show default lock path %q\nstdout:\n%s\nstderr:\n%s", wantLock, resp.Stdout, resp.Stderr)
	}
}
```
