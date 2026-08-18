## Expected

- `RunSessions` returns nil.
- Stdout mentions `resolve` and `--grok-session-id` (full sessions help text may
  include existing list/print usage; not frozen line-by-line here).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("sessions -h error: %v", resp.Err)
	}
	out := resp.Stdout
	if !strings.Contains(out, "resolve") {
		t.Fatalf("sessions help missing %q:\n%s", "resolve", out)
	}
	if !strings.Contains(out, "--grok-session-id") {
		t.Fatalf("sessions help missing %q:\n%s", "--grok-session-id", out)
	}
}
```
