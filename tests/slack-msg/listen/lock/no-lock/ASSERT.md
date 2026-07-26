---
label: unit
explanation: --no-lock disables singleton; second instance is not rejected for lock
---

## Expected

- First instance startup shows lock disabled on a **lock-related** banner line
  (e.g. `lock: (none)` / `lock-file: (none)`), not merely config `(none)`.
- Second instance output does **not** contain `another slack-msg is already running`.
- (Second may later time out while connecting; that is OK if the lock path was not the failure.)

## Exit Code

0 (first instance stopped cleanly after probe)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined1 := resp.Stdout + resp.Stderr
	hasLockNone := false
	for _, line := range strings.Split(combined1, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "lock") && strings.Contains(line, "(none)") {
			hasLockNone = true
			break
		}
	}
	if !hasLockNone {
		t.Fatalf("expected a lock line with (none) in banner; got:\n%s", combined1)
	}
	combined2 := resp.SecondStdout + resp.SecondStderr
	if strings.Contains(combined2, "another slack-msg is already running") {
		t.Fatalf("--no-lock second instance must not hit singleton conflict:\n%s", combined2)
	}
}
```
