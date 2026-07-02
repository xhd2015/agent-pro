---
label: grok
explanation: Requires real grok CLI on PATH; for design verification and debugging.
---

## Expected

- Exit code 0.
- Stderr does not report `banner not detected` or grok TUI banner timeout.
- Stderr contains prefixed session id `grok-tty: session-`.

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
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "banner not detected") {
		t.Fatalf("real grok banner not detected, stderr:\n%s", resp.Stderr)
	}
	if _, ok := parseGrokTTYSessionID(resp.Stderr); !ok {
		t.Fatalf("expected grok-tty session id on stderr:\n%s", resp.Stderr)
	}
}
```