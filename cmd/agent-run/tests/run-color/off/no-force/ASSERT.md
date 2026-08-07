---
label: e2e
---

## Expected

- Exit code 0.
- `NO_COLOR` still present (not cleared without `--color`).
- `FORCE_COLOR` is **not** `1` (baseline `0` preserved; feature does not force on).
- `CLICOLOR` / `CLICOLOR_FORCE` not forced to `1`.

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
	assertProbeHasKV(t, probe, "NO_COLOR", "1")
	if v, ok := probeKV(probe, "FORCE_COLOR"); ok && v == "1" {
		t.Fatalf("without --color, must not force FORCE_COLOR=1; probe:\n%s", probe)
	}
	if v, ok := probeKV(probe, "CLICOLOR"); ok && v == "1" {
		t.Fatalf("without --color, must not force CLICOLOR=1; probe:\n%s", probe)
	}
	if v, ok := probeKV(probe, "CLICOLOR_FORCE"); ok && v == "1" {
		t.Fatalf("without --color, must not force CLICOLOR_FORCE=1; probe:\n%s", probe)
	}
}
```
