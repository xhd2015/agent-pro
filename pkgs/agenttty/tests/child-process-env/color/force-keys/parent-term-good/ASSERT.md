## Expected

- Unset contains `NO_COLOR`.
- Force color keys present.
- TERM is the good parent value `xterm` (must not be rewritten as if dumb).
  If implementer omits TERM when already good, that is acceptable only if Set
  does not force `TERM=xterm-256color`; preferred contract matches old behavior:
  `TERM=xterm` present.

## Errors

- Rewriting good TERM to xterm-256color; missing force keys.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertBuildOK(t, resp, err)
	assertColorForceKeys(t, resp.Set, resp.Unset)
	// Must not take the dumb/empty rewrite path.
	if v, ok := setGet(resp.Set, "TERM"); ok && v == "xterm-256color" {
		t.Fatalf("good parentTERM=xterm must not be rewritten to xterm-256color; Set=%#v", resp.Set)
	}
	// Preferred: pass-through TERM=xterm (matches ApplyChildProcessEnv color branch).
	assertSetExact(t, resp.Set, "TERM", "xterm")
}
```
