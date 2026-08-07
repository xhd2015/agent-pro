---
label: e2e
---

## Expected

- Exit code 0.
- Env probe: `TERM=xterm-256color`.
- Force color trio present; `NO_COLOR` absent.

## Exit Code

0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbeHasKV(t, probe, "TERM", "xterm-256color")
	assertColorForceOn(t, probe)
}
```
