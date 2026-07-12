## Expected

- Exit code 0.
- Env probe has `A=2` (last-win), not an effective `A=1` as the sole value.
- Meta `env` preserves ordered entries including both forms or at least ends with
  effective last-win semantics on the child (`A=2`).

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

	probe := readEnvProbe(t, req.EnvProbePath)
	assertProbeHasKV(t, probe, "A", "2")
	// Ensure A is not stuck at 1: probe should not show only A=1 without A=2.
	if strings.Contains(probe, "A=1") && !strings.Contains(probe, "A=2") {
		t.Fatalf("expected last-win A=2; probe:\n%s", probe)
	}

	meta := readMetaJSON(t, req.Home, req.SessionID)
	envList := stringSliceField(meta, "env")
	if len(envList) == 0 {
		t.Fatalf("meta.env empty; meta=%v", meta)
	}
	// Ordered append of both flags is acceptable; child effective must be A=2.
	joined := strings.Join(envList, "\n")
	if !strings.Contains(joined, "A=2") && !strings.Contains(joined, "A=1") {
		t.Fatalf("meta.env missing A entries: %v", envList)
	}
}
```
