## Expected

- Exit code 0.
- Runner bound with `exited: false`.
- `resume.ready: no` with reason mentioning send and/or still active / not exited.
- Terminal reachable (or sendable yes) when probe succeeds.
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
	assertContains(t, out, "bound")
	assertContainsAny(t, low, "exited: false", "exited:false", "exited:     false")
	assertContainsAny(t, low, "ready: no", "ready:no", "ready:     no")
	assertContainsAny(t, low, "send", "still active", "not exited", "live", "active")
	assertTrailingNewline(t, out, "status stdout")
}
```
