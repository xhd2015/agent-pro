## Expected

- Error mentions `--help` (and preferably the unknown flag).
- Exit 1; no launch.

## Side Effects

- None.

## Errors

- mentions `--help`

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "--help")
	if !strings.Contains(resp.Err.Error(), "not-a-real-flag") &&
		!strings.Contains(resp.Stderr, "not-a-real-flag") {
		// unknown-flag name is preferred but --help is the locked contract
		t.Logf("error did not echo unknown flag name (ok): %v", resp.Err)
	}
	assertNoOpen(t, resp)
}
```
