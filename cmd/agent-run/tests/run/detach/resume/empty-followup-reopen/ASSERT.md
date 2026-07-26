---
label: e2e
---

## Expected

- Exit code 0.
- Must not fail with `prompt is required`.
- Stdout prints both `session-id:` and `terminal-id:`.
- No attach / event stream noise.

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
	if strings.Contains(combined, "prompt is required") ||
		strings.Contains(combined, "prompt required") {
		t.Fatalf("resume --detach empty followup must not require a prompt:\n%s", combined)
	}
	assertSuccess(t, resp)
	assertDetachIDsOnStdout(t, resp)
	if noise := forbiddenDetachNoise(resp.Stdout + "\n" + resp.Stderr); len(noise) > 0 {
		t.Fatalf("resume --detach must not stream noise %v\nstdout:\n%s\nstderr:\n%s",
			noise, resp.Stdout, resp.Stderr)
	}
}
```
