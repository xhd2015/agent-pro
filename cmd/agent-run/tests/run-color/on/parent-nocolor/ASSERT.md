---
label: e2e
---

## Expected

- Exit code 0.
- Env probe: `NO_COLOR` absent; `FORCE_COLOR=1`, `CLICOLOR=1`, `CLICOLOR_FORCE=1`.
- `meta.json` has no color-related field (`color` / `force_color` / …).

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

	meta := readMetaJSON(t, req.Home, req.SessionID)
	assertMetaNoColorField(t, meta)
}
```
