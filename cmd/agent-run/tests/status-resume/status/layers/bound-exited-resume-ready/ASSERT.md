## Expected Output

Human multi-layer layout (field labels required; spacing may use column alignment):

```text
session:   grok-tty/test-exited-s1
status:    finished
…

process:
  status:  dead|unknown
  …

terminal:
  status:  unreachable|missing
  …

runner:
  status:     bound
  session_id: 550e8400-e29b-41d4-a716-446655440000
  exited:     true

resume:
  ready: yes
```

## Expected

- Exit code 0.
- Session line includes `grok-tty/test-exited-s1` and storage status `finished`.
- Runner bound with session id and `exited: true` (or true).
- `resume` ready yes.
- Process dead/unknown and terminal unreachable/missing (no live terminal).
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
	assertContains(t, out, req.Runner+"/"+req.SessionID)
	assertContains(t, out, "finished")
	assertContains(t, out, "bound")
	assertContains(t, out, req.RunnerSessionID)
	// exited true — accept "exited: true" / "exited:     true"
	assertContainsAny(t, strings.ToLower(out), "exited: true", "exited:true", "exited:     true")
	// resume ready yes
	low := strings.ToLower(out)
	assertContains(t, low, "resume")
	assertContainsAny(t, low, "ready: yes", "ready:yes", "ready:     yes")
	// process/terminal not live
	assertContainsAny(t, low, "dead", "unknown", "unreachable", "missing")
	assertTrailingNewline(t, out, "status stdout")
}
```
