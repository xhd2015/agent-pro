---
label: e2e
---

## Expected

- Exit code 0.
- `StreamProbeSeen` is true for `PRIOR_REJECT_NEW` (new session), not `PRIOR_REJECT_OLD`.
- Stdout must **not** contain `PRIOR_REJECT_OLD`.

## Exit Code

0

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
	assertSuccess(t, resp)
	if !resp.StreamProbeSeen {
		t.Fatalf("expected new session marker %q before timeout; stdout:\n%s\nstderr:\n%s",
			priorMarkerNew, resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, priorMarkerOld) {
		t.Fatalf("matched prior session with same prompt; stdout:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, priorMarkerNew) {
		t.Fatalf("expected %q in stdout:\n%s", priorMarkerNew, resp.Stdout)
	}
}
```
