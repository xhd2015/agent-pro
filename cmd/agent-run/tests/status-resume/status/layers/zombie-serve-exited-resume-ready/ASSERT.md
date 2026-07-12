## Expected Output

Human multi-layer layout after `/exit` with zombie keep-alive serve:

```text
session:   grok-tty/test-zombie-s1
status:    running
…

process:
  status:  alive
  …
  kind:    serve

terminal:
  status:   reachable
  …
  sendable: no

runner:
  status:     bound
  session_id: 550e8400-e29b-41d4-a716-446655440222
  exited:     true

resume:
  ready: yes
```

## Expected

- Exit code 0.
- Process **alive** (keep-alive serve still up) and kind serve when printed.
- Terminal **reachable** with **sendable: no** (exit footer / no idle prompt).
- Runner **bound** with `exited: true` — must **not** treat TCP-reachable alone as still active.
- Resume **ready: yes** (bound ∧ exited).
- Stdout ends with trailing `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	out := resp.Stdout
	low := strings.ToLower(out)
	assertContains(t, out, req.Runner+"/"+req.SessionID)
	assertContains(t, out, "running")
	assertContains(t, out, "bound")
	assertContains(t, out, req.RunnerSessionID)
	// Zombie serve: process still alive / terminal still reachable
	assertContainsAny(t, low, "alive")
	assertContainsAny(t, low, "reachable")
	assertContainsAny(t, low, "sendable: no", "sendable:no", "sendable:  no")
	// Critical: exited true despite reachable terminal
	assertContainsAny(t, low, "exited: true", "exited:true", "exited:     true")
	// Resume gate open
	assertContains(t, low, "resume")
	assertContainsAny(t, low, "ready: yes", "ready:yes", "ready:     yes")
	assertTrailingNewline(t, out, "status stdout")
}
```
