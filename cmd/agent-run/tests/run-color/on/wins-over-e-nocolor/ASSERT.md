---
label: e2e
---

## Expected

- Exit code 0.
- Env probe: `NO_COLOR` still **unset** despite `-e NO_COLOR=1` (color policy last).
- Force color trio present.

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
	assertColorForceOn(t, probe)
}
```
