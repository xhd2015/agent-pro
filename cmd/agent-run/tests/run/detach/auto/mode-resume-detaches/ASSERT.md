---
label: e2e
---

## Expected

- Exit code 0.
- Both ids on stdout.
- Must not require a prompt.
- Not live send (`msg_N`); no stream noise.

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
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, combined, "--detach")
	if strings.Contains(combined, "prompt is required") {
		t.Fatalf("auto resume --detach empty followup must not require prompt:\n%s", combined)
	}
	assertSuccess(t, resp)
	assertDetachIDsOnStdout(t, resp)
	first := strings.TrimSpace(strings.Split(resp.Stdout, "\n")[0])
	if strings.HasPrefix(first, "msg_") {
		t.Fatalf("MODE=resume detach must not take send path; stdout:\n%s", resp.Stdout)
	}
	if noise := forbiddenDetachNoise(resp.Stdout + "\n" + resp.Stderr); len(noise) > 0 {
		t.Fatalf("auto MODE=resume detach stream noise %v", noise)
	}
}
```
