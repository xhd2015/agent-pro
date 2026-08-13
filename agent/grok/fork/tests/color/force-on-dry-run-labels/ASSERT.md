## Expected

- Exit 0; no launch.
- Stdout has ANSI (gray labels, SGR 90).
- Plan still contains ancestor pid 4242 and the fixture grok id.

## Side Effects

- None (dry-run).

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	assertHasANSI(t, resp.Stdout, "dry-run stdout")
	if !strings.Contains(resp.Stdout, "\x1b[90m") {
		t.Fatalf("expected gray SGR 90 on dry-run labels:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "4242") {
		t.Fatalf("missing ancestor pid:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, fixtureSessionID) {
		t.Fatalf("missing grok id:\n%s", resp.Stdout)
	}
}
```
