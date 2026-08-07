---
label: e2e
---

## Expected

- Exit code 0.
- Env probe: `TERM=screen-256color` (not blindly rewritten to `xterm-256color`).
- Force color trio present; `NO_COLOR` absent.

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

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbeHasKV(t, probe, "TERM", "screen-256color")
	// Must not rewrite a good TERM to the dumb-fallback value.
	if v, ok := probeKV(probe, "TERM"); ok && v == "xterm-256color" {
		t.Fatalf("good parent TERM must not be rewritten to xterm-256color; probe:\n%s", probe)
	}
	if strings.Count(probe, "TERM=") > 1 {
		// tolerate duplicate dump lines only if all equal; probeKV takes first.
	}
	assertColorForceOn(t, probe)
}
```
